package bot

import (
	"Deemix-Discord-Bot/deezer"
	"container/list"
	"github.com/jonas747/dca"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ServersState is a list of all servers which are currently playing music
type ServersState struct {
	servers map[string]*ServerState
	mu      sync.RWMutex
}

// ServerState contains the info about one server which is playing a music
type ServerState struct {
	// If you send anything in stopChan, the music which is currently playing will be stopped
	stopChan chan struct{}
	// If you send anything in skipChan, the music which is currently playing will be skipped
	skipChan chan struct{}
	// Queue is a queue of tracks which are playing.
	// Objects in this queue are the type of deezer.Track
	queue *list.List
	// The voice session
	session *dca.StreamingSession
	// When was the music player paused
	pausedTime time.Time
	// The channelID which the bot has joined
	channelID string
	// Mutex to lock the server state
	mu sync.RWMutex
}

// cleanupIdleServers makes the bot leave the servers which have been idle for too long
func (s *ServersState) cleanupIdleServers() {
	for {
		time.Sleep(time.Minute * 5)
		now := time.Now()
		s.mu.RLock()
		for _, server := range s.servers {
			server.mu.RLock()
			// Get all the servers which are paused for more than a minute
			if !server.pausedTime.IsZero() && now.Sub(server.pausedTime) > time.Minute*5 {
				server.stopChan <- struct{}{}
			}
			server.mu.RUnlock()
		}
		s.mu.RUnlock()
	}
}

// Stop stops the music which is currently playing on a server
func (s *ServersState) Stop(guildID string) (stopped bool) {
	s.mu.RLock()
	if server, exists := s.servers[guildID]; exists {
		stopped = true
		server.stopChan <- struct{}{}
	}
	s.mu.RUnlock()
	return
}

// Skip skips the music which is currently playing on a server
func (s *ServersState) Skip(guildID string) (skipped bool) {
	s.mu.RLock()
	if server, exists := s.servers[guildID]; exists {
		skipped = true
		server.skipChan <- struct{}{}
	}
	s.mu.RUnlock()
	return
}

// Play registers a server as playing and returns the ServerState which corresponds to this server
// If the server exists, it will add the "track" to it's queue
// If the server does not exist, it will initialize the server object
func (s *ServersState) Play(guildID, voiceChannelID string, track deezer.Track) (state *ServerState, newServer bool) {
	s.mu.Lock()
	state, exists := s.servers[guildID]
	if !exists {
		state = &ServerState{
			// The buffer of stopChan ensures that the channel will always receive the requests and never holds them
			stopChan: make(chan struct{}, math.MaxInt32),
			skipChan: make(chan struct{}, math.MaxInt32),
			// We also create a linked list to add the track
			queue: list.New(),
			// Add the channel ID
			channelID: voiceChannelID,
		}
		s.servers[guildID] = state
	}
	// State is always initialized here
	state.queue.PushBack(track)
	s.mu.Unlock()
	return state, !exists
}

// DeleteServer simply removes the server from list
func (s *ServersState) DeleteServer(guildID string) {
	s.mu.Lock()
	delete(s.servers, guildID)
	s.mu.Unlock()
}

// GetQueueText returns the list of queued musics in a server
func (s *ServersState) GetQueueText(guildID string) string {
	// Get the server
	s.mu.RLock()
	server, exists := s.servers[guildID]
	s.mu.RUnlock()
	// If the server is playing something...
	var queue strings.Builder
	if exists {
		// Loop for each song
		server.mu.RLock()
		i := 0
		for head := server.queue.Front(); head != nil; head = head.Next() {
			i++
			queue.WriteString(strconv.Itoa(i))
			queue.WriteString(". ")
			queue.WriteString(head.Value.(deezer.Track).String())
			queue.WriteByte('\n')
		}
		server.mu.RUnlock()
	}
	// Check empty queue
	if queue.Len() == 0 {
		return "Empty queue!"
	}
	return queue.String()
}

// RemoveQueuedTrack removes a queued track from a server
// The index starts at 1
func (s *ServersState) RemoveQueuedTrack(guildID string, index int) (ok bool) {
	// Get the server
	s.mu.RLock()
	server, exists := s.servers[guildID]
	s.mu.RUnlock()
	// Remove the track
	if !exists {
		return false
	}
	// Special case: First music is the playing one. Just skip it
	if index == 1 {
		server.skipChan <- struct{}{}
		return true
	}
	// Loop and find remove index
	index--
	server.mu.Lock()
	// If the index is invalid don't do anything
	if index <= 0 || index >= server.queue.Len() {
		server.mu.Unlock()
		return false
	}
	// Loop until we reach the specified track and remove it
	for head := server.queue.Front(); head != nil; head = head.Next() {
		if index == 0 {
			server.queue.Remove(head)
			break
		}
		index--
	}
	server.mu.Unlock()
	return true
}

// Pop removes the last track of queued server
func (s *ServersState) Pop(guildID string) (ok bool) {
	// Get the server
	s.mu.RLock()
	server, exists := s.servers[guildID]
	s.mu.RUnlock()
	// Remove the track
	if !exists {
		return false
	}
	// Check the queue
	server.mu.Lock()
	if server.queue.Len() == 1 {
		// I could also send this in skip channel
		server.stopChan <- struct{}{}
	} else {
		server.queue.Remove(server.queue.Back())
	}
	server.mu.Unlock()
	return true
}

// Pause pauses or unpauses the playing music
func (s *ServersState) Pause(guildID string, paused bool) {
	// Get the server
	s.mu.RLock()
	server, exists := s.servers[guildID]
	s.mu.RUnlock()
	// Set the status
	if !exists {
		return
	}
	server.mu.Lock()
	if server.session != nil {
		server.session.SetPaused(paused)
		if paused {
			server.pausedTime = time.Now()
		} else {
			server.pausedTime = time.Time{}
		}
	}
	server.mu.Unlock()
}

// DequeTrack removes the currently playing track from a server (first track in list)
// It also returns the number of remaining tracks
func (s *ServerState) DequeTrack() (remainingTracks int) {
	s.mu.Lock()
	s.queue.Remove(s.queue.Front())
	remainingTracks = s.queue.Len()
	s.mu.Unlock()
	return
}

// GetPlayingTrack gets the currently playing track from a list
// It also says if the server is playing something or not
func (s *ServersState) GetPlayingTrack(guildID string) (track deezer.Track, exists bool) {
	s.mu.RLock()
	server, ok := s.servers[guildID]
	s.mu.RUnlock()
	if ok {
		return server.GetPlayingTrack()
	}
	return
}

// HasPlayingMusicInChannel checks if the bot is playing a music in specified voice channel and server
func (s *ServersState) HasPlayingMusicInChannel(guildID, channelID string) bool {
	s.mu.RLock()
	server, ok := s.servers[guildID]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	return server.channelID == channelID
}

// GetPlayingTrack gets the currently playing track from a list
// It also says if the server is playing something or not
func (s *ServerState) GetPlayingTrack() (track deezer.Track, exists bool) {
	s.mu.RLock()
	if f := s.queue.Front(); f != nil {
		track = f.Value.(deezer.Track)
		exists = true
	}
	s.mu.RUnlock()
	return
}

// SetVoiceSession sets the voice session of a server
func (s *ServerState) SetVoiceSession(session *dca.StreamingSession) {
	s.mu.Lock()
	s.session = session
	s.mu.Unlock()
}

// RemoveVoiceSession removes the voice session of a server
func (s *ServerState) RemoveVoiceSession() {
	s.mu.Lock()
	s.session = nil
	s.mu.Unlock()
}

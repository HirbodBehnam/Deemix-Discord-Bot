package bot

import (
	"math"
	"sync"
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
}

// Stop stops the music which is currently playing on a server
func (s *ServersState) Stop(guildID string) {
	s.mu.RLock()
	if server, exists := s.servers[guildID]; exists {
		server.stopChan <- struct{}{}
	}
	s.mu.RUnlock()
}

// Play registers a server as playing and returns the ServerState which corresponds to this server
// If the returned value is null, it means that the server is currently playing a music
func (s *ServersState) Play(guildID string) *ServerState {
	s.mu.Lock()
	var state *ServerState
	_, exists := s.servers[guildID]
	if !exists {
		state = &ServerState{
			// The buffer of stopChan ensures that the channel will always receive the requests and never holds them
			stopChan: make(chan struct{}, math.MaxInt), // A VERY BIG FUCKING BUFFER
		}
		s.servers[guildID] = state
	}
	s.mu.Unlock()
	return state
}

// DeleteServer simply removes the server from list
func (s *ServersState) DeleteServer(guildID string) {
	s.mu.Lock()
	delete(s.servers, guildID)
	s.mu.Unlock()
}

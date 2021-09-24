package bot

import (
	"Deemix-Discord-Bot/deezer"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"io"
	"log"
)

// playMusic might initialize a voice connection to start playing the music,
// or it might just push the track to queue
func playMusic(s *discordgo.Session, guildID, voiceChannelID, textChannelID, text string) {
	// Get the track info or search and get the track info
	track, err := deezer.KeywordToLink(text)
	if err != nil {
		_, _ = s.ChannelMessageSend(textChannelID, "Cannot play this music: "+err.Error())
		return
	}
	// Add the track to server queue
	serverState, newServer := serverList.Play(guildID, track)
	if !newServer { // If this server is playing a music just send the info about queue and do nothing
		_, _ = s.ChannelMessageSend(textChannelID, "Queued "+track.String())
		return
	}
	// So if we reach this line, we can understand that this goroutine will be used to stream
	// the music to Discord
	// So when this goroutine is killed, we have to remove the server from this list
	defer serverList.DeleteServer(guildID)
	// Join the channel
	vc, err := s.ChannelVoiceJoin(guildID, voiceChannelID, false, true)
	if err != nil {
		log.Println("cannot join the voice channel:", err)
		return
	}
	defer func(vc *discordgo.VoiceConnection) {
		_ = vc.Disconnect()
	}(vc)
	// Loop until the queue is done
	for {
		track, exists := serverState.GetPlayingTrack()
		if !exists {
			return
		}
		shouldStop := playMusicInVoice(s, vc, serverState, textChannelID, track)
		if shouldStop || serverState.DequeTrack() == 0 {
			return
		}
	}
}

// playMusicInVoice plays a music in a voice channel
func playMusicInVoice(s *discordgo.Session, vc *discordgo.VoiceConnection, serverState *ServerState, textChannelID string, track deezer.Track) (shouldStop bool) {
	_, _ = s.ChannelMessageSend(textChannelID, "Now playing "+track.String())
	// Download the music
	tempDir, err := deezer.Download(track.Link, serverState.stopChan)
	if err != nil {
		log.Println("cannot download the music from deezer:", err)
		return true
	}
	defer tempDir.Delete()
	// Check downloaded file
	musics := tempDir.GetMusics()
	if len(musics) == 0 {
		_, _ = s.ChannelMessageSend(textChannelID, "Music not found")
		return false
	}
	// Start streaming
	_ = vc.Speaking(true)
	defer func(vc *discordgo.VoiceConnection) {
		_ = vc.Speaking(false)
	}(vc)
	// Play it
	done := make(chan error, 1)
	encodeSession, err := dca.EncodeFile(musics[0], dca.StdEncodeOptions)
	if err != nil {
		log.Println("cannot encode:", encodeSession)
		return false
	}
	defer encodeSession.Cleanup()
	// Create a stream
	stream := dca.NewStream(encodeSession, vc, done)
	serverState.SetVoiceSession(stream)
	defer serverState.RemoveVoiceSession()
	// Wait either the stream is done, or the bot is stopped
	select {
	case err = <-done:
	case <-serverState.stopChan:
		return true
	case <-serverState.skipChan:
		return false
	}
	if err != nil && err != io.EOF {
		log.Println("there was a problem streaming the song:", err)
		return true
	}
	return false
}

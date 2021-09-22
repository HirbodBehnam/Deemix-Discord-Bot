package bot

import (
	"Deemix-Discord-Bot/config"
	"Deemix-Discord-Bot/deezer"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// RunBot runs the discord bot with config.Config configurations
func RunBot() {
	dg, err := discordgo.New("Bot " + config.Config.Token)
	if err != nil {
		log.Fatalln("Error creating Discord session: ", err)
	}
	dg.AddHandler(onReady)
	dg.AddHandler(onMessage)
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates
	err = dg.Open()
	if err != nil {
		log.Fatalln("Error opening Discord session: ", err)
	}
	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = dg.Close()
	log.Println("Clean shutdown the bot")
}

func onReady(s *discordgo.Session, _ *discordgo.Ready) {
	_ = s.UpdateGameStatus(0, config.Config.Prefix+"help")
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Check the command prefix
	if !strings.HasPrefix(m.Content, config.Config.Prefix) {
		return
	}
	var command Command
	if command.Parse(m.Content[len(config.Config.Prefix):]) != nil {
		return
	}
	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		log.Println("cannot get the channel:", err)
		return
	}
	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		log.Println("cannot get the guild:", err)
		return
	}
	// Check the command
	switch command {
	case CommandHelp:
		_, _ = s.ChannelMessageSendReply(c.ID, config.HelpMessage, m.Reference())
	case CommandRepo:
		_, _ = s.ChannelMessageSendReply(c.ID, config.Repo, m.Reference())
	case CommandStop:
		serverList.Stop(g.ID)
	case CommandSkip:
		serverList.Skip(g.ID)
	case CommandQueueView:
		_, _ = s.ChannelMessageSendReply(c.ID, serverList.GetQueueText(g.ID), m.Reference())
	case CommandPlay:
		// Find the user's voice channel
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				// Play it in another goroutine
				go playMusic(s, g.ID, vs.ChannelID, c.ID, strings.Trim(m.Content[len(config.Config.Prefix)+len(playCommand):], " "))
				return
			}
		}
		_, _ = s.ChannelMessageSendReply(c.ID, "Join a voice channel!", m.Reference())
	case CommandPlayingTrack:
		track, playing := serverList.GetPlayingTrack(g.ID)
		if !playing {
			_, _ = s.ChannelMessageSendReply(c.ID, "Nothing is playing!", m.Reference())
		} else {
			_, _ = s.ChannelMessageSendReply(c.ID, "Currently playing: "+track.String(), m.Reference())
		}
	}
}

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
	dca.NewStream(encodeSession, vc, done)
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

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
	case CommandPlay:
		// Find the user's voice channel
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				// Try to play the music
				state := serverList.Play(g.ID)
				if state == nil {
					_, _ = s.ChannelMessageSendReply(c.ID, "Bot is currently playing a music!", m.Reference())
					return
				}
				// Play it in another goroutine
				go playMusic(s, g.ID, vs.ChannelID, c.ID, strings.Trim(m.Content[len(config.Config.Prefix)+len(playCommand):], " "), state)
				return
			}
		}
		_, _ = s.ChannelMessageSendReply(c.ID, "Join a voice channel!", m.Reference())
	}
}

func playMusic(s *discordgo.Session, guildID, voiceChannelID, textChannelID, text string, serverState *ServerState) {
	// Initialize the stop channel
	defer serverList.DeleteServer(guildID)
	// Search the music if needed
	track, err := deezer.KeywordToLink(text)
	if err != nil {
		_, _ = s.ChannelMessageSend(textChannelID, "Cannot play this music: "+err.Error())
		return
	}
	// Download the music
	tempDir, err := deezer.Download(track.Link, serverState.stopChan)
	if err != nil {
		log.Println("cannot download the music from deezer:", err)
		return
	}
	defer tempDir.Delete()
	// Check downloaded file
	musics := tempDir.GetMusics()
	if len(musics) == 0 {
		_, _ = s.ChannelMessageSend(textChannelID, "Music not found")
		return
	}
	_, _ = s.ChannelMessageSend(textChannelID, "Playing "+track.String())
	// Join the channel
	vc, err := s.ChannelVoiceJoin(guildID, voiceChannelID, false, true)
	if err != nil {
		log.Println("cannot join the voice channel:", err)
		return
	}
	// Start streaming
	_ = vc.Speaking(true)
	defer func(vc *discordgo.VoiceConnection) {
		_ = vc.Speaking(false)
		_ = vc.Disconnect()
	}(vc)
	// Play it
	done := make(chan error, 1)
	encodeSession, err := dca.EncodeFile(musics[0], dca.StdEncodeOptions)
	if err != nil {
		log.Println("cannot encode:", encodeSession)
		return
	}
	defer encodeSession.Cleanup()
	dca.NewStream(encodeSession, vc, done)
	// Wait either the stream is done, or the bot is stopped
	select {
	case err = <-done:
	case <-serverState.stopChan:
	}
	if err != nil && err != io.EOF {
		log.Println("there was a problem streaming the song:", err)
		return
	}
}

package bot

import (
	"Deemix-Discord-Bot/config"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// RunBot runs the discord bot with config.Config configurations
func RunBot() {
	// Start the server cleanup
	go serverList.cleanupIdleServers()
	// Start the discord bot
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
	case CommandPause:
		serverList.Pause(g.ID, true)
	case CommandResume:
		serverList.Pause(g.ID, false)
	case CommandQueueRemove:
		index, err := strconv.Atoi(strings.Trim(m.Content[len(config.Config.Prefix)+len(removeFromQueueCommand):], " "))
		if err != nil {
			_, _ = s.ChannelMessageSendReply(c.ID, "Please pass the index of the music as well.\nFor example `"+config.Config.Prefix+"remove 2`", m.Reference())
			return
		}
		ok := serverList.RemoveQueuedTrack(g.ID, index)
		if ok {
			_, _ = s.ChannelMessageSendReply(c.ID, "Removed", m.Reference())
		} else {
			_, _ = s.ChannelMessageSendReply(c.ID, "Invalid index", m.Reference())
		}
	case CommandQueuePop:
		ok := serverList.Pop(g.ID)
		if ok {
			_, _ = s.ChannelMessageSendReply(c.ID, "Popped!", m.Reference())
		} else {
			_, _ = s.ChannelMessageSendReply(c.ID, "Nothing to remove", m.Reference())
		}
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

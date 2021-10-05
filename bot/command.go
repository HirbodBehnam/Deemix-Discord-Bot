package bot

import (
	"errors"
	"strings"
)

var InvalidCommandError = errors.New("invalid command")

const playCommand = "play"
const removeFromQueueCommand = "remove"
const searchCommand = "search"

// Command is a command which is given to the bot
type Command byte

const (
	CommandHelp Command = iota
	CommandPlay
	CommandStop
	CommandRepo
	CommandPlayingTrack
	CommandSkip
	CommandQueueView
	CommandQueueRemove
	CommandQueuePop
	CommandPause
	CommandResume
	CommandSearch
)

// Parse parses the command given to bot as Command
// Please note that input must not contain the prefix
func (c *Command) Parse(input string) error {
	if strings.HasPrefix(input, playCommand+" ") {
		*c = CommandPlay
		return nil
	}
	if strings.HasPrefix(input, removeFromQueueCommand+" ") {
		*c = CommandQueueRemove
		return nil
	}
	if strings.HasPrefix(input, searchCommand+" ") {
		*c = CommandSearch
		return nil
	}
	switch input {
	case "stop":
		*c = CommandStop
	case "help":
		*c = CommandHelp
	case "repo":
		*c = CommandRepo
	case "playing":
		*c = CommandPlayingTrack
	case "skip":
		*c = CommandSkip
	case "queue":
		*c = CommandQueueView
	case "pop":
		*c = CommandQueuePop
	case "pause":
		*c = CommandPause
	case "resume":
		*c = CommandResume
	default:
		return InvalidCommandError
	}
	return nil
}

package bot

import (
	"errors"
	"strings"
)

var InvalidCommandError = errors.New("invalid command")

const playCommand = "play"

// Command is a command which is given to the bot
type Command byte

const (
	CommandHelp Command = iota
	CommandPlay
	CommandStop
	CommandRepo
)

// Parse parses the command given to bot as Command
// Please note that input must not contain the prefix
func (c *Command) Parse(input string) error {
	if strings.HasPrefix(input, playCommand) {
		*c = CommandPlay
		return nil
	}
	switch input {
	case "stop":
		*c = CommandStop
	case "help":
		*c = CommandHelp
	case "repo":
		*c = CommandRepo
	default:
		return InvalidCommandError
	}
	return nil
}

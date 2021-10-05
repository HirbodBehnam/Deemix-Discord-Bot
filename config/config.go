package config

import (
	"encoding/json"
	"log"
	"os"
)

const Version = "0.3.0"
const Repo = "https://github.com/HirbodBehnam/Deemix-Discord-Bot"

var HelpMessage string

// Config is the config of application
var Config struct {
	// Token of discord
	Token string `json:"token"`
	// Prefix of bot commands
	Prefix string `json:"prefix"`
}

// LoadConfig reads the config file from disk
func LoadConfig(location string) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		log.Fatalf("Cannot read config file: %s\n", err)
	}
	err = json.Unmarshal(bytes, &Config)
	if err != nil {
		log.Fatalf("Cannot parse config file: %s\n", err)
	}
	// Fix prefix
	if Config.Prefix == "" {
		Config.Prefix = "?"
	}
	// Fix help message
	HelpMessage = "Welcome to my private music bot v" + Version + ". Here are the list of commands which you can use:\n" +
		Config.Prefix + "help : Show this message again\n" +
		Config.Prefix + "play <link>/<keyword> : Play a song from deezer or search and play a song from deezer.\n" +
		Config.Prefix + "skip : Skip the current song\n" +
		Config.Prefix + "queue : Show the queue\n" +
		Config.Prefix + "remove <index> : Removes the nth track from queue\n" +
		Config.Prefix + "pop : Removes the last track from queue\n" +
		Config.Prefix + "playing : Show playing song name\n" +
		Config.Prefix + "search <keyword> : Search a track in deezer\n" +
		Config.Prefix + "stop : Stops the playing music\n" +
		Config.Prefix + "repo : Show the source code"
}

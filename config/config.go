package config

import (
	"encoding/json"
	"log"
	"os"
)

const Version = "0.1.0"
const Repo = "https://www.github.com/HirbodBehnam"

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
		Config.Prefix + "play <link>/<keyword> : Play a song from deezer or search and play a song from deezer; This command does nothing if a music is already playing\n" +
		Config.Prefix + "stop : Stops the playing music\n" +
		Config.Prefix + "repo : Show the source code"
}

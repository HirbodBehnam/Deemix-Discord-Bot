package main

import (
	"Deemix-Discord-Bot/bot"
	"Deemix-Discord-Bot/config"
	"os"
)

func main() {
	// Load config
	if len(os.Args) > 1 {
		config.LoadConfig(os.Args[1])
	} else {
		config.LoadConfig("config.json")
	}
	// Run the bot
	bot.RunBot()
}

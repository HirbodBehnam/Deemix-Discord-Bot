# Deemix Discord Bot
A simple discord music bot which uses deemix to play music.

## Features
* Easy to setup
* Easy on resources
* Use deezer instead of youtube. This is a disadvantage as well.

## Setup
At first create a discord bot from [here](https://discord.com/developers/applications).
Then download this repo and compile it (you can also use releases).
Copy the `config.json` file from config folder to root of your program and edit it. Options of this file are shown in next segment.
Then install deemix and run it at least once to setup the arl cookie in it.
Also install ffmpeg.
At last, run the program to start your bot.

### Config file
Config file, right now, only has two fields which one is optional:
`token`: Your discord bot token.
`prefix`(optional): The prefix of bot commands.

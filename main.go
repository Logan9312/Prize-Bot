package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"gitlab.com/logan9312/discord-auction-bot/commands"
	"gitlab.com/logan9312/discord-auction-bot/connect"
	"gitlab.com/logan9312/discord-auction-bot/database"
	"gitlab.com/logan9312/discord-auction-bot/routers"
	"github.com/bwmarrin/discordgo"
)

// Environment struct
type Environment struct {
	Environment  string `env:"ENVIRONMENT,required"`
	DiscordToken string `env:"DISCORD_TOKEN,required"`
	Migrate      bool   `env:"MIGRATE"`
	Host 		 string `env:"DB_HOST"`
	Password	 string `env:"DB_PASSWORD"`
	Grungerson 	string 	`env:"GRUNGERSON"`
}

var prodCommands = []*discordgo.ApplicationCommand{
	&commands.HelpCommand,
	&commands.SpawnExactDinoCommand,
}

var localCommands = []*discordgo.ApplicationCommand{
	&commands.HelpCommand,
	&commands.ProfileCommand,
	&commands.AuctionCommand,
}

func main() {

	environment := Environment{}

	if err := env.Parse(&environment); err != nil {
		log.Fatal("FAILED TO LOAD ENVIRONMENT VARIABLES")
	}

	//Connects main bot
	connect.BotConnect(environment.DiscordToken, environment.Environment, "Main Bot", prodCommands, localCommands)

	//Connects Sir Grungerson
	connect.BotConnect(environment.Grungerson, environment.Environment, "Sir Grungerson", prodCommands, localCommands)

	//Connects database
	database.DatabaseConnect(environment.Host, environment.Password)

	fmt.Println("Bot is running! To stop, use: docker kill $(docker ps -q)")

	routers.BotStatus()
}

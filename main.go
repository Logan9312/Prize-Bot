package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"gitlab.com/logan9312/discord-auction-bot/connect"
	"gitlab.com/logan9312/discord-auction-bot/database"
	"gitlab.com/logan9312/discord-auction-bot/routers"
)

// Environment struct
type Environment struct {
	Environment  string `env:"ENVIRONMENT,required"`
	DiscordToken string `env:"DISCORD_TOKEN,required"`
	Migrate      bool   `env:"MIGRATE"`
	Host         string `env:"DB_HOST"`
	Password     string `env:"DB_PASSWORD"`
}

func main() {

	environment := Environment{}

	if err := env.Parse(&environment); err != nil {
		fmt.Println(err)
		log.Fatal("FAILED TO LOAD ENVIRONMENT VARIABLES")
	}

	//Connects main bot
	go connect.BotConnect(environment.DiscordToken, environment.Environment, "Main Bot")

	//Connects database
	go database.DatabaseConnect(environment.Password, environment.Host, environment.Environment)

	fmt.Println("Bot is running! To stop, use: docker kill $(docker ps -q)")

	routers.BotStatus()
}

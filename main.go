package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"github.com/stripe/stripe-go/v72"
	"gitlab.com/logan9312/discord-auction-bot/commands"
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
	StripeToken  string `env:"STRIPE_TOKEN"`
}

func main() {

	environment := Environment{}

	if err := env.Parse(&environment); err != nil {
		fmt.Println(err)
		log.Fatal("FAILED TO LOAD ENVIRONMENT VARIABLES")
	}

	stripe.Key = environment.StripeToken

	//Connects database
	database.DatabaseConnect(environment.Password, environment.Host, environment.Environment)

	//Connects main bot
	mainSession, err := connect.BotConnect(environment.DiscordToken, environment.Environment)
	if err != nil {
		fmt.Println(err)
	}

	devData := database.DevSetup{
		BotID: mainSession.State.User.ID,
	}
	result := database.DB.First(&devData)
	if result.Error != nil {
		fmt.Println(result.Error)
	}

	WhiteLabels := []map[string]any{}
	result = database.DB.Model([]database.WhiteLabels{}).Find(&WhiteLabels)
	if result.Error != nil {
		fmt.Println("Error fetching whitelabels:", result.Error)
	}

	connect.Timers(mainSession)

	for _, v := range WhiteLabels {
		s, err := connect.BotConnect(v["bot_token"].(string), environment.Environment)
		if err != nil {
			fmt.Println("Error connecting bot: %w", err)
		}
		if s.State.User.ID == "995022149226082324" {
			err = s.UpdateGameStatus(0, "Bot Version "+devData.Version)
			if err != nil {
				fmt.Println("Error setting status", err)
			}
		}
		connect.Timers(s)
	}

	go commands.SetRoles(mainSession)

	fmt.Println("Bot is running!")

	//TODO Possibly put this at the start as a goroutine
	routers.BotStatus()
}

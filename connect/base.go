package connect

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/logan9312/discord-auction-bot/commands"
	"gitlab.com/logan9312/discord-auction-bot/database"
)

type slashCommands struct {
	local, prod []*discordgo.ApplicationCommand
}

func BotConnect(token, environment, botName string) {

	var c = slashCommands{
		local: []*discordgo.ApplicationCommand{
			&commands.HelpCommand,
			&commands.ProfileCommand,
			&commands.AuctionCommand,
			&commands.SpawnExactDinoCommand,
			&commands.QueueCommand,
		},
		prod: []*discordgo.ApplicationCommand{
			&commands.HelpCommand,
			&commands.AuctionCommand,
			&commands.SpawnExactDinoCommand,
		},
	}

	var status string

	fmt.Println(botName + " Starting Up...")

	s, err := discordgo.New("Bot " + token)

	commands.Session = s

	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.Open()

	if err != nil {
		fmt.Println(err)
		return
	}

	status = "Aftermath Ark"

	for _, v := range c.local {
		v.Description = "EXPERIMENTAL: " + v.Description
	}

	//Builds local commands
	if environment == "local" {
		for _, v := range s.State.Guilds {
			_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, v.ID, c.local)
			fmt.Println("Commands added to guild: " + v.Name)
			if err != nil {
				fmt.Println(err)
			}
		}
		commands.HelpBuilder(c.local)
	}

	//Builds prod commands
	if environment == "prod" {
		_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", c.prod)
		if err != nil {
			fmt.Println(err)
		}
		commands.HelpBuilder(c.prod)
	}

	s.AddHandler(CommandHandler)

	go Timers()

	err = s.UpdateGameStatus(0, status)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(botName + " Startup Complete!")
}

func Timers() {

	var Auctions []database.Auction

	database.DB.Find(&Auctions)

	for _, v := range Auctions {
		go SetTimer(v)
	}
}

func SetTimer(v database.Auction) {
	if v.EndTime.After(time.Now()) {
		commands.AuctionEnd(v.ChannelID)
	} else {
		time.Sleep(time.Until(v.EndTime))
		commands.AuctionEnd(v.ChannelID)
	}
}
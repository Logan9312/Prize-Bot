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
			&commands.QueueCommand,
			&commands.BidCommand,
		},
		prod: []*discordgo.ApplicationCommand{
			&commands.HelpCommand,
			&commands.AuctionCommand,
			&commands.BidCommand,
		},
	}

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

	go Timers(s)

	err = s.UpdateGameStatus(0, "Bot Version v0.4")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(botName + " Startup Complete!")
}

func Timers(s *discordgo.Session) {

	var Auctions []database.Auction
	var AuctionQueue []database.AuctionQueue

	database.DB.Find(&Auctions)
	for _, v := range Auctions {
		go AuctionEndTimer(v, s)
	}

	database.DB.Find(&AuctionQueue)
	for _, v := range AuctionQueue {
		go AuctionStartTimer(v, s)
	}
}

func AuctionEndTimer(v database.Auction, s *discordgo.Session) {
	if v.EndTime.Before(time.Now()) {
		commands.AuctionEnd(v.ChannelID, v.GuildID)
	} else {
		time.Sleep(time.Until(v.EndTime))
		commands.AuctionEnd(v.ChannelID, v.GuildID)
	}
}

func AuctionStartTimer(v database.AuctionQueue, s *discordgo.Session) {
	if v.StartTime.Before(time.Now()) {
		commands.AuctionCreate(s, v)
	} else {
		time.Sleep(time.Until(v.EndTime))
		commands.AuctionCreate(s, v)
	}
}

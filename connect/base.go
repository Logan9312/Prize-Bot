package connect

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/logan9312/discord-auction-bot/commands"
)

var mainID = "829527477268774953"
var grungyID = "864930428639772692"

var MainCommands = []*discordgo.ApplicationCommand{}

var GrungyCommands = []*discordgo.ApplicationCommand{
	&commands.ReviewCommand,
	&commands.ReviewEditCommand,
}

func BotConnect(token, environment, botName string) {

	var prodCommands = []*discordgo.ApplicationCommand{
		&commands.HelpCommand,
	}

	var localCommands = []*discordgo.ApplicationCommand{
		&commands.HelpCommand,
		&commands.ProfileCommand,
		&commands.AuctionCommand,
	}

	var status string

	fmt.Println(botName + " Starting Up...")

	s, err := discordgo.New("Bot " + token)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.Open()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("User ID: " + s.State.User.ID)
	switch s.State.User.ID {
	case mainID:
		status = "Aftermath Ark"
		prodCommands = append(prodCommands, MainCommands...)

	case grungyID:
		status = "suggon"
		prodCommands = append(prodCommands, GrungyCommands...)
	}

	CommandBuilder(s, environment, localCommands, prodCommands)

	s.AddHandler(commands.CommandHandler)

	err = s.UpdateGameStatus(0, status)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer fmt.Println(botName + " Startup Complete!")
}

package connect

import (
	"github.com/bwmarrin/discordgo"
	c "gitlab.com/logan9312/discord-auction-bot/commands"
)


func CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == 2 {
		switch i.ApplicationCommandData().Name {
		case "help":
			c.Help(s, i)
		case "auction":
			c.Auction(s, i, Database)
		case "inventory":
			c.Profile(s, i, Database)
		case "queue":
			c.Queue(s, i)
		case "spawn-exact-dino":
			c.SpawnExactDino(s, i)
		default:
			CommandResponse(s, i)
		}
	}
	if i.Type == 3 {
		switch i.MessageComponentData().CustomID {
		case "startbid":
			c.AuctionButton(s, i)
		case "placebid":
			c.Bid(s, i)
		case "categorymenu":
			c.CategorySelect(s, i, Database)
		default:
			CommandResponse(s, i)
		}
	}
}

func CommandResponse (s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:          []*discordgo.MessageEmbed{
				{
					Title:       "Command Selection Error",
					Description: "Command response has not been set properly, please contact Logan to fix",
				},
			},
			Flags:           64,
		},
	})
}

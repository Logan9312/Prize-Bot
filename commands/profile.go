package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	h "gitlab.com/logan9312/discord-auction-bot/helpers"
)

var ProfileCommand = discordgo.ApplicationCommand{
	Name:        "profile",
	Description: "Displays a user's profile.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "Chose who's profile to display",
			Required:    true,
		},
	},
}

func Profile(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	userID := i.ApplicationCommandData().Options[0].UserValue(s).ID
	username := i.ApplicationCommandData().Options[0].UserValue(s).Username
	Discriminator := i.ApplicationCommandData().Options[0].UserValue(s).Discriminator

	err := h.SuccessResponse(s, i, h.PresetResponse{
		Title:       "**__" + username + "__**" + "#" + Discriminator,
		Description: "Inventory For: <@" + userID + ">",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: i.ApplicationCommandData().Options[0].UserValue(s).AvatarURL(""),
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

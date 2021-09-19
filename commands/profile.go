package commands

import (
	"github.com/bwmarrin/discordgo"
)

var ProfileCommand = discordgo.ApplicationCommand{
	Name:        "profile",
	Description: "Displays a user's profile.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type: discordgo.ApplicationCommandOptionUser,
			Name: "user",
			Description: "Chose who's profile to display",
			Required: true,
		},
	},
}

func Profile(s *discordgo.Session, i *discordgo.InteractionCreate) {

	userID := i.ApplicationCommandData().Options[0].UserValue(s).ID
	username := i.ApplicationCommandData().Options[0].UserValue(s).Username
	Discriminator := i.ApplicationCommandData().Options[0].UserValue(s).Discriminator

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         "",
			Components:      []discordgo.MessageComponent{},
			Embeds:          []*discordgo.MessageEmbed{
				{
					Title: "**__" + username + "__**" + "#" + Discriminator,
					Description: "Inventory For: <@" + userID + ">",
					Color: 0x8073ff,
					Fields:    []*discordgo.MessageEmbedField{},
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL:     i.ApplicationCommandData().Options[0].UserValue(s).AvatarURL(""),
					},
				},
			},
		},
	})
}
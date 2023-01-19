package commands

import (
	"github.com/bwmarrin/discordgo"
	"gitlab.com/logan9312/discord-auction-bot/helpers"
)

var SettingsCommand = discordgo.ApplicationCommand{
	Name:                     "settings",
	Description:              "Change any settings for your bot",
	DefaultMemberPermissions: helpers.Ptr(int64(discordgo.PermissionManageServer)),
	DMPermission:             new(bool),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "claiming",
			Description: "Edit your claiming settings",
			Options: []*discordgo.ApplicationCommandOption{
				CommandOptionCategory,
				CommandOptionLogChannel,
				CommandOptionInstructions,
				CommandOptionDisableClaiming,
				//TODO Make channel prefix work for claiming
				//CommandOptionChannelPrefix,
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "auctions",
			Description: "Edit your auction settings",
			Options: []*discordgo.ApplicationCommandOption{
				CommandOptionLogChannel,
				CommandOptionCategory,
				CommandOptionAlertRole,
				CommandOptionIntegerOnly,
				CommandOptionUseCurrency,
				CommandOptionSnipeExtension,
				CommandOptionSnipeRange,
				CommandOptionChannelLock,
				CommandOptionChannelPrefix,
			},
		},
		{
			Type:         discordgo.ApplicationCommandOptionSubCommand,
			Name:         "giveaways",
			Description:  "Edit your giveaway settings",
			ChannelTypes: []discordgo.ChannelType{},
			Required:     false,
			Options: []*discordgo.ApplicationCommandOption{
				CommandOptionAlertRole,
				CommandOptionLogChannel,
				CommandOptionAutoEnter,
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "currency",
			Description: "Edit your currency settings",
			Options: []*discordgo.ApplicationCommandOption{
				//TODO Rework for multiple currencies
				CommandOptionCurrency,
				CommandOptionCurrencySide,
			},
		},
		/*{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "shop",
			Description: "Edit your shop settings",
			Options: []*discordgo.ApplicationCommandOption{},
		},*/
	},
}

func FetchSettings (target any) {
	
}
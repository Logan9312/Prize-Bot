package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/logan9312/discord-auction-bot/database"
)

var Session *discordgo.Session

var AuctionCommand = discordgo.ApplicationCommand{
	Name:        "auction",
	Description: "Put an item up for auction!",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "help",
			Description: "auction info",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setup",
			Description: "Setup auctions on your server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Sets the category to create auctions in. Name must be an exact match",
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "currency",
					Description: "Sets the auction currency",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "Create an Auction",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "item",
					Description: "The item you wish to auction off",
					Required:    true,
				},
				{
					Type:        10,
					Name:        "startingbid",
					Description: "The starting price to bid on",
					Required:    true,
				},
				{
					Type:        10,
					Name:        "duration",
					Description: "Time (in hours) that the auction will run for",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "bid",
			Description: "Bid on an Auction",
			Required:    false,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        10,
					Name:        "amount",
					Description: "Place your bid here",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "secret_bidder",
					Description: "Turn this on to protect your identity for the next bid.",
					Required:    false,
				},
			},
		},
	},
}

func Auction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Options[0].Name {
	case "setup":
		AuctionSetup(s, i)
	case "create":
		AuctionCreate(s, i)
	case "bid":
		AuctionBid(s, i)
	}
}

func AuctionSetup(s *discordgo.Session, i *discordgo.InteractionCreate) {

	//Set Variables
	currency := "$"
	category := ""

	//Variable assignment
	options := ParseSubCommand(i)
	category = options["category"].(string)
	currency = options["currency"].(string)
	catIDs := make([]string, 0)
	catMatch := false
	status := "FAILED"
	content := "There is no category in your discord server that matches the name: " + category
	catMenu := make([]discordgo.SelectMenuOption, 0)
	componentValue := []discordgo.MessageComponent{}

	if category != "" {
		channels, err := s.GuildChannels(i.GuildID)

		if err != nil {
			fmt.Println(err)
		}

		for n, v := range channels {
			if v.Type == 4 {
				if strings.EqualFold(category, v.Name) {
					catMatch = true
					catIDs = append(catIDs, v.ID)
					menuOption := discordgo.SelectMenuOption{
						Label:       fmt.Sprintf("%s (%s)", v.Name, v.ID),
						Value:       v.ID,
						Description: "Discord Channel number (includes all channels/categories): " + fmt.Sprint(n+1),
						Emoji:       discordgo.ComponentEmoji{},
						Default:     false,
					}
					catMenu = append(catMenu, menuOption)
				}
			}
		}

		if catMatch {
			status = "SUCCESS"
			content = "Successfully set the output to the category: `" + category + "`"
			if len(catIDs) > 1 {
				status = "PENDING INPUT"
				content = "You have multiple categories that match the name: **" + category + "**. Please select the correct one below."
				catOptions := catMenu
				componentValue = []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    "categorymenu",
								Placeholder: "Select Category",
								MinValues:   1,
								MaxValues:   1,
								Options:     catOptions,
							},
						},
					},
				}
			} else {
				info := database.GuildInfo{
					GuildID:         i.GuildID,
					AuctionCategory: catIDs[0],
				}
				database.DB.Create(&info)
				database.DB.Model(&info).Update("AuctionCategory", catIDs[0])
			}
		}

	}

	info := database.GuildInfo{GuildID: i.GuildID}

	database.DB.Model(&info).Update("Currency", currency)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: componentValue,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("Auction Category Setup: __%s__", status),
					Description: content,
				},
			},
			Flags: 64,
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}

func CategorySelect(s *discordgo.Session, i *discordgo.InteractionCreate) {

	category := ""
	categoryID := i.MessageComponentData().Values[0]
	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		fmt.Println(err)
	}

	info := database.GuildInfo{
		GuildID:         i.GuildID,
		AuctionCategory: "categoryID",
	}

	database.DB.Create(&info)
	database.DB.Model(&info).Update("AuctionCategory", categoryID)

	for _, v := range channels {
		if v.Type == 4 {
			if categoryID == v.ID {
				category = v.Name
			}
		}
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Auction Category Setup: __SUCCESS__",
					Description: "Successfully set the output to the category: `" + category + "`",
				},
			},
			Components: []discordgo.MessageComponent{},
			Flags:      64,
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}

func AuctionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	options := ParseSubCommand(i)
	item := options["item"].(string)
	initialBid := options["startingbid"].(float64)
	info := database.GuildInfo{}
	currentTime := time.Now()
	duration, err := time.ParseDuration(fmt.Sprint(options["duration"].(float64)) + "h")
	if err != nil {
		fmt.Println(err)
	}
	endTime := currentTime.Add(duration)
	currency := info.Currency

	if len(item) > 100 {
		return
	}

	database.DB.First(&info, i.Interaction.GuildID)

	channelInfo := discordgo.GuildChannelCreateData{
		Name:     "💸│" + item,
		Type:     0,
		ParentID: info.AuctionCategory,
	}

	channel, err := s.GuildChannelCreateComplex(i.Interaction.GuildID, channelInfo)

	if err != nil {
		fmt.Println(err)
	}

	message, err := s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content: "",
		Embed: &discordgo.MessageEmbed{
			Title:       "Item: " + item,
			Description: fmt.Sprintf("Current Highest Bid: %s%s", currency, fmt.Sprint(initialBid)),
			Color:       0x00bfff,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "**Auction End Time:**",
					Value:  fmt.Sprintf("<t:%d>", endTime.Unix()),
					Inline: false,
				},
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "End Auction Early",
						Style: 4,
						Emoji: discordgo.ComponentEmoji{
							Name: "🛑",
						},
						CustomID: "endauction",
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println(err)
	}

	database.DB.Create(&database.Auction{
		ChannelID: message.ChannelID,
		Bid:       initialBid,
		MessageID: message.ID,
		EndTime:   endTime,
		Winner:    "No bidders",
		GuildID:	i.GuildID,
	})

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "",
			Components: []discordgo.MessageComponent{},
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "**Auction Started**",
					Description: "Auction has successfully been started, I might have some bugs to work out so please contact me if there is a failure.",
					Timestamp:   "",
					Color:       0x00bfff,
					Author:    &discordgo.MessageEmbedAuthor{
						Name:         "Auction hosted by: " + i.Member.Mention(),
					},
				},
			},
			Flags: 64,
		},
	})

	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(duration)
	AuctionEnd(channel.ID, i.GuildID)
}

func AuctionBid(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := ParseSubCommand(i)
	bidAmount := options["amount"].(float64)
	var info database.Auction
	var guildInfo database.GuildInfo
	info.ChannelID = i.ChannelID
	database.DB.First(&info, i.ChannelID)
	database.DB.First(&guildInfo, i.GuildID)
	currency := guildInfo.Currency
	var Content string

	if bidAmount > info.Bid {
		info.Bid = bidAmount
		info.Winner = fmt.Sprintf("<@%s>", i.Member.User.ID)
		Winner := info.Winner

		database.DB.Model(&info).Updates(info)

		updateAuction, err := s.ChannelMessage(info.ChannelID, info.MessageID)
		if err != nil {
			fmt.Println(err)
		}

		updateAuction.Embeds[0].Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "**Auction End Time:**",
				Value:  fmt.Sprintf("<t:%d>", info.EndTime.Unix()),
				Inline: false,
			},
			{
				Name:  "**Current Winner**",
				Value: fmt.Sprint(Winner),
			},
		}

		updateAuction.Embeds[0].Description = fmt.Sprintf("Current Highest Bid: %s%s", currency, fmt.Sprint(info.Bid))

		_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Components: updateAuction.Components,
			Embed:      updateAuction.Embeds[0],
			ID:         info.MessageID,
			Channel:    info.ChannelID,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		Content = "Bid has successfully been placed"
	} else {
		Content = "You must bid higher than: " + fmt.Sprint(info.Bid)
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: Content,
					Color: 0x00bfff,
				},
			},
			Flags: 64,
		},
	})
}

func AuctionEnd(ChannelID, GuildID string) {
	var info database.Auction
	var auctionInfo database.GuildInfo
	info.ChannelID = ChannelID
	database.DB.First(&info, ChannelID)
	database.DB.First(&auctionInfo, GuildID)

	messageSend := discordgo.MessageSend{
		Content:         "",
		Embed:           &discordgo.MessageEmbed{
			Title:       "",
			Description: "",
			Timestamp:   "",
			Color:       0x00bfff,
			Fields:      []*discordgo.MessageEmbedField{},
		},
	}

	Session.ChannelMessageSendComplex(auctionInfo.LogChannel, &messageSend)

	_, err := Session.ChannelDelete(ChannelID)
	if err != nil {
		fmt.Println(err)
	}
}

func AuctionEndButton(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Member.Permissions != 8 {
		fmt.Println("User "+i.Member.User.Username+" does not have correct permissions. User permissions: ", i.Member.Permissions)
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "",
			Components: []discordgo.MessageComponent{},
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "**Auction Ended**",
					Description: "",
					Timestamp:   "",
					Color:       0,
				},
			},
			AllowedMentions: &discordgo.MessageAllowedMentions{},
			Flags:           64,
		},
	})

	if err != nil {
		fmt.Println(err)
	}

	_, err = Session.ChannelDelete(i.ChannelID)
	if err != nil {
		fmt.Println(err)
	}
}

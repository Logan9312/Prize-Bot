package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	h "gitlab.com/logan9312/discord-auction-bot/helpers"
)

const (
	EventTypeAuction  = "Auction"
	EventTypeShop     = "Shop"
	EventTypeGiveaway = "Giveaway"
)

// TODO Rework this file
func EventFormat(s *discordgo.Session, data map[string]interface{}, eventType string, guildID string) (h.PresetResponse, error) {

	content := ""
	imageURL := ""
	embeds := []*discordgo.MessageEmbed{}

	if data["item"] != nil && len(data["item"].(string)) > 100 {
		return h.PresetResponse{}, fmt.Errorf("title cannot be over 100 characters long")
	}

	if data["image_url"] != nil {
		imageURL = data["image_url"].(string)
	}

	description := fmt.Sprintf("**Host:** <@%s>.\n", data["host"])

	if data["description"] != nil {
		description += fmt.Sprintf("**Description:** %s\n", data["description"])
	}

	if data["winners"] != nil {
		description += fmt.Sprintf("**Winners:** %d\n", int(data["winners"].(float64)))
	}

	if data["increment_min"] != nil {
		description += fmt.Sprintf("**Minimum Bid:** + %s above previous.\n", PriceFormat(data["increment_min"].(float64), guildID, data["currency"]))
	}

	if data["increment_max"] != nil {
		description += fmt.Sprintf("**Maximum Bid:** + %s above previous.\n", PriceFormat(data["increment_max"].(float64), guildID, data["currency"]))
	}

	if data["target_price"] != nil {
		description += "**Target Price:** The host has set a hidden target price for this auction.\n"
	}

	if data["integer_only"] != nil {
		description += fmt.Sprintf("**Integer Only:** %t.\n", data["integer_only"].(bool))
	}

	if data["snipe_extension"] != nil && data["snipe_range"] != nil {
		description += fmt.Sprintf("**Anti Snipe:** If a bid is placed within the last %s, the auction will be extended by %s.\n", data["snipe_range"], data["snipe_extension"].(time.Duration).String())
	}

	if data["buyout"] != nil {
		description += fmt.Sprintf("**Buyout Price:** %s.\n", PriceFormat(data["buyout"].(float64), guildID, data["currency"]))
	}

	auctionfields := []*discordgo.MessageEmbedField{
		{
			Name:  "__**Details:**__",
			Value: description,
		},
	}

	if data["end_time"] != nil {
		auctionfields = append(auctionfields, &discordgo.MessageEmbedField{
			Name:   "__**End Time**__",
			Value:  fmt.Sprintf("<t:%d:R>", data["end_time"].(time.Time).Unix()),
			Inline: true,
		})
	}

	if data["bid"] != nil {
		if data["winner"] != nil {
			auctionfields = append(auctionfields, &discordgo.MessageEmbedField{
				Name:   "__**Current Highest Bid:**__",
				Value:  PriceFormat(data["bid"].(float64), guildID, data["currency"]),
				Inline: true,
			})
		} else {
			auctionfields = append(auctionfields, &discordgo.MessageEmbedField{
				Name:   "__**Starting Bid:**__",
				Value:  PriceFormat(data["bid"].(float64), guildID, data["currency"]),
				Inline: true,
			})
		}
	}

	if data["winner"] != nil {
		auctionfields = append(auctionfields, &discordgo.MessageEmbedField{
			Name:   "__**Current Winner**__",
			Value:  fmt.Sprintf("<@%s>", data["winner"]),
			Inline: true,
		})
	}

	if eventType == EventTypeAuction {
		auctionfields = append(auctionfields, &discordgo.MessageEmbedField{
			Name:  "__**How to Bid**__",
			Value: "Use the /bid command or type `/bid {value}` in chat\nEx: `/bid 550`\n",
		})
	}

	if eventType == EventTypeGiveaway {
		auctionfields = append(auctionfields, &discordgo.MessageEmbedField{
			Name:   "__**How to Enter**__",
			Value:  "To enter, select the 🎁 reaction below! Removing your reaction will remove your entry.",
			Inline: false,
		})
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		fmt.Println("Error fetching guild: ", err)
		return h.PresetResponse{}, err
	}

	if data["alert_role"] != nil {
		content = fmt.Sprintf("<@&%s>", strings.Trim(data["alert_role"].(string), " "))
	}
	if data["alert_role"] == guildID {
		content = "@everyone"
	}

	components := []discordgo.MessageComponent{}

	if eventType == EventTypeAuction {
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "End Auction",
						Style: 4,
						Emoji: discordgo.ComponentEmoji{
							Name: "🛑",
						},
						CustomID: "endauction",
					},
					discordgo.Button{
						Label:    "Clear Chat",
						Style:    3,
						CustomID: "clearauction",
						Emoji: discordgo.ComponentEmoji{
							Name: "restart",
							ID:   "835685528917114891",
						},
						Disabled: false,
					},
				},
			},
		}
	}

	if eventType == EventTypeShop {
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Purchase",
						Style:    discordgo.SuccessButton,
						CustomID: "additem",
					},
				},
			},
		}
	}

	if data["bid_history"] != nil {
		if len(data["bid_history"].(string)) > 4095 {
			data["bid_history"] = data["bid_history"].(string)[len(data["bid_history"].(string))-4095:]
		}
		embeds = []*discordgo.MessageEmbed{{
			Title:       "**Bid History**",
			Description: data["bid_history"].(string),
			Color:       0x8073ff,
			Image: &discordgo.MessageEmbedImage{
				URL: "https://i.imgur.com/9wo7diC.png",
			},
		}}
	}

	return h.PresetResponse{
		Content:    content,
		Title:      fmt.Sprintf("%s Item: __**%s**__", eventType, data["item"]),
		Fields:     auctionfields,
		Thumbnail:  &discordgo.MessageEmbedThumbnail{URL: guild.IconURL()},
		Image:      &discordgo.MessageEmbedImage{URL: imageURL},
		Components: components,
		Embeds:     embeds,
		Files:      []*discordgo.File{},
	}, nil
}

func HasRole(m *discordgo.Member, roleID string) bool {
	for _, role := range m.Roles {
		if role == roleID {
			return true
		}
	}
	return false
}

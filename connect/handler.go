package connect

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	c "gitlab.com/logan9312/discord-auction-bot/commands"
	"gitlab.com/logan9312/discord-auction-bot/database"
	h "gitlab.com/logan9312/discord-auction-bot/helpers"
	"gitlab.com/logan9312/discord-auction-bot/subscriptions"
)

var commandMap = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"auction":        c.Auction,
	"bid":            c.AuctionBid,
	"profile":        c.Profile,
	"giveaway":       c.Giveaway,
	"shop":           c.Shop,
	"claim":          c.Claim,
	"privacy_policy": c.Privacy,
	"dev":            c.Dev,
	"premium":        subscriptions.Premium,
	"settings":       c.Settings,
}

var buttonMap = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"endauction":             c.AuctionEndButton,
	"claim_prize":            c.ClaimPrizeButton,
	"clearauction":           c.ClearAuctionButton,
	"delete_auction_queue":   c.DeleteAuctionQueue,
	"delete_auction_channel": c.DeleteAuctionChannel,
	"reroll_giveaway":        c.RerollGiveawayButton,
	"clear_auction_setup":    c.AuctionSetupClearButton,
	"clear_giveaway_setup":   c.GiveawaySetupClearButton,
	"clear_claim_setup":      c.ClaimSetupClearButton,
	"claim_cancel":           c.CancelButton,
	"claim_complete":         c.CompleteButton,
	"reopen_ticket":          c.ReopenTicket,
	"additem":                c.AddItem,
	"bid_history":            c.AuctionBidHistory,
}

var autoCompleteMap = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"auction":  c.AuctionAutoComplete,
	"giveaway": c.GiveawayAutoComplete,
}

func CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if i.Member == nil {
			h.ErrorResponse(s, i, "Commands cannot be run in a DM. Please contact support if you're not currently in a DM with the bot.")
			return
		}
		fmt.Println(i.ApplicationCommandData().Name, "is being run by:", i.Member.User.Username)
		if f, ok := commandMap[i.ApplicationCommandData().Name]; ok {
			f(s, i)
		} else {
			h.ErrorResponse(s, i, "Command response has not been set properly, please contact Logan to fix")
		}
	case discordgo.InteractionMessageComponent:
		fmt.Println(i.MessageComponentData().CustomID)
		if f, ok := buttonMap[strings.Split(i.MessageComponentData().CustomID, ":")[0]]; ok {
			f(s, i)
		} else {
			h.ErrorResponse(s, i, "Button response has not been set properly, please contact Logan to fix")
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		if f, ok := autoCompleteMap[i.ApplicationCommandData().Name]; ok {
			f(s, i)
		} else {
			h.ErrorResponse(s, i, "AutoComplete response has not been set properly, please contact Logan to fix")
		}
	}
}

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	bidValue := ""
	mentioned := false
	if m.Content == "" {
		return
	}

	for _, v := range m.Message.Mentions {
		if v.ID == s.State.User.ID {
			mentioned = true
		}
	}

	if !mentioned {
		return
	}

	splitString := strings.Split(m.Content, " ")

	for n, v := range splitString {
		if strings.ToLower(v) == "bid" {
			if len(splitString) <= n+1 {
				h.ErrorMessage(s, m.ChannelID, "Error Bidding. Your message must contain the bid amount after the word bid. Ex: Bid 500")
				return
			}
			bidValue = splitString[n+1]
			bidAmount, err := strconv.ParseFloat(bidValue, 64)
			if err != nil {
				h.ErrorMessage(s, m.ChannelID, err.Error())
				fmt.Println(err)
				return
			}
			response, err := c.AuctionBidFormat(s, database.Auction{
				ChannelID: m.ChannelID,
				Bid:       bidAmount,
				Winner:    m.Author.ID,
				GuildID:   m.GuildID,
			})
			if err != nil {
				fmt.Println(err)
				h.ErrorMessage(s, m.ChannelID, err.Error())
			}
			message, err := h.SuccessMessage(s, m.ChannelID, response)
			if err != nil {
				fmt.Println(err)
			}

			time.Sleep(10 * time.Second)
			err = s.ChannelMessageDelete(m.ChannelID, m.Message.ID)
			if err != nil {
				fmt.Println(err)
			}
			if message != nil {
				err = s.ChannelMessageDelete(m.ChannelID, message.ID)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

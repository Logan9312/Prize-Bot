package commands

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gitlab.com/logan9312/discord-auction-bot/database"
	"gitlab.com/logan9312/discord-auction-bot/events"
	h "gitlab.com/logan9312/discord-auction-bot/helpers"
	r "gitlab.com/logan9312/discord-auction-bot/responses"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/gorm"
)

var AuctionCommand = discordgo.ApplicationCommand{
	Name:        "auction",
	Description: "Put an item up for auction!",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "schedule",
			Description: "Schedule an auction",
			Options: []*discordgo.ApplicationCommandOption{
				Require(CommandOptionTimeUntil),
				Require(CommandOptionItem),
				Require(CommandOptionBid),
				Require(CommandOptionDuration),
				CommandOptionDescription,
				CommandOptionCurrency,
				CommandOptionUseCurrency,
				CommandOptionIntegerOnly,
				CommandOptionIncrementMax,
				CommandOptionIncrementMin,
				CommandOptionBuyout,
				CommandOptionTargetPrice,
				CommandOptionCategory,
				CommandOptionImage,
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "Create an Auction",
			Options: []*discordgo.ApplicationCommandOption{
				Require(CommandOptionItem),
				Require(CommandOptionBid),
				Require(CommandOptionDuration),
				CommandOptionDescription,
				CommandOptionCurrency,
				CommandOptionUseCurrency,
				CommandOptionIntegerOnly,
				CommandOptionIncrementMax,
				CommandOptionIncrementMin,
				CommandOptionBuyout,
				CommandOptionTargetPrice,
				CommandOptionCategory,
				CommandOptionImage,
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "queue",
			Description: "Display the current auction Queue",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "edit",
			Description: "Edit any auction details",
			Options: []*discordgo.ApplicationCommandOption{
				//CommandOptionQueueNumber
				CommandOptionItem,
				CommandOptionHost,
				CommandOptionBid,
				CommandOptionHost,
				CommandOptionWinner,
				CommandOptionExtend,
				CommandOptionDescription,
				CommandOptionCurrency,
				CommandOptionUseCurrency,
				CommandOptionCurrencySide,
				CommandOptionIntegerOnly,
				CommandOptionIncrementMax,
				CommandOptionIncrementMin,
				CommandOptionBuyout,
				CommandOptionTargetPrice,
				CommandOptionImage,
			},
		},
	},
}

var BidCommand = discordgo.ApplicationCommand{
	Name:        "bid",
	Description: "Bid on an Auction",
	Options: []*discordgo.ApplicationCommandOption{
		Require(CommandOptionBid),
	},
}

func Auction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	switch i.ApplicationCommandData().Options[0].Name {
	case "schedule":
		return AuctionSchedule(s, i)
	case "create":
		return AuctionCreate(s, i)
	case "queue":
		return AuctionQueue(s, i)
	case "edit":
		return AuctionEdit(s, i)
	}
	return fmt.Errorf("Unknown Auction command, please contact support")
}

func SaveAuction(auction database.Auction) error {
	return database.DB.Save(&auction).Error
}

func AuctionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := h.ParseSubCommand(i)
	errors := []string{}

	//Splits the auction names
	auctions := strings.Split(options["item"].(string), ";")

	if len(auctions) > 5 && !CheckPremiumGuild(i.GuildID) {
		r.PremiumError(s, i, "Free users can only start 5 auctions in bulk. Upgrade to premium to start up to 100 in bulk.")
	}

	if len(auctions) > 100 {
		return fmt.Errorf("You can only start 100 auctions in bulk at once. You attempted to start: %d.", len(auctions))
	}

	eventData := events.InteractionEvent(s, i, events.EventTypeAuction, options)

	//TODO Optimize selecting multiple auctions
	for _, item := range auctions {
		auctionMap := map[string]any{}
		for k, v := range options {
			auctionMap[k] = v
		}
		auctionMap["item"] = item

		channelID, err := AuctionHandler(s, i, auctionMap)
		if err != nil {
			errors = append(errors, err.Error())
		}

		_, err = r.FollowUpSuccessResponse(s, i, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "**Auction Starting**",
					Description: fmt.Sprintf("Auction has successfully been started in <#%s>!", channelID),
				},
			},
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("One or more auctions failed to start:\n", strings.Join(errors, "\n"))
	}
	return nil
}

func AuctionSchedule(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	//TODO Generate a preview of auction and schedule it
	return nil
}

func AuctionHandler(s *discordgo.Session, i *discordgo.InteractionCreate, event events.Event, data map[string]any) (channelID string, err error) {
	auctionSettings := &database.AuctionSetup{}
	currencySettings := &database.CurrencySetup{}

	duration, err := h.ParseTime(strings.ToLower(data["duration"].(string)))
	if err != nil {
		return "", r.Errorfmt(err)
	}

	result := database.DB.First(&auctionSettings, i.GuildID)
	if result.Error != nil {
		fmt.Println("Error getting auction settings: " + result.Error.Error())
	}

	result = database.DB.First(&currencySettings, i.GuildID)
	if result.Error != nil {
		fmt.Println("Error getting currency settings: " + result.Error.Error())
	}

	auctionMap["host"] = member.User.ID

	if auctionMap["currency"] == nil {
		auctionMap["currency"] = currencyMap["currency"]
	}

	for _, key := range []string{"category", "snipe_extension", "snipe_range", "currency_side", "integer_only", "alert_role", "channel_lock", "use_currency", "channel_prefix"} {
		if auctionMap[key] == nil {
			auctionMap[key] = auctionSetup[key]
		}
	}

	if auctionMap["schedule"] != nil {
		err = AuctionSchedule(s, auctionMap, duration)
		if err != nil {
			return "", fmt.Errorf("Error scheduling auction: %w", err)
		}
	} else {
		auctionMap["end_time"] = time.Now().Add(duration)

		channelID, err := AuctionStart(s, auctionMap)
		if err != nil {
			return channelID, fmt.Errorf("Error starting auction: %w", err)
		}

		return channelID, nil

	}

	return "", nil
}

func AuctionSchedule(s *discordgo.Session, auctionMap map[string]any, duration time.Duration) error {

	var AuctionQueue []database.AuctionQueue

	database.DB.Where(map[string]interface{}{"guild_id": auctionMap["guild_id"].(string)}).Find(&AuctionQueue)
	if len(AuctionQueue) >= 25 {
		return fmt.Errorf("You can only schedule 25 auctions in advance.")
	}

	startTimeDuration, err := h.ParseTime(strings.ToLower(auctionMap["schedule"].(string)))
	if err != nil {
		return err
	}

	auctionMap["end_time"] = time.Now().Add(duration).Add(startTimeDuration)
	auctionMap["start_time"] = time.Now().Add(startTimeDuration)
	delete(auctionMap, "schedule")

	result := database.DB.Model(database.AuctionQueue{}).Create(&auctionMap)
	if result.Error != nil {
		return fmt.Errorf("Error saving queued auction to database: %w", result.Error)
	}

	go AuctionStartTimer(s, auctionMap)

	return nil
}

func AuctionStartTimer(s *discordgo.Session, auctionMap map[string]any) {
	time.Sleep(time.Until(auctionMap["start_time"].(time.Time)))

	_, err := AuctionStart(s, auctionMap)
	if err != nil {
		fmt.Println(err)
	}
}

func AuctionEndTimer(s *discordgo.Session, auctionMap map[string]any) {
	time.Sleep(time.Until(auctionMap["end_time"].(time.Time)))
	err := AuctionEnd(s, auctionMap["channel_id"].(string), auctionMap["guild_id"].(string))
	if err != nil {
		fmt.Println(err)
	}
}

func AuctionHostCheck(auctionSetup map[string]any, member *discordgo.Member) bool {
	if auctionSetup["host_role"] == nil {
		return true
	}
	for _, v := range member.Roles {
		if v == auctionSetup["host_role"].(string) {
			return true
		}
	}

	return member.Permissions&discordgo.PermissionManageServer == 8
}

func AuctionStart(s *discordgo.Session, auctionMap map[string]interface{}) (string, error) {
	if auctionMap["id"] != nil {
		result := database.DB.Delete(database.AuctionQueue{}, auctionMap["id"])
		if result.Error != nil {
			fmt.Println(result.Error)
		}
	}

	auctionMessage, err := EventFormat(s, auctionMap, EventTypeAuction, auctionMap["guild_id"].(string))
	if err != nil {
		return "", err
	}

	delete(auctionMap, "start_time")
	delete(auctionMap, "id")
	delete(auctionMap, "alert_role")
	delete(auctionMap, "snipe_extension")
	delete(auctionMap, "snipe_range")

	if auctionMap["category"] == nil {
		auctionMap["category"] = ""
	}
	//TODO Make prefix work for editing and fix prefix working on create channel.
	if auctionMap["channel_prefix"] == nil {
		auctionMap["channel_prefix"] = "💸│"
	}

	if auctionMap["channel_lock"] != true {
		channel, err := s.GuildChannelCreateComplex(auctionMap["guild_id"].(string), discordgo.GuildChannelCreateData{
			Name:     auctionMap["channel_prefix"].(string) + auctionMap["item"].(string),
			Type:     discordgo.ChannelTypeGuildText,
			ParentID: auctionMap["category"].(string),
		})
		if err != nil {
			return "", err
		}
		auctionMap["channel_id"] = channel.ID
	}

	message, err := h.SuccessMessage(s, auctionMap["channel_id"].(string), auctionMessage)
	if err != nil {
		return auctionMap["channel_id"].(string), err
	}
	auctionMap["message_id"] = message.ID

	delete(auctionMap, "channel_prefix")
	delete(auctionMap, "category")

	result := database.DB.Model(database.Auction{}).Create(auctionMap)
	if result.Error != nil {
		return auctionMap["channel_id"].(string), fmt.Errorf("Error saving auction to database, auction will not work: %w", result.Error)
	}

	go AuctionEndTimer(s, auctionMap)

	return auctionMap["channel_id"].(string), nil
}

func AuctionBid(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	options := h.ParseSlashCommand(i)

	err := AuctionBidPlace(s, options["amount"].(float64), i.Member, i.ChannelID, i.GuildID)
	if err != nil {
		return err
	}

	err = h.SuccessResponse(s, i, h.PresetResponse{
		Title: "Bid has been successfully placed!",
	})
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func AuctionBidPlace(s *discordgo.Session, amount float64, member *discordgo.Member, channelID, guildID string) error {

	auctionSetup := map[string]any{}
	auctionMap := map[string]any{}
	userMap := map[string]any{}
	p := message.NewPrinter(language.English)

	result := database.DB.Model(database.Auction{}).First(&auctionMap, channelID)
	if result.Error != nil {
		return fmt.Errorf("Error fetching auction data from the database. Error Message: " + result.Error.Error())
	}
	result = database.DB.Model(database.AuctionSetup{}).First(&auctionSetup, guildID)
	if result.Error != nil {
		fmt.Println(result.Error)
	}

	//FIXME This was just a quick fix to prevent both bots trying to place a bit. Rework if needed once I can save auctions with bot ID
	message, err := s.ChannelMessage(channelID, auctionMap["message_id"].(string))
	if err == nil {
		if message.Author.ID != s.State.User.ID {
			return nil
		}
	}

	if auctionSetup["snipe_range"] != nil && auctionSetup["snipe_extension"] != nil {
		if time.Until(auctionMap["end_time"].(time.Time)) < auctionSetup["snipe_range"].(time.Duration) {
			auctionMap["end_time"] = auctionMap["end_time"].(time.Time).Add(auctionSetup["snipe_extension"].(time.Duration))
			h.SuccessMessage(s, channelID, h.PresetResponse{
				Title:       "**Anti-Snipe Activated!**",
				Description: fmt.Sprintf("New End Time: <t:%d>", auctionMap["end_time"].(time.Time).Unix()),
			})
		}
	}

	if auctionMap["use_currency"] != nil && auctionMap["use_currency"].(bool) {
		result = database.DB.Model(database.UserProfile{}).First(&userMap, map[string]any{
			"guild_id": guildID,
			"user_id":  member.User.ID,
		})
		if result.Error != nil {
			fmt.Println(result.Error)
		}
		if userMap["balance"] == nil {
			return fmt.Errorf("You have %s currency and cannot bid on this auction.", PriceFormat(0, guildID, auctionMap["currency"]))
		}
		if amount > userMap["balance"].(float64) {
			return fmt.Errorf("You do not have enough currency to bid on this auction. You need %s and you have %s", PriceFormat(amount, guildID, auctionMap["currency"]), PriceFormat(userMap["balance"].(float64), guildID, auctionMap["currency"]))
		}
	}

	//Checking if the auction has ended.
	if auctionMap["end_time"].(time.Time).Before(time.Now()) {
		return fmt.Errorf("cannot Bid, Auction has ended")
	}

	if auctionMap["buyout"] == nil || (auctionMap["buyout"] != nil && amount < auctionMap["buyout"].(float64)) {
		//Checking if the auction is capped and the current winner is bidding.
		if member.User.ID == auctionMap["winner"] && auctionMap["increment_max"] != nil {
			return fmt.Errorf("cannot out bid yourself on a capped bid auction")
		}

		//Checking if integer only bidding is enabled.
		if auctionMap["integer_only"] != nil && auctionMap["integer_only"].(bool) && amount != math.Floor(amount) {
			return fmt.Errorf("Your bid must be an integer for this auction! For example: " + fmt.Sprint(math.Floor(amount)) + " instead of " + PriceFormat(amount, guildID, auctionMap["currency"]))
		}

		//Checking if bid is higher than minimum increment.
		if auctionMap["increment_min"] != nil && amount-auctionMap["bid"].(float64) < auctionMap["increment_min"].(float64) {
			return fmt.Errorf("Bid must be higher than the previous bid by: %s\n\u200b", PriceFormat(auctionMap["increment_min"].(float64), guildID, auctionMap["currency"]))
		}

		//Checking if bid is lower than maximum increment.
		if auctionMap["increment_max"] != nil && amount-auctionMap["bid"].(float64) > auctionMap["increment_max"].(float64) {
			return fmt.Errorf("Bid must be no more than %s higher than the previous bid. \n\u200b", PriceFormat(auctionMap["increment_max"].(float64), auctionMap["guild_id"].(string), auctionMap["currency"]))
		}
	}

	if amount < auctionMap["bid"].(float64) {
		return fmt.Errorf("You must bid higher than: " + PriceFormat(auctionMap["bid"].(float64), auctionMap["guild_id"].(string), auctionMap["currency"]))
	}

	if amount == auctionMap["bid"].(float64) && auctionMap["winner"] != nil {
		return fmt.Errorf("You must bid higher than: " + PriceFormat(auctionMap["bid"].(float64), auctionMap["guild_id"].(string), auctionMap["currency"]))
	}

	if auctionMap["bid_history"] == nil {
		auctionMap["bid_history"] = ""
	}

	auctionMap["bid"] = amount
	auctionMap["winner"] = member.User.ID
	auctionMap["bid_history"] = auctionMap["bid_history"].(string) + "\n-> " + member.User.Username + ": " + strings.TrimRight(strings.TrimRight(p.Sprintf("%f", amount), "0"), ".")

	if auctionMap["buyout"] != nil && amount >= auctionMap["buyout"].(float64) {
		auctionMap["end_time"] = time.Now()
	}

	result = database.DB.Model(database.Auction{
		ChannelID: channelID,
	}).Updates(auctionMap)
	if result.Error != nil {
		return result.Error
	}

	if auctionMap["buyout"] != nil && amount >= auctionMap["buyout"].(float64) {
		go AuctionEnd(s, channelID, guildID)
	}

	auctionMap["snipe_extension"] = auctionSetup["snipe_extension"]
	auctionMap["snipe_range"] = auctionSetup["snipe_range"]

	m, err := EventFormat(s, auctionMap, EventTypeAuction, guildID)
	if err != nil {
		return err
	}

	_, err = h.SuccessMessageEdit(s, channelID, auctionMap["message_id"].(string), m)
	if err != nil {
		return err
	}

	return nil
}

// AuctionEdit handles the AuctionEdit event
func AuctionEdit(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	options := h.ParseSubCommand(i)

	if options["image"] != nil {
		options["image_url"] = i.ApplicationCommandData().Resolved.Attachments[options["image"].(string)].URL
		delete(options, "image")
	}

	if options["queue_number"] != nil {
		err := AuctionQueueUpdate(options, i.GuildID)
		if err != nil {
			return err
		}
	} else {
		err := AuctionUpdate(s, options, i.Member, i.ChannelID, i.GuildID)
		if err != nil {
			return err
		}
	}

	return h.SuccessResponse(s, i, h.PresetResponse{
		Title:       "Success",
		Description: "Auction has successfully been edited",
	})
}

// AuctionUpdate updates a currently running auction
func AuctionUpdate(s *discordgo.Session, options map[string]any, member *discordgo.Member, channelID, guildID string) error {

	auctionMap := map[string]any{}
	p := message.NewPrinter(language.English)

	result := database.DB.Model(database.Auction{}).First(&auctionMap, channelID)
	if result.Error != nil {
		return result.Error
	}

	if member.Permissions&(1<<3) != 8 && member.User.ID != auctionMap["host"] {
		return fmt.Errorf("User must have be host or have administrator permissions to run this command. User: <@%s>, Host: <@%s>", member.User.ID, auctionMap["host"])
	}

	editedOptions := ""

	for key, value := range options {
		switch key {
		case "winner", "host":
			editedOptions += fmt.Sprintf("\n\u3000- %s set to: <@%s>", key, value)
		case "bid":
			editedOptions += fmt.Sprintf("\n\u3000- %s set to: %s", key, strings.TrimRight(strings.TrimRight(p.Sprintf("%f", options["bid"].(float64)), "0"), "."))
		default:
			editedOptions += fmt.Sprintf("\n\u3000- %s set to: %s", key, p.Sprint(value))
		}
	}

	if options["extend"] != nil {
		extraDuration, err := h.ParseTime(strings.ToLower(options["extend"].(string)))
		if err != nil {
			return err
		}
		options["end_time"] = auctionMap["end_time"].(time.Time).Add(extraDuration)
		delete(options, "extend")
	}

	if auctionMap["bid_history"] == nil {
		auctionMap["bid_history"] = ""
	}

	options["bid_history"] = auctionMap["bid_history"].(string) + "\n-> Auction edited by " + member.User.Username + ":" + editedOptions

	result = database.DB.Model(database.Auction{
		ChannelID: channelID,
	}).Updates(options)
	if result.Error != nil {
		return result.Error
	}

	if options["bid"] != nil && options["winner"] != nil {
		member, err := s.GuildMember(guildID, options["winner"].(string))
		if err != nil {
			return err
		}
		options["bid_history"] = options["bid_history"].(string) + "\n-> " + member.User.Username + ": " + strings.TrimRight(strings.TrimRight(p.Sprintf("%f", options["bid"].(float64)), "0"), ".")
	}

	if options["item"] != nil {
		channel, err := s.Channel(channelID)
		if err != nil {
			fmt.Println(err)
		}
		_, err = s.ChannelEditComplex(channelID, &discordgo.ChannelEdit{
			Name:     "💸│" + options["item"].(string),
			Position: channel.Position,
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	result = database.DB.Model(database.Auction{}).First(&auctionMap, channelID)
	if result.Error != nil {
		return result.Error
	}

	message, err := EventFormat(s, auctionMap, EventTypeAuction, guildID)
	if err != nil {
		return err
	}

	_, err = h.SuccessMessageEdit(s, channelID, auctionMap["message_id"].(string), message)
	if err != nil {
		return err
	}

	return nil
}

//FIXME Fix Auction Queue

func AuctionQueueUpdate(options map[string]any, guildID string) error {

	//Need to fix this since the database covers all guilds.
	guildQueue := []database.AuctionQueue{}

	result := database.DB.Where(map[string]interface{}{"guild_id": guildID}).Find(&guildQueue)
	if result.Error != nil {
		return result.Error
	}

	queueNumber := guildQueue[int(options["queue_number"].(float64))-1].ID

	delete(options, "queue_number")

	result = database.DB.Model(database.AuctionQueue{
		ID: queueNumber,
	}).Updates(options)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

//Buttons

func AuctionBidHistory(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	claimMap := map[string]interface{}{}

	result := database.DB.Model(database.Claim{}).First(claimMap, i.Message.ID)

	if result.Error != nil {
		return result.Error
	}

	if claimMap["bid_history"] == nil {
		return fmt.Errorf("No bid history found for this auction.")
	}

	bidHistory := claimMap["bid_history"].(string)

	if len(bidHistory) > 4095 {
		bidHistory = bidHistory[len(bidHistory)-4095:]
	}

	return h.SuccessResponse(s, i, h.PresetResponse{
		Title:       "**Bid History**",
		Description: bidHistory,
		Image: &discordgo.MessageEmbedImage{
			URL: "https://i.imgur.com/9wo7diC.png",
		},
	})
}

//Extra Responses

func AuctionAutoComplete(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var choices []*discordgo.ApplicationCommandOptionChoice
	var focusedData discordgo.ApplicationCommandInteractionDataOption
	var choiceName string

	switch i.ApplicationCommandData().Options[0].Name {
	case "create":
		options := h.ParseSubCommand(i)

		if options["duration"] != nil {
			choices = TimeSuggestions(options["duration"].(string))
		} else if options["schedule"] != nil {
			choices = TimeSuggestions(options["schedule"].(string))
		} else if options["snipe_range"] != nil {
			choices = TimeSuggestions(options["snipe_range"].(string))
		} else if options["snipe_extension"] != nil {
			choices = TimeSuggestions(options["snipe_extension"].(string))
		} else {
			choices = []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "",
					Value: "",
				},
			}
		}
	case "edit":
		currentValues := map[string]interface{}{}

		for _, v := range i.ApplicationCommandData().Options[0].Options {
			if v.Focused {
				focusedData = *v
			}
		}

		database.DB.Model(database.Auction{}).First(&currentValues, i.ChannelID)

		name := focusedData.Name

		if currentValues[name] != nil || name == "extend" {
			switch name {
			case "extend":
				choiceName = fmt.Sprintf("Time remaining: %s", time.Until(currentValues["end_time"].(time.Time)))
			default:
				choiceName = fmt.Sprintf("Current Value: %v", currentValues[name])
			}
		} else {
			choiceName = "Not Configured"
		}

		choices = []*discordgo.ApplicationCommandOptionChoice{
			{
				Name:  choiceName,
				Value: focusedData.Value,
			},
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		fmt.Println("Response Error:", err)
		return err
	}

	return nil
}

func TimeSuggestions(input string) []*discordgo.ApplicationCommandOptionChoice {

	re := regexp.MustCompile(`[0-9]+\.?[0-9]*`)
	input = re.FindString(input)

	if input == "" {
		return []*discordgo.ApplicationCommandOptionChoice{}
	}

	_, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	choices := []*discordgo.ApplicationCommandOptionChoice{
		{
			Name:  input + "d",
			Value: input + "d",
		},
		{
			Name:  input + "h",
			Value: input + "h",
		},
		{
			Name:  input + "m",
			Value: input + "m",
		},
		{
			Name:  input + "s",
			Value: input + "s",
		},
	}
	return choices
}

func AuctionEnd(s *discordgo.Session, channelID, guildID string) error {

	auctionMap := map[string]any{}
	auctionSetup := map[string]any{}

	result := database.DB.Model(database.AuctionSetup{}).First(&auctionSetup, guildID)
	if result.Error != nil {
		fmt.Println(result.Error.Error())
	}

	result = database.DB.Model(database.Auction{}).First(&auctionMap, channelID)
	if result.Error != nil {
		fmt.Println(result.Error.Error())
	}

	if len(auctionMap) == 0 {
		return fmt.Errorf("Auction Contained no Data")
	}

	if auctionMap["end_time"] == nil {
		s.ChannelMessageSend("943175605858496602", fmt.Sprint(auctionMap))
	}

	//Pause auction ending until end time if the auction is not over yet.
	if auctionMap["end_time"] != nil && auctionMap["end_time"].(time.Time).After(time.Now()) {
		fmt.Println(time.Until(auctionMap["end_time"].(time.Time)))
		time.Sleep(time.Until(auctionMap["end_time"].(time.Time)))
		err := AuctionEnd(s, channelID, guildID)
		return err
	}

	if auctionMap["message_id"] == nil {
		auctionMap["message_id"] = ""
	}

	message := discordgo.NewMessageEdit(channelID, auctionMap["message_id"].(string))
	messageEmbeds, err := s.ChannelMessage(channelID, auctionMap["message_id"].(string))
	if err != nil {
		result = database.DB.Delete(database.Auction{
			ChannelID: channelID,
		})
		if result.Error != nil {
			fmt.Println(result.Error.Error())
		}
		return err
	}

	if auctionSetup["log_channel"] == nil {
		fmt.Println("Log channel has not been set for guild: " + guildID)
		// TODO Instead of sending an error message, update the auction with the error
		_, err := h.ErrorMessage(s, channelID, "Auction cannot end because log channel has not been set. Please setup an auction log using `/settings auction`. You might need to end the auction manually after setting the channel.")
		if err != nil {
			return err
		}
		return err
	}

	if auctionMap["target_price"] != nil && auctionMap["target_price"].(float64) > auctionMap["bid"].(float64) {
		auctionMap["target_message"] = fmt.Sprintf("The host had set a target price of %s that has not been reached.", PriceFormat(auctionMap["target_price"].(float64), guildID, auctionMap["currency"]))
		delete(auctionMap, "winner")
	}

	if auctionMap["use_currency"] != nil && auctionMap["use_currency"].(bool) && auctionMap["winner"] != nil {
		err = CurrencySubtractUser(guildID, auctionMap["winner"].(string), auctionMap["bid"].(float64))
		if err != nil {
			return err
		}
		err = CurrencyAddUser(guildID, auctionMap["host"].(string), auctionMap["bid"].(float64))
		if err != nil {
			return err
		}
		//TODO Add in a message about this when the auction ends
		//TODO Move this to the claiming process.
	}

	if auctionMap["buyout"] != nil {
		if auctionMap["bid"].(float64) >= auctionMap["buyout"].(float64) {
			auctionMap["buyout_message"] = fmt.Sprintf("%s\n\u200bBUYOUT!", PriceFormat(auctionMap["buyout"].(float64), auctionMap["guild_id"].(string), auctionMap["currency"]))
		}
	}

	auctionMap["log_channel"] = auctionSetup["log_channel"]

	err = ClaimOutput(s, auctionMap, "Auction")
	if err != nil {
		return fmt.Errorf("Claim Output Error: " + err.Error())
	}

	if message != nil {
		message.Embeds = append(messageEmbeds.Embeds, &discordgo.MessageEmbed{
			Title:       "Auction has ended!",
			Description: "Thank you for participating!",
			Color:       0x32CD32,
			Image: &discordgo.MessageEmbedImage{
				URL: "https://c.tenor.com/MvFFZxXwdpwAAAAC/sold-ray.gif",
			},
		})
		message.Components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Support Server",
						Style: discordgo.LinkButton,
						Emoji: discordgo.ComponentEmoji{
							Name:     "logo",
							ID:       "889025400120950804",
							Animated: false,
						},
						URL: "https://discord.gg/RxP2z5NGtj",
					},
				},
			},
		}
		_, err = s.ChannelMessageEditComplex(message)
		if err != nil {
			fmt.Println(err)
		}
	}

	if auctionMap["channel_lock"] != true {
		time.Sleep(30 * time.Second)
		_, err = s.ChannelDelete(channelID)
		if err != nil {
			fmt.Println(err)
		}
	}

	result = database.DB.Delete(database.Auction{}, channelID)
	if result.Error != nil {
		fmt.Println(result.Error.Error())
	}

	return nil
}

func AuctionQueue(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	var AuctionQueueInfo []database.AuctionQueue
	var AuctionQueue []database.AuctionQueue
	var fields []*discordgo.MessageEmbedField
	var selectOptions []discordgo.SelectMenuOption

	database.DB.Find(&AuctionQueueInfo)

	for _, v := range AuctionQueueInfo {
		if v.GuildID == i.GuildID {
			AuctionQueue = append(AuctionQueue, v)
		}
	}

	sort.Slice(AuctionQueue, func(i, j int) bool { return AuctionQueue[i].StartTime.Unix() < AuctionQueue[j].StartTime.Unix() })

	for _, v := range AuctionQueue {

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("**%s. %s**", fmt.Sprint(len(fields)+1), v.Item),
			Value: fmt.Sprintf("**Start time:** <t:%d:R>\n**End Time:** <t:%d>\n**Starting Price:** %s\n\u200b", v.StartTime.Unix(), v.EndTime.Unix(), PriceFormat(v.Bid, v.GuildID, v.Currency)),
		})
		selectOptions = append(selectOptions, discordgo.SelectMenuOption{
			Label:       v.Item,
			Value:       fmt.Sprint(v.ID),
			Description: fmt.Sprintf("Start time: %s", v.StartTime.String()),
		})
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "delete_auction_queue",
					Placeholder: "💣 Remove auction from queue",
					MinValues:   h.Ptr(0),
					MaxValues:   len(AuctionQueue),
					Options:     selectOptions,
				},
			},
		},
	}

	if len(AuctionQueue) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "**No Scheduled Auctions**",
			Value:  "Use the `schedule` parameter when creating auctions to plan them in advance!",
			Inline: false,
		})
		components = []discordgo.MessageComponent{}
	}

	err := h.SuccessResponse(s, i, h.PresetResponse{
		Title:       "**Auction Queue**",
		Description: "Displays upcoming auctions!",
		Fields:      fields,
		Components:  components,
	})

	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func AuctionEndButton(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	auctionMap := map[string]interface{}{}

	result := database.DB.Model(database.Auction{}).First(&auctionMap, i.ChannelID)
	if result.Error != nil {
		fmt.Println(result.Error)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label: "Delete Auction Channel",
								Style: 4,
								Emoji: discordgo.ComponentEmoji{
									Name: "🛑",
								},
								CustomID: "delete_auction_channel",
							},
							discordgo.Button{
								Label: "Support Server",
								Style: discordgo.LinkButton,
								Emoji: discordgo.ComponentEmoji{
									Name:     "logo",
									ID:       "889025400120950804",
									Animated: false,
								},
								URL: "https://discord.gg/RxP2z5NGtj",
							},
						},
					},
				},
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Error",
						Description: "There was an error. Please contact support if you need help.\n**If you would like to close this auction channel, press the button below.**",
						Color:       0xff0000,
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:  "**Error Message:**",
								Value: result.Error.Error(),
							},
						},
					},
				},
				Flags: 64,
			},
		})
		return nil
	}

	if i.Member.Permissions&(1<<3) != 8 && i.Member.User.ID != auctionMap["host"] {
		return fmt.Errorf("User must have administrator permissions or be host to run this command")
	}

	err := h.SuccessResponse(s, i, h.PresetResponse{
		Title:       "End Auction",
		Description: "Auction Ending...",
	})

	if err != nil {
		fmt.Println(err)
	}

	result = database.DB.Model(database.Auction{
		ChannelID: i.ChannelID,
	}).Update("end_time", time.Now())
	if result.Error != nil {
		return result.Error
	}

	err = AuctionEnd(s, i.ChannelID, i.GuildID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteAuctionChannel(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	_, err := s.ChannelDelete(i.ChannelID)
	if err != nil {
		fmt.Println(err)
		h.ErrorResponse(s, i, err.Error())
	}
	return nil
}

func ClearAuctionButton(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	var auctionInfo database.Auction

	result := database.DB.First(&auctionInfo, i.ChannelID)
	if result.Error != nil {
		fmt.Println("Error checking for auction. Auction has likely ended")
		auctionInfo.Host = ""
	}

	if i.Member.Permissions&(1<<3) != 8 && i.Member.User.ID != auctionInfo.Host {
		return fmt.Errorf("User must be host or have administrator permissions to run this command")
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	})

	for {
		messageIDs := make([]string, 0)
		messages, err := s.ChannelMessages(i.ChannelID, 100, "", i.Message.ID, "")
		if err != nil {
			h.DeferredErrorResponse(s, i, err.Error())
			return nil
		}

		for _, v := range messages {
			if !v.Pinned {
				messageIDs = append(messageIDs, v.ID)
			}
		}

		if len(messageIDs) == 0 {
			break
		}

		err = s.ChannelMessagesBulkDelete(i.ChannelID, messageIDs)
		if err != nil {
			h.DeferredErrorResponse(s, i, err.Error())
			return nil
		}
	}

	_, err := h.DeferredResponse(s, i, h.PresetResponse{
		Title:       "Success!",
		Description: "Clearing Chat",
	})

	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func DeleteAuctionQueue(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	if i.Member.Permissions&(1<<3) != 8 {
		return fmt.Errorf("User must have administrator permissions to run this command")
	}

	IDs := i.MessageComponentData().Values

	for _, v := range IDs {
		database.DB.Delete(database.AuctionQueue{}, v)
	}

	AuctionQueue(s, i)
	return nil
}

func AuctionSetupClearButton(s *discordgo.Session, i *discordgo.InteractionCreate) error {

	options := i.MessageComponentData().Values
	clearedMap := map[string]interface{}{}

	info := database.AuctionSetup{
		GuildID: i.GuildID,
	}

	clearedSettings := "No Settings Cleared!"
	if len(options) > 0 {
		clearedSettings = ""
	}

	for _, v := range options {
		clearedSettings += fmt.Sprintf("• %s\n", strings.Title(strings.ReplaceAll(v, "_", " ")))
		clearedMap[v] = gorm.Expr("NULL")
	}

	database.DB.Model(&info).Updates(clearedMap)

	h.SuccessResponse(s, i, h.PresetResponse{
		Title:       "**Cleared Auction Settings**",
		Description: "You have successfully cleared the following settings. Run `/settings auction` to see your changes.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "**Cleared Settings**",
				Value: clearedSettings,
			},
		},
	})
	return nil
}

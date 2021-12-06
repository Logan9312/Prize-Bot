package connect

import (
	"fmt"
	"strings"
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
			&commands.BidCommand,
			&commands.GiveawayCommand,
		},
		prod: []*discordgo.ApplicationCommand{
			&commands.HelpCommand,
			&commands.AuctionCommand,
			&commands.BidCommand,
			&commands.GiveawayCommand,
		},
	}

	fmt.Println(botName + " Starting Up...")
	var s *discordgo.Session
	var err error

	s, err = discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("discordgo.New error:" + err.Error())
	}

	s.ChannelMessageSend("915768633752449054", "Bot has finished restarting")

	commands.Session = s

	err = s.Open()

	if err != nil {
		fmt.Println("s.Open error: " + err.Error())
		return
	}

	//Builds local commands
	if environment == "local" {
		for _, v := range s.State.Guilds {
			_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, v.ID, c.local)
			fmt.Println("Commands added to guild: " + v.Name)
			if err != nil {
				fmt.Println("Bulk Overwrite Error:", err)
			}
		}
		commands.HelpBuilder(c.local)
		database.DB.Create(database.AuctionSetup{
			GuildID:    "915767892467920967",
			Category:   "915768615742103625",
			LogChannel: "915768633752449054",
		})
	}

	//Builds prod commands
	if environment == "prod" {
		_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", c.prod)
		if err != nil {
			fmt.Println("Bulk Overwrite Error:", err)
		}
		commands.HelpBuilder(c.prod)
	}

	s.AddHandler(CommandHandler)
	s.AddHandler(MessageHandler)

	DataFix()

	Timers(s)

	err = s.UpdateGameStatus(0, "Bot Version v0.9")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(botName + " Startup Complete!")
}

func Timers(s *discordgo.Session) {

	var Auctions []database.Auction
	var AuctionQueue []database.AuctionQueue
	var Giveaways []database.Giveaway

	fmt.Println("Beginning Startup Timers")

	database.DB.Find(&Auctions)
	for _, v := range Auctions {
		go AuctionEndTimer(v, s)
	}

	database.DB.Find(&AuctionQueue)
	for _, v := range AuctionQueue {
		go AuctionStartTimer(v, s)
	}

	database.DB.Find(&Giveaways)
	for _, v := range Giveaways {
		go GiveawayEndTimer(v, s)
	}
}

func AuctionEndTimer(v database.Auction, s *discordgo.Session) {
	fmt.Println("Auction Timer Re-Started: ", v.Item, "GuildID: ", v.GuildID, "ImageURL", v.ImageURL, "Host", v.Host, "End Time", v.EndTime.String())
	if v.EndTime.Before(time.Now()) {
		commands.AuctionEnd(v.ChannelID, v.GuildID)
	} else {
		time.Sleep(time.Until(v.EndTime))
		commands.AuctionEnd(v.ChannelID, v.GuildID)
	}
}

func AuctionStartTimer(v database.AuctionQueue, s *discordgo.Session) {
	fmt.Println("Auction Re-Queued: ", v.Item, "GuildID: ", v.GuildID, "ImageURL", v.ImageURL, "Host", v.Host, "Start Time", v.StartTime.String())
	if v.StartTime.Before(time.Now()) {
		commands.AuctionCreate(s, v)
	} else {
		time.Sleep(time.Until(v.StartTime))
		commands.AuctionCreate(s, v)
	}
}

func GiveawayEndTimer(v database.Giveaway, s *discordgo.Session) {
	fmt.Println("Giveaway Timer Re-Started: ", v.Item, "GuildID: ", v.GuildID, "ImageURL", v.ImageURL, "Host", v.Host, "End Time", v.EndTime.String())
	if v.EndTime.Before(time.Now()) {
		if v.Finished {
			time.Sleep(time.Until(v.EndTime.Add(24 * time.Hour)))
			database.DB.Delete(database.Giveaway{}, v.MessageID)
		} else {
			commands.GiveawayEnd(commands.Session, v.MessageID)
		}
	} else {
		time.Sleep(time.Until(v.EndTime))
		commands.GiveawayEnd(commands.Session, v.MessageID)
	}
}

func DataFix() {
	auctiondata := []database.AuctionSetup{}

	database.DB.Find(&auctiondata)

	for _, v := range auctiondata {
		if v.AlertRole != "" {
			database.DB.Model(&database.AuctionSetup{
				GuildID: v.GuildID,
			}).Update("alert_role", strings.Trim(v.AlertRole, "<@&>"))
		}
	}
}

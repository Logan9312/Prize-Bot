package events

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	h "gitlab.com/logan9312/discord-auction-bot/helpers"
)

type Event struct {
	ID          uint `gorm:"primaryKey"`
	BotID       string
	EventType   EventType
	GuildID     string
	Host        string
	ChannelID   *string
	MessageID   *string
	StartTime   *time.Time
	EndTime     *time.Time
	ImageURL    *string
	Description *string
	Note        *string
}

type EventType string

const (
	EventTypeAuction  EventType = "Auction"
	EventTypeShop     EventType = "Shop"
	EventTypeGiveaway EventType = "Giveaway"
)

func (event Event) StartTimers() {

}

func InteractionEvent(s *discordgo.Session, i *discordgo.InteractionCreate, et EventType, options map[string]any) error {
	eventData := &Event{
		BotID:     s.State.User.ID,
		EventType: et,
		GuildID:   i.GuildID,
	}

	if options["image"] != nil {
		eventData.ImageURL = h.ImageToURL(i, options["image"].(string))
		delete(options, "image")
	}

	if options["duration"] != nil {
		duration, err := h.ParseTime(options["duration"].(string))
		if err != nil {
			return fmt.Errorf("Error parsing time input: %w", err)
		}
		eventData.EndTime = h.Ptr(time.Now().Add(duration))
		delete(options, "duration")
	}
}

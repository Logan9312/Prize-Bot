package events

import (
	"time"
)

type Event struct {
	ID        uint `gorm:"primaryKey"`
	BotID     string
	EventType string
	GuildID   string
	ChannelID *string
	MessageID *string
	StartTime *time.Time
	EndTime   *time.Time
	ImageURL  *string
}

func (event Event) StartTimers() {

}

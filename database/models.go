package database

import (
	"time"

	"gitlab.com/logan9312/discord-auction-bot/events"
)

type DevSetup struct {
	BotID   string `gorm:"primaryKey"`
	Version string
}

type WhiteLabels struct {
	BotID    string `gorm:"primaryKey;autoIncrement:false"`
	UserID   string `gorm:"primaryKey;autoIncrement:false"`
	BotToken string
}

type AuctionSetup struct {
	GuildID         string `gorm:"primaryKey"`
	Category        string
	AlertRole       string
	Currency        string
	LogChannel      string
	HostRole        string
	SnipeExtension  time.Duration
	SnipeRange      time.Duration
	CurrencySide    string
	IntegerOnly     bool
	ChannelOverride string
	ChannelLock     bool
	ChannelPrefix   string
	UseCurrency     bool
}

type Auction struct {
	Event        events.Event
	EventID      string
	Item         string
	Bid          float64
	Winner       string
	IncrementMin float64
	IncrementMax float64
	TargetPrice  float64
	Buyout       float64
	IntegerOnly  bool
	BidHistory   string
	ChannelLock  bool
	Currency     *Currency
}

type AuctionQueue struct {
	ID              int `gorm:"primaryKey"`
	ChannelID       string
	Bid             float64
	StartTime       time.Time
	EndTime         time.Time
	GuildID         string
	Item            string
	Host            string
	Currency        string
	IncrementMin    float64
	IncrementMax    float64
	Description     string
	ImageURL        string
	Category        string
	TargetPrice     float64
	Buyout          float64
	CurrencySide    string
	IntegerOnly     bool
	SnipeExtension  time.Duration
	SnipeRange      time.Duration
	AlertRole       string
	Note            string
	ChannelOverride string
	ChannelLock     bool
	UseCurrency     bool
	ChannelPrefix   string
}

// ClaimSetup FromMake sure to remove LogChannel and ClaimMessage from auction log
type ClaimSetup struct {
	GuildID         string `gorm:"primaryKey"`
	Category        string
	StaffRole       string
	Instructions    string
	LogChannel      string
	Expiration      string
	DisableClaiming bool
	ChannelPrefix   string
}

type Claim struct {
	MessageID   string `gorm:"primaryKey"`
	ChannelID   string
	GuildID     string
	Item        string
	Type        string
	Winner      string
	Cost        float64
	Host        string
	BidHistory  string
	Note        string
	ImageURL    string
	TicketID    string
	Description string
}

type Giveaway struct {
	MessageID   string `gorm:"primaryKey"`
	ChannelID   string
	GuildID     string
	Item        string
	EndTime     time.Time
	Description string
	Host        string
	Winners     float64
	ImageURL    string
	Finished    bool
}

type GiveawaySetup struct {
	GuildID    string `gorm:"primaryKey"`
	HostRole   string
	AlertRole  string
	LogChannel string
}

type ShopSetup struct {
	GuildID    string `gorm:"primaryKey"`
	HostRole   string
	AlertRole  string
	LogChannel string
}

type Currency struct {
	Symbol      string
	RightSide   bool // True if the currency should display on the right side
	UseCurrency bool // True if user's currency is used for transaction.
}

type CurrencySetup struct {
	GuildID  string `gorm:"primaryKey"`
	Currency string
	Side     string
}

type UserProfile struct {
	UserID  string `gorm:"primaryKey;autoIncrement:false"`
	GuildID string `gorm:"primaryKey;autoIncrement:false"`
	Balance float64
}

type Quest struct {
	MessageID string `gorm:"primaryKey;autoIncrement:false"`
}

type Errors struct {
	ErrorID string `gorm:"primaryKey"`
	UserID  string `gorm:"primaryKey;autoIncrement:false"`
}

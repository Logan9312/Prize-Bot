package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func ParseSlashCommand(i *discordgo.InteractionCreate) map[string]interface{} {
	return ParseCommand(i, i.ApplicationCommandData().Options)

}

func ParseSubCommand(i *discordgo.InteractionCreate) map[string]interface{} {
	return ParseCommand(i, i.ApplicationCommandData().Options[0].Options)
}

func ParseSubSubCommand(i *discordgo.InteractionCreate) map[string]interface{} {
	return ParseCommand(i, i.ApplicationCommandData().Options[0].Options[0].Options)
}

func ParseCommand(i *discordgo.InteractionCreate, data []*discordgo.ApplicationCommandInteractionDataOption) map[string]any {
	options := make(map[string]interface{})
	for _, option := range data {
		options[option.Name] = option.Value
	}
	if options["image"] != nil {
		options["image"] = ImageToURL(i, options)
	}
	return options
}

func ParseTime(inputDuration string) (time.Duration, error) {
	if strings.HasSuffix(strings.ToLower(inputDuration), "d") {
		inputDuration = strings.TrimSuffix(strings.ToLower(inputDuration), "d")
		float, err := strconv.ParseFloat(inputDuration, 64)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		inputDuration = fmt.Sprint(float*24) + "h"
	}

	return time.ParseDuration(strings.ToLower(inputDuration))
}

func Ptr[T any](v T) *T {
	return &v
}

func ImageToURL(i *discordgo.InteractionCreate, image string) *string {
	return &i.ApplicationCommandData().Resolved.Attachments[image].URL
}

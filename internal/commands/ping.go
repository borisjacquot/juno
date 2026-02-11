package commands

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type PingCommand struct {
	logger *log.Logger
}

func NewPingCommand(logger *log.Logger) *PingCommand {
	return &PingCommand{
		logger: logger,
	}
}

func (c *PingCommand) Name() string {
	return "ping"
}

func (c *PingCommand) Description() string {
	return "Responds with pong"
}

func (c *PingCommand) Category() string {
	return "General"
}

func (c *PingCommand) ExecuteSlash(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	c.logger.WithField("user", i.Member.User.Username).Debug("Slash command ping executed")

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🏓 Pong!",
		},
	})
}

func (c *PingCommand) ToApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: c.Description(),
	}
}

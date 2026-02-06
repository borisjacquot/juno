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

func (c *PingCommand) Usage() string {
	return "!ping"
}

func (c *PingCommand) Examples() []string {
	return []string{"!ping"}
}

func (c *PingCommand) Category() string {
	return "General"
}

func (c *PingCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	c.logger.WithField("user", m.Author.Username).Debug("Received ping command")

	// react with a ping emoji
	err := s.MessageReactionAdd(m.ChannelID, m.ID, "🏓")
	if err != nil {
		c.logger.WithError(err).Error("Failed to add reaction to ping command")
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "🏓 Pong!")
	return err
}

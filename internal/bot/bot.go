package bot

import (
	"strings"

	"github.com/borisjacquot/juno/internal/commands"
	"github.com/borisjacquot/juno/internal/overwatch"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	session    *discordgo.Session
	owClient   *overwatch.Client
	cmdHandler *commands.Handler
	logger     *log.Logger
}

// NewBot creates a new Bot instance
func NewBot(token, overfastURL string, logger *log.Logger) (*Bot, error) {
	logger.Debug("Creating Discord session...")

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	logger.WithField("url", overfastURL).Debug("Creating Overwatch client...")
	owClient := overwatch.NewClient(overfastURL, logger)

	cmdHandler := commands.NewHandler(owClient, logger)

	bot := &Bot{
		session:    session,
		owClient:   owClient,
		cmdHandler: cmdHandler,
		logger:     logger,
	}

	// record handlers
	session.AddHandler(bot.messageCreate)
	session.AddHandler(bot.ready)

	// set intents
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	logger.Debug("Bot instance created successfully")
	return bot, nil
}

// Start opens the Discord session and starts the bot
func (b *Bot) Start() error {
	b.logger.Info("Opening WebSocket connection to Discord...")
	return b.session.Open()
}

// Stop closes the Discord session and stops the bot
func (b *Bot) Stop() error {
	b.logger.Info("Closing WebSocket connection to Discord...")
	return b.session.Close()
}

// ready is called when the bot has connected to Discord and is ready to receive events
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	b.logger.WithFields(log.Fields{
		"username": s.State.User.Username,
		"id":       s.State.User.ID,
		"guilds":   len(event.Guilds),
	}).Info("Bot is ready")

	// set the bot's presence
	err := s.UpdateGameStatus(0, "Overwatch | !help")
	if err != nil {
		b.logger.WithError(err).Error("Failed to set bot presence")
	}
}

// messageCreate is called when a new message is created in a channel the bot has access to
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// ignore messages that don't start with the command prefix
	if len(m.Content) == 0 || m.Content[0] != '!' {
		return
	}

	// parse command and arguments
	parts := strings.Fields(m.Content)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	b.logger.WithFields(log.Fields{
		"user":    m.Author.Username,
		"command": command,
		"args":    args,
		"channel": m.ChannelID,
		"guild":   m.GuildID,
	}).Debug("Received command")

	b.cmdHandler.Handle(s, m, command, args)
}

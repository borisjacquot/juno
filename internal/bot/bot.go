package bot

import (
	"github.com/borisjacquot/juno/internal/commands"
	"github.com/borisjacquot/juno/internal/database"
	"github.com/borisjacquot/juno/internal/overwatch"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	session    *discordgo.Session
	owClient   *overwatch.Client
	db         *database.Database
	cmdHandler *commands.Handler
	logger     *log.Logger
}

// NewBot creates a new Bot instance
func NewBot(token, overfastURL string, db *database.Database, logger *log.Logger) (*Bot, error) {
	logger.Debug("Creating Discord session...")

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	logger.WithField("url", overfastURL).Debug("Creating Overwatch client...")
	owClient := overwatch.NewClient(overfastURL, logger)

	cmdHandler := commands.NewHandler(owClient, db, logger)

	bot := &Bot{
		session:    session,
		owClient:   owClient,
		db:         db,
		cmdHandler: cmdHandler,
		logger:     logger,
	}

	// record handlers
	session.AddHandler(bot.interactionCreate)
	session.AddHandler(bot.ready)

	// set intents
	session.Identify.Intents = discordgo.IntentsGuilds

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
	b.logger.Info("Deleting slash commands...")
	registeredCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, "")
	if err != nil {
		b.logger.WithError(err).Error("Failed to fetch registered commands")
	} else {
		for _, cmd := range registeredCommands {
			err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", cmd.ID)
			if err != nil {
				b.logger.WithError(err).WithField("command", cmd.Name).Error("Failed to delete command")
			} else {
				b.logger.WithField("command", cmd.Name).Debug("Deleted command")
			}
		}
	}

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

	// register slash commands
	if err := b.cmdHandler.RegisterSlashCommands(s); err != nil {
		b.logger.WithError(err).Error("Failed to register slash commands")
		return
	}

	// set the bot's presence
	err := s.UpdateGameStatus(0, "Overwatch | /help")
	if err != nil {
		b.logger.WithError(err).Error("Failed to set bot presence")
	}
}

// interactionCreate is called when a new interaction is created (e.g. a slash command is used)
func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	b.logger.WithFields(log.Fields{
		"user":    i.Member.User.Username,
		"command": i.ApplicationCommandData().Name,
		"options": i.ApplicationCommandData().Options,
		"channel": i.ChannelID,
		"guild":   i.GuildID,
	}).Debug("Received interaction")

	b.cmdHandler.HandleSlashCommand(s, i)
}

package commands

import (
	"fmt"

	owcommands "github.com/borisjacquot/juno/internal/commands/overwatch"
	"github.com/borisjacquot/juno/internal/database"
	"github.com/borisjacquot/juno/internal/overwatch"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	registry *Registry
	logger   *log.Logger
}

// NewHandler creates a new command handler
func NewHandler(owClient *overwatch.Client, db *database.Database, logger *log.Logger) *Handler {
	registry := NewRegistry(logger)

	// Register general commands
	pingCmd := NewPingCommand(logger)
	if err := registry.Register(pingCmd); err != nil {
		logger.WithError(err).Error("Failed to register ping command")
	}
	registerCmd := NewRegisterCommand(db, logger)
	if err := registry.Register(registerCmd); err != nil {
		logger.WithError(err).Error("Failed to register register command")
	}

	// Register Overwatch commands
	profileCmd := owcommands.NewProfileCommand(owClient, db, logger)
	if err := registry.Register(profileCmd); err != nil {
		logger.WithError(err).Error("Failed to register profile command")
	}

	// Register help command
	helpCmd := NewHelpCommand(registry, logger)
	if err := registry.Register(helpCmd); err != nil {
		logger.WithError(err).Error("Failed to register help command")
	}

	logger.WithField("count", len(registry.commands)).Info("Registered commands")
	for name, cmd := range registry.commands {
		logger.WithFields(log.Fields{
			"name":     name,
			"category": cmd.Category(),
		}).Debug("Command available")
	}

	return &Handler{
		registry: registry,
		logger:   logger,
	}
}

// RegisterSlashCommands registers all commands as slash commands with Discord
func (h *Handler) RegisterSlashCommands(s *discordgo.Session) error {
	commands := h.registry.All()

	for _, cmd := range commands {
		slashCmd := cmd.ToApplicationCommand()

		h.logger.WithField("command", slashCmd.Name).Debug("Registering slash command")

		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", slashCmd)
		if err != nil {
			h.logger.WithError(err).WithField("command", slashCmd.Name).Error("Failed to register slash command")
			return err
		}
	}

	h.logger.WithField("count", len(commands)).Info("Registered slash commands")
	return nil
}

// HandleSlashCommand handles incoming slash command interactions
func (h *Handler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name

	cmd, ok := h.registry.Get(cmdName)
	if !ok {
		h.logger.WithField("command", cmdName).Debug("Slash command not found")
		respondWithError(s, i, "Unknown command.")
		return
	}

	h.logger.WithFields(log.Fields{
		"user":    i.Member.User.Username,
		"command": cmdName,
		"options": i.ApplicationCommandData().Options,
	}).Info("Executing slash command")

	if err := cmd.ExecuteSlash(s, i); err != nil {
		h.logger.WithError(err).WithField("command", cmdName).Error("Error executing slash command")
		respondWithError(s, i, fmt.Sprintf("Error executing command: %v", err))
	}
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

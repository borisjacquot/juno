package commands

import (
	"strings"

	"github.com/borisjacquot/juno/internal/overwatch"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	registry *Registry
	logger   *log.Logger
}

// NewHandler creates a new command handler
func NewHandler(owClient *overwatch.Client, logger *log.Logger) *Handler {
	registry := NewRegistry(logger)

	// Register general commands
	pingCmd := NewPingCommand(logger)
	if err := registry.Register(pingCmd); err != nil {
		logger.WithError(err).Error("Failed to register ping command")
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

// Handle handles commands
func (h *Handler) Handle(s *discordgo.Session, m *discordgo.MessageCreate, command string, args []string) {
	// remove prefix
	cmdName := strings.TrimPrefix(command, "!")

	h.registry.Execute(s, m, cmdName, args)
}

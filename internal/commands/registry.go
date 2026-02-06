package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// Command represents a bot command
type Command interface {
	// Name returns the name of the command (e.g. "ping")
	Name() string

	// Description returns a short description of the command (e.g. "Responds with pong")
	Description() string

	// Usage returns a string describing how to use the command (e.g. "!ping <args>")
	Usage() string

	// Examples returns a list of example usages of the command (e.g. ["!ping", "!ping arg1 arg2"])
	Examples() []string

	// Category returns the category of the command (e.g. "General", "Overwatch", etc.)
	Category() string

	// Execute executes the command with the given arguments and context
	Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error
}

// Registry is a registry of commands that can be executed by the bot
type Registry struct {
	commands map[string]Command
	logger   *log.Logger
}

// NewRegistry creates a new command registry
func NewRegistry(logger *log.Logger) *Registry {
	return &Registry{
		commands: make(map[string]Command),
		logger:   logger,
	}
}

// Register registers a command in the registry
func (r *Registry) Register(cmd Command) error {
	name := strings.ToLower(cmd.Name())

	if _, exists := r.commands[name]; exists {
		return fmt.Errorf("command with name '%s' already exists", name)
	}

	r.commands[name] = cmd
	r.logger.WithField("command", name).Debug("Registered command")

	return nil
}

// MustRegister registers a command and panics if there is an error
func (r *Registry) MustRegister(cmd Command) {
	if err := r.Register(cmd); err != nil {
		panic(fmt.Sprintf("failed to register command '%s': %v", cmd.Name(), err))
	}
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, bool) {
	cmd, exists := r.commands[strings.ToLower(name)]
	return cmd, exists
}

// All returns a slice of all registered commands
func (r *Registry) All() []Command {
	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	// sort commands by categories and name
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Category() == commands[j].Category() {
			return commands[i].Name() < commands[j].Name()
		}
		return commands[i].Category() < commands[j].Category()
	})
	return commands
}

// Execute executes a command
func (r *Registry) Execute(s *discordgo.Session, m *discordgo.MessageCreate, cmdName string, args []string) {
	cmd, ok := r.Get(cmdName)
	if !ok {
		r.logger.WithField("command", cmdName).Debug("Command not found")
		s.ChannelMessageSend(m.ChannelID, "❌ Unknown command. Use `!help` to see the list of available commands.")
		return
	}

	r.logger.WithFields(log.Fields{
		"user":    m.Author.Username,
		"command": cmdName,
		"args":    args,
	}).Info("Executing command")

	if err := cmd.Execute(s, m, args); err != nil {
		r.logger.WithError(err).WithField("command", cmdName).Error("Error executing command")
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error executing command: %v", err))
	}
}

// GetByCategory returns a map of commands grouped by category
func (r *Registry) GetByCategory() map[string][]Command {
	categories := make(map[string][]Command)

	for _, cmd := range r.commands {
		category := cmd.Category()
		categories[category] = append(categories[category], cmd)
	}

	// sort commands in each category by name
	for category := range categories {
		sort.Slice(categories[category], func(i, j int) bool {
			return categories[category][i].Name() < categories[category][j].Name()
		})
	}

	return categories
}

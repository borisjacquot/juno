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

	// Category returns the category of the command (e.g. "General", "Overwatch", etc.)
	Category() string

	// ExecuteSlash executes the command as a slash command with the given arguments and context
	ExecuteSlash(s *discordgo.Session, i *discordgo.InteractionCreate) error

	// ToApplicationCommand converts the command to a Discord application command (for slash commands)
	ToApplicationCommand() *discordgo.ApplicationCommand
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
	r.logger.WithFields(log.Fields{
		"name":     cmd.Name(),
		"category": cmd.Category(),
	}).Debug("Registered command")

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

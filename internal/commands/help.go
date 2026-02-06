package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type HelpCommand struct {
	registry *Registry
	logger   *log.Logger
}

func NewHelpCommand(registry *Registry, logger *log.Logger) *HelpCommand {
	return &HelpCommand{
		registry: registry,
		logger:   logger,
	}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "Provides information about available commands"
}

func (c *HelpCommand) Usage() string {
	return "!help [command]"
}

func (c *HelpCommand) Examples() []string {
	return []string{"!help", "!help ping"}
}

func (c *HelpCommand) Category() string {
	return "General"
}

func (c *HelpCommand) Execute(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	c.logger.WithFields(log.Fields{
		"user": m.Author.Username,
		"args": args,
	}).Debug("Received help command")

	// if an argument is provided, show detailed help for that command
	if len(args) > 0 {
		return c.showDetailedHelp(s, m, args[0])
	}

	// otherwise, show a list of all commands
	return c.showGeneralHelp(s, m)
}

// showGeneralHelp sends a message with a list of all commands and their descriptions
func (c *HelpCommand) showGeneralHelp(s *discordgo.Session, m *discordgo.MessageCreate) error {
	categories := make(map[string][]Command)

	// sort categories
	categoryNames := make([]string, 0, len(categories))
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)

	embed := &discordgo.MessageEmbed{
		Title:       "📖 Help - Juno Bot",
		Description: "Here's a list of all available commands. Use `!help [command]` to get detailed information.",
		Color:       0xf06c9b,
		Fields:      []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Total: %d commands | Juno Bot", len(c.registry.commands)),
			IconURL: "https://raw.githubusercontent.com/borisjacquot/juno/main/assets/img/icon.png",
		},
	}

	// group commands by category
	for _, category := range categoryNames {
		commands := categories[category]

		var commandList strings.Builder
		for _, cmd := range commands {
			commandList.WriteString(fmt.Sprintf("**!%s** - %s\n", cmd.Name(), cmd.Description()))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("📂 %s", category),
			Value:  commandList.String(),
			Inline: false,
		})
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

// showDetailedHelp sends a message with detailed information about a specific command
func (c *HelpCommand) showDetailedHelp(s *discordgo.Session, m *discordgo.MessageCreate, cmdName string) error {
	cmd, exists := c.registry.Get(cmdName)
	if !exists {
		return fmt.Errorf("command '%s' not found", cmdName)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📖 Help - !%s Command", cmd.Name()),
		Description: cmd.Description(),
		Color:       0xf06c9b,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📝 Usage",
				Value:  fmt.Sprintf("`%s`", cmd.Usage()),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Category: %s | Juno Bot", cmd.Category()),
			IconURL: "https://raw.githubusercontent.com/borisjacquot/juno/main/assets/img/icon.png",
		},
	}

	// add example if available
	if examples := cmd.Examples(); len(examples) > 0 {
		var exampleList strings.Builder
		for _, example := range examples {
			exampleList.WriteString(fmt.Sprintf("`%s`\n", example))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "💡 Examples",
			Value:  exampleList.String(),
			Inline: false,
		})
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

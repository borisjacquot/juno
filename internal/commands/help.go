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

func (c *HelpCommand) Category() string {
	return "General"
}

func (c *HelpCommand) ExecuteSlash(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options

	// instant response to acknowledge the command
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return err
	}

	// if an argument is provided, show detailed help for that command
	if len(options) > 0 {
		cmdName := options[0].StringValue()
		return c.showDetailedHelp(s, i.ChannelID, cmdName, i.Interaction)
	}

	// otherwise, show a list of all commands
	return c.showGeneralHelp(s, i.ChannelID, i.Interaction)
}

// showGeneralHelp sends a message with a list of all commands and their descriptions
func (c *HelpCommand) showGeneralHelp(s *discordgo.Session, channelID string, interaction *discordgo.Interaction) error {
	categories := c.registry.GetByCategory()

	categoryNames := make([]string, 0, len(categories))
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)

	embed := &discordgo.MessageEmbed{
		Title:       "📖 Help - Juno Bot",
		Description: "Here's a list of all available commands. Use `/help command:<name>` for detailed information about a specific command.",
		Color:       0xD183C9,
		Fields:      []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Total: %d commands | Juno Bot", len(c.registry.commands)),
			IconURL: "https://raw.githubusercontent.com/borisjacquot/juno/main/assets/img/icon.png",
		},
	}

	for _, category := range categoryNames {
		commands := categories[category]

		var commandList strings.Builder
		for _, cmd := range commands {
			commandList.WriteString(fmt.Sprintf("**/%s** - %s\n", cmd.Name(), cmd.Description()))
		}

		fieldValue := commandList.String()
		if fieldValue == "" {
			fieldValue = "No commands in this category"
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("📌 %s", category),
			Value:  fieldValue,
			Inline: false,
		})
	}

	if interaction != nil {
		_, err := s.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
		return err
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// showDetailedHelp sends a message with detailed information about a specific command
func (c *HelpCommand) showDetailedHelp(s *discordgo.Session, channelID, cmdName string, interaction *discordgo.Interaction) error {
	cmd, ok := c.registry.Get(cmdName)
	if !ok {
		errMsg := fmt.Sprintf("❌ Command '%s' not found", cmdName)

		if interaction != nil {
			_, err := s.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
				Content: &errMsg,
			})
			return err
		}

		_, err := s.ChannelMessageSend(channelID, errMsg)
		return err
	}

	// get the application command data for the command to show usage and options
	appCmd := cmd.ToApplicationCommand()

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📖 Help - /%s", cmd.Name()),
		Description: cmd.Description(),
		Color:       0xD183C9,
		Fields:      []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Category: %s", cmd.Category()),
			IconURL: "https://raw.githubusercontent.com/borisjacquot/juno/main/assets/img/icon.png",
		},
	}

	if len(appCmd.Options) > 0 {
		var optionsList strings.Builder
		for _, opt := range appCmd.Options {
			required := ""
			if opt.Required {
				required = " *(required)*"
			}
			optionsList.WriteString(fmt.Sprintf("• **%s**%s - %s\n", opt.Name, required, opt.Description))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚙️ Options",
			Value:  optionsList.String(),
			Inline: false,
		})
	}

	if interaction != nil {
		_, err := s.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
		return err
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

func (c *HelpCommand) ToApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: c.Description(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "command",
				Description: "The name of the command to get detailed help for",
				Required:    false,
			},
		},
	}
}

package commands

import (
	"fmt"
	"regexp"

	"github.com/borisjacquot/juno/internal/database"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var battleTagRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,12}#[0-9]{4,5}$`)

type RegisterCommand struct {
	db     *database.Database
	logger *log.Logger
}

func NewRegisterCommand(db *database.Database, logger *log.Logger) *RegisterCommand {
	return &RegisterCommand{
		db:     db,
		logger: logger,
	}
}

func (c *RegisterCommand) Name() string {
	return "register"
}

func (c *RegisterCommand) Description() string {
	return "Register your Overwatch BattleTag to link it with your Discord account"
}

func (c *RegisterCommand) Category() string {
	return "General"
}

func (c *RegisterCommand) ExecuteSlash(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options

	if len(options) == 0 {
		return fmt.Errorf("you must provide your BattleTag as an argument (e.g. /register battleTag:Player#1234)")
	}

	battleTag := options[0].StringValue()

	// validate BattleTag format
	if !c.isValidBattleTag(battleTag) {
		return c.respondError(s, i, "❌ Invalid BattleTag format. It should be in the format `Player#1234`")
	}

	// convert to overwatch api format (Player-1234)
	battleTagForAPI := c.convertBattleTagFormat(battleTag)

	c.logger.WithFields(log.Fields{
		"user":      i.Member.User.Username,
		"battleTag": battleTag,
		"guild_id":  i.GuildID,
	}).Info("Registering BattleTag for user")

	// answer immediately to acknowledge the command
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}

	// save the BattleTag in the database
	err = c.db.RegisterUser(i.GuildID, i.Member.User.ID, battleTagForAPI)
	if err != nil {
		c.logger.WithError(err).Error("Failed to register BattleTag in database")
		return c.editResponse(s, i, "❌ Failed to register your BattleTag. Please try again later.")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ BattleTag Registered",
		Description: fmt.Sprintf("Your Discord account has been linked to `%s`!", battleTag),
		Color:       0xD183C9,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📝 BattleTag",
				Value:  battleTag,
				Inline: true,
			},
			{
				Name:   "👤 Discord User",
				Value:  i.Member.User.Username,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /profile to view your Overwatch stats",
		},
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})

	return err
}

func (c *RegisterCommand) isValidBattleTag(battleTag string) bool {
	return battleTagRegex.MatchString(battleTag)
}

func (c *RegisterCommand) convertBattleTagFormat(battleTag string) string {
	// Convertir Pseudo#1234 en Pseudo-1234 pour l'API Overwatch
	return regexp.MustCompile(`#`).ReplaceAllString(battleTag, "-")
}

func (c *RegisterCommand) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *RegisterCommand) editResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr(message),
	})
	return err
}

func stringPtr(s string) *string {
	return &s
}

func (c *RegisterCommand) ToApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: c.Description(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "battletag",
				Description: "Your Overwatch BattleTag (e.g. Player#1234)",
				Required:    true,
			},
		},
	}
}

package overwatch

import (
	"fmt"
	"strings"

	"github.com/borisjacquot/juno/internal/database"
	"github.com/borisjacquot/juno/internal/overwatch"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type ProfileCommand struct {
	owClient *overwatch.Client
	db       *database.Database
	logger   *log.Logger
}

func NewProfileCommand(owClient *overwatch.Client, db *database.Database, logger *log.Logger) *ProfileCommand {
	return &ProfileCommand{
		owClient: owClient,
		db:       db,
		logger:   logger,
	}
}

func (c *ProfileCommand) Name() string {
	return "profile"
}

func (c *ProfileCommand) Description() string {
	return "Display the Overwatch profile of a user"
}

func (c *ProfileCommand) Category() string {
	return "Overwatch"
}

func (c *ProfileCommand) ExecuteSlash(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// answer immediately
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return err
	}

	// get the target
	var targetUser *discordgo.User
	options := i.ApplicationCommandData().Options

	if len(options) > 0 && options[0].UserValue(s) != nil {
		targetUser = options[0].UserValue(s)
	} else {
		targetUser = i.Member.User
	}

	c.logger.WithFields(log.Fields{
		"requester": i.Member.User.Username,
		"target":    targetUser.Username,
		"guild_id":  i.GuildID,
	}).Info("Fetching profile for user")

	// search for the user's BattleTag in the database
	battleTag, err := c.db.GetUserBattleTag(i.GuildID, targetUser.ID)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get BattleTag from database")
		return c.editResponse(s, i, "❌ Failed to retrieve BattleTag.")
	}

	if battleTag == "" {
		username := targetUser.Username
		if targetUser.ID == i.Member.User.ID {
			return c.editResponse(s, i, "❌ You haven't registered your BattleTag yet. Use `/register` to link your Overwatch account.")
		}
		return c.editResponse(s, i, fmt.Sprintf("❌ **%s** hasn't registered their BattleTag yet.", username))
	}

	// get ow stats from the API
	player, err := c.owClient.GetPlayer(battleTag)
	if err != nil {
		c.logger.WithError(err).Error("Failed to fetch player profile from Overwatch API")
		return c.editResponse(s, i, "❌ Failed to fetch player profile. Please try again later.")
	}

	c.logger.WithFields(log.Fields{
		"battletag": battleTag,
		"name":      player.Name,
	}).Debug("Successfully fetched player profile from Overwatch API")

	embed := c.buildProfileEmbed(player, targetUser, battleTag)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})

	return err
}

func (c *ProfileCommand) buildProfileEmbed(player *overwatch.Player, discordUser *discordgo.User, battleTag string) *discordgo.MessageEmbed {
	displayBattleTag := convertToDisplayFormat(battleTag)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 Overwatch Profile - %s", player.Name),
		Description: fmt.Sprintf("<@%s>'s profile", discordUser.ID),
		Color:       getRankColor(player),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: player.Avatar,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎮 BattleTag",
				Value:  displayBattleTag,
				Inline: true,
			},
			{
				Name:   "⭐ Endorsement Level",
				Value:  fmt.Sprintf("%d", player.Endorsement.Level),
				Inline: true,
			},
			{
				Name:   "👤 Discord",
				Value:  discordUser.Username,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Data provided by Overfast API",
		},
	}

	// add namecard image if available
	if player.NameCard != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: player.NameCard,
		}
	}

	// add PC competitive ranks
	if player.Competitive.PC.Season > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🏆 Competitive Season",
			Value:  fmt.Sprintf("Season %d", player.Competitive.PC.Season),
			Inline: false,
		})

		// Tank
		if player.Competitive.PC.Tank.Division != "" {
			tankRank := formatRank(player.Competitive.PC.Tank)
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🛡️ Tank",
				Value:  tankRank,
				Inline: true,
			})
		}

		// Damage
		if player.Competitive.PC.Damage.Division != "" {
			damageRank := formatRank(player.Competitive.PC.Damage)
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "⚔️ Damage",
				Value:  damageRank,
				Inline: true,
			})
		}

		// Support
		if player.Competitive.PC.Support.Division != "" {
			supportRank := formatRank(player.Competitive.PC.Support)
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "💚 Support",
				Value:  supportRank,
				Inline: true,
			})
		}

		// Open Queue (si disponible)
		if player.Competitive.PC.Open.Division != "" {
			openRank := formatRank(player.Competitive.PC.Open)
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🌐 Open Queue",
				Value:  openRank,
				Inline: true,
			})
		}
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "ℹ️ Competitive Stats",
			Value:  "No competitive data available",
			Inline: false,
		})
	}

	return embed
}

// formatRank formats the competitive rank information into a readable string
func formatRank(rank overwatch.CompetitiveStatsRole) string {
	if rank.Division == "" {
		return "Unranked"
	}

	// map divisions to emojis
	divisionEmoji := getDivisionEmoji(rank.Division)

	return fmt.Sprintf("%s **%s %d**", divisionEmoji, rank.Division, rank.Tier)
}

// getDivisionEmoji returns an emoji corresponding to the competitive division
func getDivisionEmoji(division string) string {
	division = strings.ToLower(division)

	switch {
	case strings.Contains(division, "bronze"):
		return "🥉"
	case strings.Contains(division, "silver"):
		return "🥈"
	case strings.Contains(division, "gold"):
		return "🥇"
	case strings.Contains(division, "platinum"):
		return "💎"
	case strings.Contains(division, "diamond"):
		return "💠"
	case strings.Contains(division, "master"):
		return "👑"
	case strings.Contains(division, "grandmaster"):
		return "🏆"
	case strings.Contains(division, "champion"):
		return "⭐"
	default:
		return "🎮"
	}
}

// getRankColor returns a color code based on the player's highest competitive rank
func getRankColor(player *overwatch.Player) int {
	highestRank := getHighestRank(player)

	switch {
	case strings.Contains(strings.ToLower(highestRank), "champion"):
		return 0xFF6B9D
	case strings.Contains(strings.ToLower(highestRank), "grandmaster"):
		return 0xFFB900
	case strings.Contains(strings.ToLower(highestRank), "master"):
		return 0xFF9900
	case strings.Contains(strings.ToLower(highestRank), "diamond"):
		return 0x6B48FF
	case strings.Contains(strings.ToLower(highestRank), "platinum"):
		return 0x00D4D4
	case strings.Contains(strings.ToLower(highestRank), "gold"):
		return 0xFFB900
	case strings.Contains(strings.ToLower(highestRank), "silver"):
		return 0xC0C0C0
	case strings.Contains(strings.ToLower(highestRank), "bronze"):
		return 0xCD7F32
	default:
		return 0xF99E1A
	}
}

// getHighestRank determines the player's highest competitive rank across all roles
func getHighestRank(player *overwatch.Player) string {
	ranks := []string{
		player.Competitive.PC.Tank.Division,
		player.Competitive.PC.Damage.Division,
		player.Competitive.PC.Support.Division,
		player.Competitive.PC.Open.Division,
	}

	rankOrder := map[string]int{
		"champion":    8,
		"grandmaster": 7,
		"master":      6,
		"diamond":     5,
		"platinum":    4,
		"gold":        3,
		"silver":      2,
		"bronze":      1,
	}

	highestRank := ""
	highestValue := 0

	for _, rank := range ranks {
		if rank == "" {
			continue
		}

		rankLower := strings.ToLower(rank)
		for key, value := range rankOrder {
			if strings.Contains(rankLower, key) && value > highestValue {
				highestValue = value
				highestRank = rank
			}
		}
	}

	return highestRank
}

func (c *ProfileCommand) editResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr(message),
	})
	return err
}

func (c *ProfileCommand) ToApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: c.Description(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The Discord user whose profile you want to see (yourself by default)",
				Required:    false,
			},
		},
	}
}

// convertToDisplayFormat converts a BattleTag from "Pseudo-1234" to "Pseudo#1234"
func convertToDisplayFormat(battleTag string) string {
	// Convertir Pseudo-1234 en Pseudo#1234 pour l'affichage
	if len(battleTag) > 0 {
		for i := len(battleTag) - 1; i >= 0; i-- {
			if battleTag[i] == '-' {
				return battleTag[:i] + "#" + battleTag[i+1:]
			}
		}
	}
	return battleTag
}

func stringPtr(s string) *string {
	return &s
}

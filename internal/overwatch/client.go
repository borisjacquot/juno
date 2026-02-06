package overwatch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Client is a client for the Overwatch API
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
}

// Endorsement represents a player's endorsement level
type Endorsement struct {
	Level int    `json:"level"`
	Frame string `json:"frame"`
}

// CompetitiveStatsRole represents a player's competitive stats per role
type CompetitiveStatsRole struct {
	Division string `json:"division"`
	Tier     int    `json:"tier"`
	RoleIcon string `json:"role_icon"`
	RankIcon string `json:"rank_icon"`
	TierIcon string `json:"tier_icon"`
}

// CompetitivePlatformStats represents a player's competitive stats for a specific platform
type CompetitivePlatformStats struct {
	Season  int                  `json:"season"`
	Tank    CompetitiveStatsRole `json:"tank"`
	Damage  CompetitiveStatsRole `json:"damage"`
	Support CompetitiveStatsRole `json:"support"`
	Open    CompetitiveStatsRole `json:"open"`
}

// Competitive represents a player's competitive stats
type Competitive struct {
	PC      CompetitivePlatformStats `json:"pc"`
	Console CompetitivePlatformStats `json:"console"`
}

// Player represents an Overwatch player
type Player struct {
	Name          string      `json:"username"`
	Avatar        string      `json:"avatar"`
	NameCard      string      `json:"namecard"`
	Endorsement   Endorsement `json:"endorsement"`
	Competitive   Competitive `json:"competitive"`
	LastUpdatedAt int         `json:"last_updated_at"`
}

// NewClient creates a new Overwatch API client
func NewClient(baseURL string, logger *log.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetPlayer retrieves a player's profile from the Overwatch API
func (c *Client) GetPlayer(battletag string) (*Player, error) {
	url := fmt.Sprintf("%s/players/%s/summary", c.baseURL, battletag)

	c.logger.WithFields(log.Fields{
		"battletag": battletag,
		"url":       url,
	}).Info("Fetching player profile from Overwatch API")

	start := time.Now()
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.logger.WithError(err).WithField("url", url).Error("Failed to fetch player profile from Overwatch API")
		return nil, fmt.Errorf("failed to fetch player profile: %w", err)
	}
	defer resp.Body.Close()

	c.logger.WithFields(log.Fields{
		"battletag": battletag,
		"url":       url,
		"status":    resp.StatusCode,
		"duration":  time.Since(start),
	}).Debug("Fetched player profile from Overwatch API")

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(log.Fields{
			"battletag": battletag,
			"status":    resp.StatusCode,
		}).Warn("Player profile not found in Overwatch API")
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var player Player
	if err := json.NewDecoder(resp.Body).Decode(&player); err != nil {
		c.logger.WithError(err).WithField("battletag", battletag).Error("Failed to decode player profile from Overwatch API")
		return nil, fmt.Errorf("failed to decode player profile: %w", err)
	}

	c.logger.WithFields(log.Fields{
		"name":            player.Name,
		"last_updated_at": time.Unix(int64(player.LastUpdatedAt), 0).Format(time.RFC3339),
	}).Info("Successfully retrieved player profile from Overwatch API")

	return &player, nil
}

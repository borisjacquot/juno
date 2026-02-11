package database

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Database struct {
	db     *gorm.DB
	logger *log.Logger
}

// New creates a new Database instance
func New(dbPath string, lgr *log.Logger) (*Database, error) {
	lgr.WithField("path", dbPath).Info("Connecting to database...")

	// create directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// config logger for gorm
	gormLogger := logger.Default.LogMode(logger.Silent)

	// see db logs if log level is debug
	if lgr.Level == log.DebugLevel {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// auto migrate schemas
	lgr.Info("Migrating database models...")
	if err := db.AutoMigrate(&UserRegistration{}); err != nil {
		return nil, fmt.Errorf("error during migration: %w", err)
	}

	lgr.Info("Database connection established successfully")
	return &Database{
		db:     db,
		logger: lgr,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql DB from gorm: %w", err)
	}

	return sqlDB.Close()
}

// RegisterUser registers a user with their BattleTag in the database for a specific guild
func (d *Database) RegisterUser(guildID, userID, battleTag string) error {
	d.logger.WithFields(log.Fields{
		"guild_id":  guildID,
		"user_id":   userID,
		"battleTag": battleTag,
	}).Debug("Registering user in database")

	registration := UserRegistration{
		GuildID:   guildID,
		UserID:    userID,
		BattleTag: battleTag,
	}

	result := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"battle_tag", "updated_at"}),
	}).Create(&registration)

	if result.Error != nil {
		return fmt.Errorf("failed to register user: %w", result.Error)
	}

	d.logger.WithFields(log.Fields{
		"guild_id":  guildID,
		"user_id":   userID,
		"battleTag": battleTag,
	}).Info("User registered successfully")

	return nil
}

// GetUserBattleTag retrieves a user's BattleTag from the database for a specific guild
func (d *Database) GetUserBattleTag(guildID, userID string) (string, error) {
	d.logger.WithFields(log.Fields{
		"guild_id": guildID,
		"user_id":  userID,
	}).Debug("Retrieving user BattleTag from database")

	var registration UserRegistration
	result := d.db.Where(&UserRegistration{
		GuildID: guildID,
		UserID:  userID,
	}).First(&registration)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil // user not found, return empty string
		}
		return "", fmt.Errorf("failed to retrieve user BattleTag: %w", result.Error)
	}

	return registration.BattleTag, nil
}

// UnregisterUser removes a user's registration from the database for a specific guild
func (d *Database) UnregisterUser(guildID, userID string) error {
	d.logger.WithFields(log.Fields{
		"guild_id": guildID,
		"user_id":  userID,
	}).Debug("Unregistering user from database")

	result := d.db.Where(&UserRegistration{
		GuildID: guildID,
		UserID:  userID,
	}).Delete(&UserRegistration{})

	if result.Error != nil {
		return fmt.Errorf("failed to unregister user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // no record deleted, user not found
	}

	d.logger.WithFields(log.Fields{
		"guild_id": guildID,
		"user_id":  userID,
	}).Info("User unregistered successfully")

	return nil
}

// GetGuildRegistrations retrieves all user registrations for a specific guild
func (d *Database) GetGuildRegistrations(guildID string) ([]UserRegistration, error) {
	d.logger.WithField("guild_id", guildID).Debug("Retrieving all user registrations for guild")

	var registrations []UserRegistration
	result := d.db.Where(&UserRegistration{
		GuildID: guildID,
	}).Find(&registrations)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve guild registrations: %w", result.Error)
	}

	return registrations, nil
}

// GetUserStats returns number of registered users
func (d *Database) GetUserStats() (int64, error) {
	var count int64
	result := d.db.Model(&UserRegistration{}).Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to count user registrations: %w", result.Error)
	}

	return count, nil
}

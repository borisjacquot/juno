package database

import (
	"time"

	"gorm.io/gorm"
)

// UserRegistration represents a link between a Discord user and their BattleTag
type UserRegistration struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      // Timestamp of when the registration was created
	UpdatedAt time.Time      // Timestamp of when the registration was last updated
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete field

	// foreign keys
	GuildID string `gorm:"uniqueIndex:idx_user_guild;not null"` // Discord Guild ID
	UserID  string `gorm:"uniqueIndex:idx_user_guild;not null"` // Discord User ID

	// data
	BattleTag string `gorm:"not null"` // User's BattleTag (e.g. "Player#1234")
}

// TableName specifies the table name for UserRegistration
func (UserRegistration) TableName() string {
	return "user_registrations"
}

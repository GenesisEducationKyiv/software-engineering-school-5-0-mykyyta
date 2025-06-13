package model

import (
	"time"
)

// Subscription represents a user's weather subscription entry.
// It stores email, city, frequency, status flags, and metadata.
//
// Notes for production:
// - UUID is stored as a string instead of a native UUID type for compatibility (e.g. SQLite).
// - Frequency validation is handled in application logic (no DB-level CHECK constraint).
// - Token is not exposed in JSON (used for confirmation/unsubscribe).
type Subscription struct {
	ID             string    `gorm:"primaryKey" json:"id"`                // UUID stored as string for compatibility
	Email          string    `gorm:"not null;uniqueIndex" json:"email"`   // Unique per user
	City           string    `gorm:"not null" json:"city"`                // Target city for weather updates
	Frequency      string    `gorm:"type:text;not null" json:"frequency"` // "daily" or "hourly" — validated in code
	IsConfirmed    bool      `gorm:"default:false" json:"isConfirmed"`    // True if user confirmed via email
	IsUnsubscribed bool      `gorm:"default:false" json:"isUnsubscribed"` // True if user opted out
	Token          string    `gorm:"not null" json:"-"`                   // Used for confirmation & unsubscribe; hidden from API responses
	CreatedAt      time.Time `json:"createdAt"`                           // Timestamp of subscription
}

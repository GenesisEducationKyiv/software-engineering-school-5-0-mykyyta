package api

import "gorm.io/gorm"

// DB is a globally accessible database connection used across the API package.
// It is initialized once via SetDB and then reused in handlers and helpers.
var DB *gorm.DB

// SetDB initializes the global DB instance.
// This should be called once during application startup (e.g. in main.go).
func SetDB(db *gorm.DB) {
	DB = db
}

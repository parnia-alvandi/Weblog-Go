package models

// User represents a registered account.
// Passwords are always stored as bcrypt hashes, never plaintext.
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;size:50;not null"`
	Password string `gorm:"not null"` // bcrypt hash
}

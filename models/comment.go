package models

import "time"

// Comment holds only text plus the author's username, as required by the spec.
type Comment struct {
	ID      uint   `gorm:"primaryKey"`
	Content string `gorm:"type:text;not null"`

	PostID uint

	AuthorID uint
	Author   User `gorm:"foreignKey:AuthorID"`

	CreatedAt time.Time
}

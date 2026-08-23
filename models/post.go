package models

import "time"

// Post represents a "board" (blog post) as described in the spec:
// Title, Content, optional Image, Author, and Privacy Status.
type Post struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"size:200;not null"`
	Content   string `gorm:"type:text;not null"`
	ImagePath string // empty string means no image was attached
	IsPrivate bool   `gorm:"default:false"`

	AuthorID uint
	Author   User `gorm:"foreignKey:AuthorID"`

	// SharedWith holds the specific users a private post has been shared with.
	// Ignored entirely when IsPrivate is false.
	SharedWith []User `gorm:"many2many:post_shares;"`

	Comments []Comment `gorm:"foreignKey:PostID"`

	CreatedAt time.Time
}

// CanBeViewedBy encodes the access-control rule from the spec:
// - public posts: everyone (including guests) can view
// - private posts: only the author and users explicitly shared with
func (p *Post) CanBeViewedBy(userID uint, isLoggedIn bool) bool {
	if !p.IsPrivate {
		return true
	}
	if !isLoggedIn {
		return false
	}
	if p.AuthorID == userID {
		return true
	}
	for _, u := range p.SharedWith {
		if u.ID == userID {
			return true
		}
	}
	return false
}

package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"weblog/database"
	"weblog/middleware"
	"weblog/models"
)

type CommentHandler struct{}

func NewCommentHandler() *CommentHandler {
	return &CommentHandler{}
}

// Create adds a comment to a post, but only if the commenter is logged in
// AND has permission to view that post (same rule as viewing it).
func (h *CommentHandler) Create(c echo.Context) error {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	var post models.Post
	if err := database.DB.Preload("SharedWith").First(&post, postID).Error; err != nil {
		return c.String(http.StatusNotFound, "Post not found")
	}

	userID := middleware.CurrentUserID(c)
	if !post.CanBeViewedBy(userID, middleware.IsLoggedIn(c)) {
		return c.String(http.StatusForbidden, "You do not have permission to comment on this post")
	}

	content := strings.TrimSpace(c.FormValue("content"))
	if content == "" {
		return c.Redirect(http.StatusSeeOther, "/weblog/"+strconv.Itoa(postID))
	}

	comment := models.Comment{
		Content:   content,
		PostID:    post.ID,
		AuthorID:  userID,
		CreatedAt: time.Now(),
	}
	database.DB.Create(&comment)

	return c.Redirect(http.StatusSeeOther, "/weblog/"+strconv.Itoa(postID))
}

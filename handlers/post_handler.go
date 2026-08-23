package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"weblog/database"
	"weblog/middleware"
	"weblog/models"
)

type PostHandler struct {
	UploadDir string // e.g. "static/uploads"
}

func NewPostHandler(uploadDir string) *PostHandler {
	return &PostHandler{UploadDir: uploadDir}
}

// Home lists every post the current visitor is allowed to see:
// all public posts, plus private posts they authored or were shared with.
func (h *PostHandler) Home(c echo.Context) error {
	var allPosts []models.Post
	if err := database.DB.
		Preload("Author").
		Preload("SharedWith").
		Order("created_at desc").
		Find(&allPosts).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load posts")
	}

	loggedIn := middleware.IsLoggedIn(c)
	userID := middleware.CurrentUserID(c)

	visible := make([]models.Post, 0, len(allPosts))
	for _, p := range allPosts {
		if p.CanBeViewedBy(userID, loggedIn) {
			visible = append(visible, p)
		}
	}

	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"LoggedIn": loggedIn,
		"Username": c.Get("username"),
		"Posts":    visible,
	})
}

func (h *PostHandler) CreatePage(c echo.Context) error {
	return c.Render(http.StatusOK, "create.html", map[string]interface{}{
		"LoggedIn": true,
		"Username": c.Get("username"),
	})
}

// Create handles the new-post form, including optional image upload
// and (if private) a comma-separated list of usernames to share with.
func (h *PostHandler) Create(c echo.Context) error {
	userID := middleware.CurrentUserID(c)

	title := strings.TrimSpace(c.FormValue("title"))
	content := strings.TrimSpace(c.FormValue("content"))
	isPrivate := c.FormValue("privacy") == "private"
	sharedUsernamesRaw := c.FormValue("shared_with") // e.g. "alice, bob"

	if title == "" || content == "" {
		return c.Render(http.StatusBadRequest, "create.html", map[string]interface{}{
			"Error": "Title and content are required.",
		})
	}

	post := models.Post{
		Title:     title,
		Content:   content,
		IsPrivate: isPrivate,
		AuthorID:  userID,
		CreatedAt: time.Now(),
	}

	// Optional image upload
	if file, err := c.FormFile("image"); err == nil {
		src, err := file.Open()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to open image file")
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		newName := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
		destPath := filepath.Join(h.UploadDir, newName)

		dst, err := os.Create(destPath)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to save image")
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return c.String(http.StatusInternalServerError, "Failed to write image file")
		}

		post.ImagePath = "/uploads/" + newName
	}

	if err := database.DB.Create(&post).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to save post")
	}

	// If private, resolve the shared usernames into user records and link them.
	if isPrivate && sharedUsernamesRaw != "" {
		var sharedUsers []models.User
		names := strings.Split(sharedUsernamesRaw, ",")
		for i := range names {
			names[i] = strings.TrimSpace(names[i])
		}
		if err := database.DB.Where("username IN ?", names).Find(&sharedUsers).Error; err == nil {
			_ = database.DB.Model(&post).Association("SharedWith").Append(&sharedUsers)
		}
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/weblog/%d", post.ID))
}

// Detail shows a single post plus its comments, enforcing the same
// visibility rule as Home.
func (h *PostHandler) Detail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	var post models.Post
	if err := database.DB.
		Preload("Author").
		Preload("SharedWith").
		First(&post, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Post not found")
	}

	// Comments are loaded separately (with their authors, in chronological order)
	// and attached to the post afterwards.
	var comments []models.Comment
	database.DB.Preload("Author").Where("post_id = ?", post.ID).Order("created_at asc").Find(&comments)
	post.Comments = comments

	loggedIn := middleware.IsLoggedIn(c)
	userID := middleware.CurrentUserID(c)

	if !post.CanBeViewedBy(userID, loggedIn) {
		return c.String(http.StatusForbidden, "You do not have permission to view this post")
	}

	return c.Render(http.StatusOK, "detail.html", map[string]interface{}{
		"LoggedIn":  loggedIn,
		"Username":  c.Get("username"),
		"Post":      post,
		"IsOwner":   loggedIn && post.AuthorID == userID,
		"CurrentID": userID,
	})
}

// Delete removes a post, but only if the requester is its author.
func (h *PostHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	var post models.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Post not found")
	}

	userID := middleware.CurrentUserID(c)
	if post.AuthorID != userID {
		return c.String(http.StatusForbidden, "Only the author can delete this post")
	}

	database.DB.Where("post_id = ?", post.ID).Delete(&models.Comment{})
	database.DB.Model(&post).Association("SharedWith").Clear()
	database.DB.Delete(&post)

	return c.Redirect(http.StatusSeeOther, "/")
}

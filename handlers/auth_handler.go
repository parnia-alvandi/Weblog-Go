package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"weblog/database"
	"weblog/middleware"
	"weblog/models"
)

type AuthHandler struct {
	Store sessions.Store
}

func NewAuthHandler(store sessions.Store) *AuthHandler {
	return &AuthHandler{Store: store}
}

func (h *AuthHandler) SignupPage(c echo.Context) error {
	return c.Render(http.StatusOK, "signup.html", map[string]interface{}{
		"LoggedIn": middleware.IsLoggedIn(c),
	})
}

func (h *AuthHandler) Signup(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.Render(http.StatusBadRequest, "signup.html", map[string]interface{}{
			"Error": "Username and password are required.",
		})
	}

	var existing models.User
	if err := database.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		return c.Render(http.StatusConflict, "signup.html", map[string]interface{}{
			"Error": "This username is already taken.",
		})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "signup.html", map[string]interface{}{
			"Error": "Internal server error.",
		})
	}

	user := models.User{Username: username, Password: string(hash)}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Render(http.StatusInternalServerError, "signup.html", map[string]interface{}{
			"Error": "Could not create your account.",
		})
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *AuthHandler) LoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"LoggedIn": middleware.IsLoggedIn(c),
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{
			"Error": "Incorrect username or password.",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{
			"Error": "Incorrect username or password.",
		})
	}

	sess, _ := h.Store.Get(c.Request(), middleware.SessionName)
	sess.Values["user_id"] = user.ID
	sess.Values["username"] = user.Username
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.Render(http.StatusInternalServerError, "login.html", map[string]interface{}{
			"Error": "Failed to create session.",
		})
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	sess, _ := h.Store.Get(c.Request(), middleware.SessionName)
	sess.Options.MaxAge = -1 // delete cookie
	_ = sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusSeeOther, "/")
}

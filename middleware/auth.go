package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const SessionName = "weblog_session"

// LoadUser reads the session cookie (if any) and stashes the current
// user's ID and username on the echo.Context so every handler/template
// can check "who is logged in" without repeating session code.
func LoadUser(store sessions.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, _ := store.Get(c.Request(), SessionName)

			if uid, ok := sess.Values["user_id"].(uint); ok && uid != 0 {
				c.Set("user_id", uid)
				c.Set("username", sess.Values["username"])
				c.Set("logged_in", true)
			} else {
				c.Set("logged_in", false)
			}
			return next(c)
		}
	}
}

// RequireAuth blocks a route unless the user is logged in.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if loggedIn, _ := c.Get("logged_in").(bool); !loggedIn {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		return next(c)
	}
}

// CurrentUserID is a small helper for handlers.
func CurrentUserID(c echo.Context) uint {
	if uid, ok := c.Get("user_id").(uint); ok {
		return uid
	}
	return 0
}

func IsLoggedIn(c echo.Context) bool {
	loggedIn, _ := c.Get("logged_in").(bool)
	return loggedIn
}

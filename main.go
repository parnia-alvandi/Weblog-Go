package main

import (
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"weblog/config"
	"weblog/database"
	"weblog/handlers"
	appmw "weblog/middleware"
)

// Renderer implements echo.Renderer using html/template.
// Each page template is pre-combined with the shared layout at startup.
type Renderer struct {
	templates map[string]*template.Template
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return echo.NewHTTPError(500, "template not found: "+name)
	}
	return tmpl.ExecuteTemplate(w, "layout", data)
}

func newRenderer() *Renderer {
	pages := []string{"home.html", "detail.html", "login.html", "signup.html", "create.html"}
	templates := make(map[string]*template.Template)

	for _, page := range pages {
		t := template.Must(template.ParseFiles(
			filepath.Join("templates", "layout.html"),
			filepath.Join("templates", page),
		))
		templates[page] = t
	}
	return &Renderer{templates: templates}
}

func main() {
	// .env is optional — on platforms like Railway/Render, env vars are
	// injected directly and this file won't exist. That's fine.
	_ = godotenv.Load()

	uploadDir := filepath.Join("static", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("could not create upload dir: %v", err)
	}

	database.Connect()

	store := sessions.NewCookieStore([]byte(config.SessionSecret()))

	e := echo.New()
	e.Renderer = newRenderer()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(appmw.LoadUser(store))

	e.Static("/static", "static")
	e.Static("/uploads", uploadDir)

	authHandler := handlers.NewAuthHandler(store)
	postHandler := handlers.NewPostHandler(uploadDir)
	commentHandler := handlers.NewCommentHandler()

	// Public routes
	e.GET("/", postHandler.Home)
	e.GET("/weblog/:id", postHandler.Detail)
	e.GET("/signup", authHandler.SignupPage)
	e.POST("/signup", authHandler.Signup)
	e.GET("/login", authHandler.LoginPage)
	e.POST("/login", authHandler.Login)
	e.POST("/logout", authHandler.Logout)

	// Routes that require login
	auth := e.Group("")
	auth.Use(appmw.RequireAuth)
	auth.GET("/weblog/new", postHandler.CreatePage)
	auth.POST("/weblog/new", postHandler.Create)
	auth.POST("/weblog/:id/delete", postHandler.Delete)
	auth.POST("/weblog/:id/comments", commentHandler.Create)

	log.Printf("listening on :%s", config.Port())
	e.Logger.Fatal(e.Start(":" + config.Port()))
}

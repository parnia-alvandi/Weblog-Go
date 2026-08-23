# Weblog

A lightweight weblog application where users can publish public or private blog posts, share private posts with specific people, and comment on posts they have access to. Built with **Go**, the **Echo** web framework, and **PostgreSQL**.

## Features

- **Authentication** — sign up and log in with a username and password (passwords are hashed with bcrypt)
- **Posts (boards)** — create posts with a title, content, an optional image, and a privacy status (public/private)
- **Private sharing** — when creating a private post, share it with specific usernames; only the author and those users can view it
- **Home feed** — shows every public post plus any private posts the logged-in user authored or was shared
- **Post detail page** (`/weblog/{id}`) — full content, image, and comment thread
- **Comments** — logged-in users can comment on any post they have permission to view
- **Ownership rules** — only a post's author can delete it; there is no edit functionality by design

## Tech Stack

- **Language:** Go 1.22+
- **Web framework:** [Echo](https://echo.labstack.com/)
- **ORM:** [GORM](https://gorm.io/) with the PostgreSQL driver
- **Database:** PostgreSQL
- **Rendering:** Server-side rendering with Go's `html/template`
- **Sessions:** cookie-based sessions via `gorilla/sessions`
- **Password hashing:** bcrypt

## Getting Started

### Prerequisites

- Go 1.22 or later
- PostgreSQL (or Docker, to run PostgreSQL in a container)

### 1. Clone and configure

\`\`\`bash
git clone <your-repo-url>
cd weblog
cp .env.example .env
\`\`\`

Edit `.env` and set your database credentials and a strong `SESSION_SECRET`.

### 2. Start PostgreSQL

If you don't already have PostgreSQL running locally, spin one up with Docker:

\`\`\`bash
docker compose up -d
\`\`\`

### 3. Install dependencies and run

\`\`\`bash
go mod tidy
go run main.go
\`\`\`

The app will be available at `http://localhost:8080`. The database schema is created automatically on startup via GORM's auto-migration.

## Deployment

A `Dockerfile` is included and works out of the box on platforms such as Railway, Render, or Fly.io.

1. Push this repository to GitHub.
2. Create a new service on your platform of choice from this repo (the `Dockerfile` will be detected automatically).
3. Add a PostgreSQL database on the same platform.
4. Set the following environment variables on the app service: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`, `DB_SSLMODE` (use `require` for most managed databases), and `SESSION_SECRET`.
5. Deploy and grab the public URL from your platform's dashboard.

> **Note:** Uploaded images are stored on local disk (`static/uploads`). On some free-tier hosting platforms the filesystem is ephemeral and may be wiped on redeploy — check your platform's current storage policy, or use an object storage service (e.g. S3, Cloudflare R2) for persistent image storage in production.

## Project Structure

\`\`\`
weblog/
├── main.go            # entry point, routes, template renderer
├── config/            # environment configuration
├── database/          # PostgreSQL connection and auto-migration
├── models/            # User, Post, Comment + access-control logic
├── handlers/          # auth, post, and comment request handlers
├── middleware/         # session loading and route protection
├── templates/         # HTML templates (Go html/template)
├── static/            # CSS and uploaded images
├── Dockerfile          # production image
└── docker-compose.yml  # local PostgreSQL for development
\`\`\`
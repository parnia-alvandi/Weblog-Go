package config

import "os"

func SessionSecret() string {
	if s := os.Getenv("SESSION_SECRET"); s != "" {
		return s
	}
	// Fallback only for local dev; ALWAYS set a real secret in production.
	return "dev-insecure-secret-change-me"
}

func Port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

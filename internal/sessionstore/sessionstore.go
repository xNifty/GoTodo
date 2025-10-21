package sessionstore

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var Store *sessions.CookieStore

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v\n", err)
	}

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("SESSION_KEY environment variable is not set")
	}
	if len(sessionKey) < 32 {
		log.Fatal("SESSION_KEY must be at least 32 characters long")
	}

	Store = sessions.NewCookieStore([]byte(sessionKey))

	Store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

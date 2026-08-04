// Jackbot is a Discord chat bot backed by OpenAI for replies and
// moderation, with conversation history mirrored to Turso.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/thesimpledev/Jackbot/internal/ai"
	"github.com/thesimpledev/Jackbot/internal/bot"
	"github.com/thesimpledev/Jackbot/internal/db"
	"github.com/thesimpledev/Jackbot/internal/history"
)

func main() {
	loadDotEnv()
	requireEnv("DISCORD_TOKEN", "OPENAI_TOKEN", "TURSO_DATABASE_URL", "TURSO_TOKEN")

	store, err := db.New(os.Getenv("TURSO_DATABASE_URL"), os.Getenv("TURSO_TOKEN"))
	if err != nil {
		log.Fatalf("database setup failed: %v", err)
	}

	botName := os.Getenv("BOT_NAME")
	hist := history.New(store, os.Getenv("BOT_PROMPT"))
	aiService := ai.New(os.Getenv("OPENAI_TOKEN"), os.Getenv("CHAT_MODEL"), botName, hist)
	handler := bot.NewHandler(aiService, hist, store, botName, os.Getenv("ANIMAL"))

	session, err := bot.New(os.Getenv("DISCORD_TOKEN"), botName, handler)
	if err != nil {
		log.Fatalf("discord setup failed: %v", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			log.Printf("failed to close discord session: %v", err)
		}
	}()

	waitForShutdown()
}

// loadDotEnv mirrors dotenv/config: a .env file is loaded when present
// and its absence is not an error.
func loadDotEnv() {
	if _, err := os.Stat(".env"); err != nil {
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}
}

func requireEnv(names ...string) {
	for _, name := range names {
		if os.Getenv(name) == "" {
			log.Fatalf("required environment variable %s is not set", name)
		}
	}
}

func waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	sig := <-stop
	log.Printf("received %s, shutting down", sig)
}

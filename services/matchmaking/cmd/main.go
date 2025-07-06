package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jyablonski/elohell/services/matchmaking/internal/matchmaking"
)

func main() {
	// Configure slog with JSON output for production
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	redisAddr := os.Getenv("REDIS_CONN")
	if redisAddr == "" {
		redisAddr = "redis:6379" // default fallback
	}

	dbURL := os.Getenv("DB_CONN")
	if dbURL == "" {
		slog.Error("DB_CONN environment variable is required")
		os.Exit(1)
	}

	mm, err := matchmaking.NewMatchmaker(redisAddr, dbURL)
	if err != nil {
		slog.Error("failed to initialize matchmaker", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting matchmaking loop...")

	// Create a background context for the long-running loop.
	ctx := context.Background()
	mm.BasicMatchmakingLoop(ctx)
}

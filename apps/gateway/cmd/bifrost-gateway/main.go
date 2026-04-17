package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/kittors/bifrost/apps/gateway/internal/config"
	"github.com/kittors/bifrost/apps/gateway/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load(nil)
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			return db.PingContext(pingCtx)
		},
		ReadyTime: time.Now().UTC().Format(time.RFC3339),
		Upstreams: cfg.Upstreams,
	})

	httpServer := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("gateway listening", "listenAddress", cfg.ListenAddress)

	if err := httpServer.ListenAndServe(); err != nil && !server.IsServerClosed(err) {
		logger.Error("gateway exited", "error", err)
		os.Exit(1)
	}
}

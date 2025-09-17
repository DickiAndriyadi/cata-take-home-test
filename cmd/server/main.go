package main

import (
	"cata-take-home-test/internal/api"
	"cata-take-home-test/internal/cache"
	"cata-take-home-test/internal/config"
	"cata-take-home-test/internal/db"
	"cata-take-home-test/internal/job"
	"cata-take-home-test/internal/logger"
	"cata-take-home-test/internal/pokeapi"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger()
	defer log.Sync()

	cfg := config.Load(log)

	mysql, err := db.NewMySQL(cfg.MySQLDSN, log)
	if err != nil {
		log.Fatal("Failed to initialize MySQL", zap.Error(err))
	}
	defer mysql.DB.Close()

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, log)
	defer redisCache.Close()

	apiClient := pokeapi.NewClient(cfg.APIBaseURL, cfg.APITimeout, cfg.MaxRetryAttempts, log)

	handler := api.NewHandler(apiClient, mysql, redisCache, log, cfg.CacheTTL)

	// HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.SyncHandler(w, r)
	})

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ItemsHandler(w, r)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: mux,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		log.Info("Shutting down server...")
		cancel()
		srv.Shutdown(context.Background())
	}()

	refresher := job.NewRefresher(handler, cfg.SyncInterval, log)
	go refresher.Start(ctx)

	log.Info("Server is starting", zap.String("port", cfg.ServerPort))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error", zap.Error(err))
	}
	log.Info("Server stopped")
}

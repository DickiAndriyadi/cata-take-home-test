package api

import (
	"cata-take-home-test/internal/cache"
	"cata-take-home-test/internal/db"
	"cata-take-home-test/internal/pokeapi"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Handler struct {
	apiClient  *pokeapi.Client
	mysql      *db.MySQL
	redisCache *cache.RedisCache
	logger     *zap.Logger
	cacheTTL   time.Duration
}

func NewHandler(apiClient *pokeapi.Client, mysql *db.MySQL, redisCache *cache.RedisCache, logger *zap.Logger, cacheTTL time.Duration) *Handler {
	return &Handler{
		apiClient:  apiClient,
		mysql:      mysql,
		redisCache: redisCache,
		logger:     logger,
		cacheTTL:   cacheTTL,
	}
}

// /sync handler fetches from external API and updates DB
func (h *Handler) SyncHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.Sync(ctx); err != nil {
		http.Error(w, "Failed to sync data from external API", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Sync completed successfully"}`))
}

// /items handler returns cached or DB pokemon list
func (h *Handler) ItemsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Try cache first
	cached, err := h.redisCache.Get(ctx, "items_cache")
	if err == nil && cached != "" {
		h.logger.Info("Serving /items response from cache")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	items, err := h.mysql.GetAllPokemon(ctx)
	if err != nil {
		h.logger.Error("Failed to fetch pokemon from DB", zap.Error(err))
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(items)
	if err != nil {
		h.logger.Error("Failed to marshal response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set cache asynchronously (don't block client)
	go func() {
		err := h.redisCache.Set(context.Background(), "items_cache", string(payload), h.cacheTTL)
		if err != nil {
			h.logger.Warn("Failed to cache /items response", zap.Error(err))
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// Sync performs the sync operation, can be called from background job
func (h *Handler) Sync(ctx context.Context) error {
	h.logger.Info("Starting sync with external API")

	results, err := h.apiClient.FetchPokemonList(ctx, 20)
	if err != nil {
		h.logger.Error("Failed to fetch pokemon list", zap.Error(err))
		return err
	}

	for _, res := range results {
		detail, err := h.apiClient.FetchPokemonDetails(ctx, res.URL)
		if err != nil {
			h.logger.Error("Failed to fetch pokemon detail", zap.String("url", res.URL), zap.Error(err))
			continue
		}
		err = h.mysql.UpsertPokemon(ctx, detail.ID, detail.Name, detail.BaseExperience)
		if err != nil {
			h.logger.Error("Failed to upsert pokemon", zap.Int("id", detail.ID), zap.Error(err))
		}
	}

	// Invalidate cache after sync
	err = h.redisCache.Delete(ctx, "items_cache")
	if err != nil {
		h.logger.Warn("Failed to invalidate items cache", zap.Error(err))
	}

	h.logger.Info("Sync completed successfully")
	return nil
}

package tests

import (
	"cata-take-home-test/internal/api"
	"cata-take-home-test/internal/cache"
	"cata-take-home-test/internal/db"
	"cata-take-home-test/internal/logger"
	"cata-take-home-test/internal/pokeapi"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestItemsHandler(t *testing.T) {
	log := logger.NewLogger()
	defer log.Sync()

	mysql, err := db.NewMySQL("root:password@tcp(localhost:3306)/pokemondb?parseTime=true", log)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}
	redisCache := cache.NewRedisCache("localhost:6379", "", log)
	apiClient := pokeapi.NewClient("https://pokeapi.co/api/v2", 5*time.Second, 3, log)
	handler := api.NewHandler(apiClient, mysql, redisCache, log, 5*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	w := httptest.NewRecorder()

	handler.ItemsHandler(w, req)

	if status := w.Result().StatusCode; status != http.StatusOK {
		t.Errorf("unexpected status code: got %v want %v", status, http.StatusOK)
	}
}

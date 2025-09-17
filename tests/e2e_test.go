package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Pastikan struktur Pokemon sesuai
type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

func TestEndToEndSyncAndCaching(t *testing.T) {
	// Tambahkan timeout untuk mencegah hanging
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// --- Trigger /sync endpoint ---
	resp, err := client.Get("http://localhost:8080/sync")
	assert.NoError(t, err, "Sync request should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Sync should return OK status")

	// --- Fetch items ---
	resp, err = client.Get("http://localhost:8080/items")
	assert.NoError(t, err, "Items request should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Items should return OK status")

	// Dekode JSON
	var pokemons []Pokemon
	err = json.NewDecoder(resp.Body).Decode(&pokemons)
	assert.NoError(t, err, "Should decode Pokemon JSON")

	// Validasi data
	assert.NotEmpty(t, pokemons, "Pokemon list should not be empty")
}

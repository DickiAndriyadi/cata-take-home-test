package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type MySQL struct {
	DB     *sql.DB
	logger *zap.Logger
}

func NewMySQL(dsn string, logger *zap.Logger) (*MySQL, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	m := &MySQL{
		DB:     db,
		logger: logger,
	}
	if err := m.ensureSchema(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MySQL) ensureSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS pokemon (
    id INT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    base_experience INT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
`
	_, err := m.DB.Exec(schema)
	if err != nil {
		m.logger.Error("Error ensuring schema", zap.Error(err))
	}
	return err
}

// UpsertPokemon inserts or updates a Pokémon, idempotent
func (m *MySQL) UpsertPokemon(ctx context.Context, id int, name string, baseExp int) error {
	query := `
INSERT INTO pokemon (id, name, base_experience) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE name = VALUES(name), base_experience = VALUES(base_experience), updated_at = CURRENT_TIMESTAMP;
`
	_, err := m.DB.ExecContext(ctx, query, id, name, baseExp)
	if err != nil {
		m.logger.Error("Failed to upsert pokemon", zap.Int("id", id), zap.Error(err))
	}
	return err
}

// GetAllPokemon returns all Pokémon from DB
func (m *MySQL) GetAllPokemon(ctx context.Context) ([]Pokemon, error) {
	rows, err := m.DB.QueryContext(ctx, "SELECT id, name, base_experience FROM pokemon ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pokemons []Pokemon
	for rows.Next() {
		var p Pokemon
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseExperience); err != nil {
			return nil, err
		}
		pokemons = append(pokemons, p)
	}
	return pokemons, rows.Err()
}

// Pokemon struct used in DB package (also can be reused)
type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

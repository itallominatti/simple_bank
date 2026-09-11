package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Row, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func Connect(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("abrir conexão: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	const maxAttempts = 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err == nil {
			return db, nil
		}
		log.Printf("banco ainda não respondeu (tentativa %d/%d): %v", attempt, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}

	db.Close()
	return nil, fmt.Errorf("Não foi possivel conectar ao banco: %w", err)
}

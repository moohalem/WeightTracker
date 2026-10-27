// Package db opens the SQLite database and provides measurement operations.
package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := createSchema(ctx, database); err != nil {
		database.Close()
		return nil, fmt.Errorf("create database schema: %w", err)
	}

	return database, nil
}

func createSchema(ctx context.Context, database *sql.DB) error {
	_, err := database.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			weight REAL NOT NULL,
			measuringTime DATE NOT NULL
		)
	`)
	return err
}

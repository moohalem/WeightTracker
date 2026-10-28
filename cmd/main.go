// Package main starts the WeightTracker command-line application.
package main

import (
	"context"
	"log"
	"os"
	"time"

	db "github.com/moohalem/WeightTracker/internal/db"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.Open(ctx, "app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := MainMenu(database, os.Stdin, os.Stdout); err != nil {
		log.Printf("menu error: %v", err)
	}
}

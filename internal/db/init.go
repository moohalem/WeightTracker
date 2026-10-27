package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"

	utils "github.com/moohalem/WeightTracker/internal/utils"
)

func Init() error {
	dbName := "app.db"
	exists, err := utils.FileExist(dbName)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Println("DB does not exist, creating one ...")
		if err := createDB(dbName); err != nil {
			return err
		}
		createTable(dbName)
		utils.ClearScreen()
		fmt.Println("Database created")
	}
	return nil
}

func createDB(dbName string) error {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}
	return nil
}

func createTable(dbName string) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			age INTEGER NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

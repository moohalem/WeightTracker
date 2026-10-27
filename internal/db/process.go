package db

import (
	"database/sql"
	"fmt"
	"time"
)

func AddData(db *sql.DB, userName string, userWeight float64) error {
	measuringTime := time.Now()
	_, err := db.Exec(
		`
		INSERT INTO users (name, weight, measuringTime)
		VALUES (?, ?, ?)`,
		userName,
		userWeight,
		measuringTime.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Measurement saved! Name: %s, Weight: %.2f kg, Time: %s\n",
		userName,
		userWeight,
		measuringTime.Format(time.RFC3339),
	)
	return nil
}

func SelectData(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT id, name, weight, measuringTime
     	FROM users
		ORDER BY measuringTime DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var id int64
		var name string
		var weight float64
		var measuringTime string

		err := rows.Scan(&id, &name, &weight, &measuringTime)
		if err != nil {
			return err
		}

		fmt.Printf(
			"ID: %d | Name: %s | Weight: %.2f kg | Time: %s\n",
			id, name, weight, measuringTime,
		)
	}

	if !found {
		fmt.Println("No measurements found.")
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

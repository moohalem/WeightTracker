// Package db opens the SQLite database and provides measurement operations.
package db

import (
	"database/sql"
	"time"
)

type Measurement struct {
	ID            int64
	Name          string
	Weight        float64
	MeasuringTime string
}

func AddData(database *sql.DB, userName string, userWeight float64) error {
	measuringTime := time.Now().Format(time.RFC3339)
	_, err := database.Exec(
		`
		INSERT INTO users (name, weight, measuringTime)
		VALUES (?, ?, ?)`,
		userName,
		userWeight,
		measuringTime,
	)
	return err
}

func RecentMeasurements(database *sql.DB) ([]Measurement, error) {
	return queryMeasurements(database, `
		SELECT id, name, weight, measuringTime
		FROM users
		ORDER BY measuringTime DESC, id DESC
		LIMIT 10
	`)
}

func MeasurementsByName(database *sql.DB, name string) ([]Measurement, error) {
	return queryMeasurements(database, `
		SELECT id, name, weight, measuringTime
		FROM users
		WHERE name = ?
		ORDER BY measuringTime DESC, id DESC
	`, name)
}

func Names(database *sql.DB) ([]string, error) {
	rows, err := database.Query(`
		SELECT DISTINCT name
		FROM users
		ORDER BY name COLLATE NOCASE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

func queryMeasurements(database *sql.DB, query string, args ...any) ([]Measurement, error) {
	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measurements []Measurement
	for rows.Next() {
		var measurement Measurement
		if err := rows.Scan(
			&measurement.ID,
			&measurement.Name,
			&measurement.Weight,
			&measurement.MeasuringTime,
		); err != nil {
			return nil, err
		}
		measurements = append(measurements, measurement)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return measurements, nil
}

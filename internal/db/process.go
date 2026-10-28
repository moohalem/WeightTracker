// Package db opens the SQLite database and provides measurement operations.
package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode"
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
		normalizeName(userName),
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
		WHERE lower(name) = ?
		ORDER BY measuringTime DESC, id DESC
	`, normalizeName(name))
}

func UpdateMeasurementWeight(database *sql.DB, id int64, weight float64) error {
	_, err := database.Exec("UPDATE users SET weight = ? WHERE id = ?", weight, id)
	return err
}

func DeleteMeasurement(database *sql.DB, id int64) error {
	result, err := database.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return fmt.Errorf("measurement %d was not found", id)
	}
	return nil
}

func DeleteMeasurementsByName(database *sql.DB, name string) error {
	_, err := database.Exec("DELETE FROM users WHERE lower(name) = ?", normalizeName(name))
	return err
}

func Names(database *sql.DB) ([]string, error) {
	rows, err := database.Query(`
		SELECT DISTINCT lower(name)
		FROM users
		ORDER BY lower(name) COLLATE NOCASE
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
		names = append(names, displayName(name))
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
		measurement.Name = displayName(measurement.Name)
		measurements = append(measurements, measurement)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return measurements, nil
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func displayName(name string) string {
	runes := []rune(normalizeName(name))
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

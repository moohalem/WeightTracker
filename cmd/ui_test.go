package main

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	db "github.com/moohalem/WeightTracker/internal/db"
)

func openManagementTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "measurements.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	return database
}

func seedManagementTestDatabase(t *testing.T, database *sql.DB) {
	t.Helper()
	for index, weight := range []float64{60, 61} {
		name := "Ghina"
		if index == 1 {
			name = "ghina"
		}
		if err := db.AddData(database, name, weight); err != nil {
			t.Fatal(err)
		}
	}
}

func runManagementInput(t *testing.T, database *sql.DB, input string) string {
	t.Helper()
	var output strings.Builder
	if err := MainMenu(database, strings.NewReader(input), &output); err != nil {
		t.Fatalf("MainMenu returned an error: %v\nOutput:\n%s", err, output.String())
	}
	return output.String()
}

func TestManageMeasurementsUpdatesOnlySelectedWeight(t *testing.T) {
	database := openManagementTestDatabase(t)
	seedManagementTestDatabase(t, database)
	if _, err := database.Exec("UPDATE users SET measuringTime = ? WHERE weight = 60", "2025-02-01T12:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("UPDATE users SET measuringTime = ? WHERE weight = 61", "2024-02-01T12:00:00Z"); err != nil {
		t.Fatal(err)
	}

	output := runManagementInput(t, database, "m\nu\n1\n1\n70\n\nc\n0\n")
	if !strings.Contains(output, fmt.Sprintf("%s[1]%s  60.00 kg", yellow, reset)) {
		t.Fatal("update measurement choice should use yellow option numbers")
	}

	measurements, err := db.MeasurementsByName(database, "GHINA")
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 2 {
		t.Fatalf("got %d measurements, want 2", len(measurements))
	}
	if measurements[0].Weight != 70 || measurements[1].Weight != 61 {
		t.Fatalf("weights after update = %.2f, %.2f; want 70, 61", measurements[0].Weight, measurements[1].Weight)
	}
	if measurements[0].MeasuringTime != "2025-02-01T12:00:00Z" || measurements[1].MeasuringTime != "2024-02-01T12:00:00Z" {
		t.Fatalf("measurement dates changed or are out of order: %q, %q", measurements[0].MeasuringTime, measurements[1].MeasuringTime)
	}
}

func TestManageMeasurementsDeletesSelectedEntry(t *testing.T) {
	database := openManagementTestDatabase(t)
	seedManagementTestDatabase(t, database)

	runManagementInput(t, database, "m\nd\n1\ns\n1\ny\n\nc\n0\n")

	measurements, err := db.MeasurementsByName(database, "Ghina")
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 1 || measurements[0].Weight != 60 {
		t.Fatalf("remaining measurements = %#v, want one measurement of 60 kg", measurements)
	}
}

func TestManageMeasurementsDeletesOneOfDuplicateEntries(t *testing.T) {
	database := openManagementTestDatabase(t)
	for range 2 {
		if err := db.AddData(database, "Ghina", 60); err != nil {
			t.Fatal(err)
		}
	}

	measurements, err := db.MeasurementsByName(database, "Ghina")
	if err != nil {
		t.Fatal(err)
	}
	deletedID := measurements[0].ID

	output := runManagementInput(t, database, "m\nd\n1\ns\n1\ny\n\nc\n0\n")
	if !strings.Contains(output, fmt.Sprintf("%s[1]%s  60.00 kg", green, reset)) {
		t.Fatal("measurement selection does not show the green option number and weight")
	}
	if !strings.Contains(output, fmt.Sprintf("%s[a]%s", red, reset)) || !strings.Contains(output, fmt.Sprintf("%s[s]%s", red, reset)) {
		t.Fatal("delete scope choices should use red option labels")
	}
	if strings.Contains(output, "ID ") || strings.Contains(output, fmt.Sprintf("Measurement #%d", deletedID)) {
		t.Fatal("delete flow should not display measurement IDs")
	}

	remaining, err := db.MeasurementsByName(database, "Ghina")
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 || remaining[0].ID == deletedID {
		t.Fatalf("remaining measurements = %#v, want only the other duplicate entry", remaining)
	}

	viewOutput := runManagementInput(t, database, "3\n1\n\n0\n")
	if strings.Contains(viewOutput, fmt.Sprintf("%s %d%s", yellow, deletedID, reset)) {
		t.Fatalf("deleted measurement #%d still appears in the name view", deletedID)
	}
}

func TestManageMeasurementsDeletesEntireName(t *testing.T) {
	database := openManagementTestDatabase(t)
	seedManagementTestDatabase(t, database)

	runManagementInput(t, database, "m\nd\n1\na\ny\n\nc\n0\n")

	measurements, err := db.MeasurementsByName(database, "ghina")
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 0 {
		t.Fatalf("got %d measurements after deleting the name, want 0", len(measurements))
	}
}

func TestManageMeasurementsCanCancelEntrySelection(t *testing.T) {
	database := openManagementTestDatabase(t)
	seedManagementTestDatabase(t, database)

	runManagementInput(t, database, "m\nu\n1\nc\nc\n0\n")

	measurements, err := db.MeasurementsByName(database, "ghina")
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 2 || measurements[0].Weight != 61 || measurements[1].Weight != 60 {
		t.Fatalf("measurements changed after cancel: %#v", measurements)
	}
}

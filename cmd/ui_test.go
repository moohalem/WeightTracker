package main

import (
	"context"
	"database/sql"
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

func runManagementInput(t *testing.T, database *sql.DB, input string) {
	t.Helper()
	var output strings.Builder
	if err := MainMenu(database, strings.NewReader(input), &output); err != nil {
		t.Fatalf("MainMenu returned an error: %v\nOutput:\n%s", err, output.String())
	}
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

	runManagementInput(t, database, "m\nu\n1\n1\n70\n\nc\n0\n")

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

	runManagementInput(t, database, "m\nd\n1\nm\n1\ny\n\nc\n0\n")

	measurements, err := db.MeasurementsByName(database, "Ghina")
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 1 || measurements[0].Weight != 60 {
		t.Fatalf("remaining measurements = %#v, want one measurement of 60 kg", measurements)
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

package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestNamesNormalizeCaseAndLookup(t *testing.T) {
	database, err := Open(context.Background(), filepath.Join(t.TempDir(), "measurements.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	for _, name := range []string{"Ghina", "ghina", "ghInA"} {
		if err := AddData(database, name, 60); err != nil {
			t.Fatal(err)
		}
	}

	var storedName string
	if err := database.QueryRow("SELECT name FROM users ORDER BY id LIMIT 1").Scan(&storedName); err != nil {
		t.Fatal(err)
	}
	if storedName != "ghina" {
		t.Fatalf("stored name = %q, want %q", storedName, "ghina")
	}

	names, err := Names(database)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "Ghina" {
		t.Fatalf("names = %q, want [Ghina]", names)
	}

	measurements, err := MeasurementsByName(database, "gHiNa")
	if err != nil {
		t.Fatal(err)
	}
	if len(measurements) != 3 {
		t.Fatalf("got %d measurements, want 3", len(measurements))
	}
	for _, measurement := range measurements {
		if measurement.Name != "Ghina" {
			t.Errorf("returned name = %q, want %q", measurement.Name, "Ghina")
		}
	}
}

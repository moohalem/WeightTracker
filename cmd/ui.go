package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	db "github.com/moohalem/WeightTracker/internal/db"
)

func MainMenu(database *sql.DB, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)

	for {
		fmt.Fprint(output, "\033[2J\033[H")
		fmt.Fprintln(output, "\nWEIGHT TRACKER")
		fmt.Fprintln(output, "==============")
		fmt.Fprintln(output, "1. Add measurement")
		fmt.Fprintln(output, "2. View latest 10 measurements")
		fmt.Fprintln(output, "3. View measurements by name")
		fmt.Fprintln(output, "0. Exit")
		fmt.Fprint(output, "Choose an option: ")

		if !scanner.Scan() {
			return scanner.Err()
		}

		choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil {
			fmt.Fprintln(output, "Please enter 1, 2, or 0.")
			continue
		}

		switch choice {
		case 1:
			if err := addMeasurement(scanner, output, database); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				fmt.Fprintf(output, "Could not save measurement: %v\n", err)
			}
		case 2:
			measurements, err := db.RecentMeasurements(database)
			if err != nil {
				fmt.Fprintf(output, "Could not load measurements: %v\n", err)
			} else {
				renderMeasurements(output, "LATEST 10 MEASUREMENTS", measurements)
				if err := pauseForMenu(scanner, output); err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
			}
		case 3:
			if err := viewMeasurementsByName(scanner, output, database); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				fmt.Fprintf(output, "Could not load measurements: %v\n", err)
			}
		case 0:
			fmt.Fprintln(output, "Goodbye.")
			return nil
		default:
			fmt.Fprintln(output, "Please enter 1, 2, 3, or 0.")
		}
	}
}

func addMeasurement(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	name, err := readLine(scanner, output, "Name: ")
	if err != nil {
		return err
	}
	if name == "" {
		fmt.Fprintln(output, "Name cannot be empty.")
		return nil
	}

	weightText, err := readLine(scanner, output, "Weight (kg): ")
	if err != nil {
		return err
	}
	weight, err := strconv.ParseFloat(weightText, 64)
	if err != nil || weight <= 0 {
		fmt.Fprintln(output, "Enter a valid weight greater than zero.")
		return nil
	}

	if err := db.AddData(database, name, weight); err != nil {
		return err
	}
	fmt.Fprintln(output, "Measurement saved.")
	return nil
}

func viewMeasurementsByName(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	names, err := db.Names(database)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintln(output, "No names found. Add a measurement first.")
		return nil
	}

	fmt.Fprintln(output, "\nChoose a name:")
	for index, name := range names {
		fmt.Fprintf(output, "%d. %s\n", index+1, name)
	}
	choiceText, err := readLine(scanner, output, "Name number: ")
	if err != nil {
		return err
	}
	choice, err := strconv.Atoi(choiceText)
	if err != nil || choice < 1 || choice > len(names) {
		fmt.Fprintln(output, "Please choose a number from the list.")
		return nil
	}

	name := names[choice-1]
	measurements, err := db.MeasurementsByName(database, name)
	if err != nil {
		return err
	}
	renderMeasurements(output, "MEASUREMENTS FOR "+name, measurements)
	if err := pauseForMenu(scanner, output); err != nil {
		return err
	}
	return nil
}

func renderMeasurements(output io.Writer, title string, measurements []db.Measurement) {
	fmt.Fprint(output, "\033[2J\033[H")
	fmt.Fprintln(output, "WEIGHT TRACKER")
	fmt.Fprintln(output, strings.Repeat("=", 72))
	fmt.Fprintln(output, title)
	fmt.Fprintln(output)

	if len(measurements) == 0 {
		fmt.Fprintln(output, "No measurements found.")
		return
	}

	table := tabwriter.NewWriter(output, 0, 0, 3, ' ', 0)
	fmt.Fprintln(table, "ID\tNAME\tWEIGHT\tMEASURED AT")
	fmt.Fprintln(table, "--\t----\t------\t------------")
	for _, measurement := range measurements {
		fmt.Fprintf(
			table,
			"%d\t%s\t%.2f kg\t%s\n",
			measurement.ID,
			measurement.Name,
			measurement.Weight,
			formatMeasurementTime(measurement.MeasuringTime),
		)
	}
	table.Flush()
}

func formatMeasurementTime(value string) string {
	measuredAt, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return measuredAt.Local().Format("2006-01-02 15:04")
}

func pauseForMenu(scanner *bufio.Scanner, output io.Writer) error {
	fmt.Fprint(output, "\nPress Enter to return to the menu...")
	if !scanner.Scan() {
		return scannerError(scanner)
	}
	return nil
}

func readLine(scanner *bufio.Scanner, output io.Writer, prompt string) (string, error) {
	fmt.Fprint(output, prompt)
	if !scanner.Scan() {
		return "", scannerError(scanner)
	}
	return strings.TrimSpace(scanner.Text()), nil
}

func scannerError(scanner *bufio.Scanner) error {
	if err := scanner.Err(); err != nil {
		return err
	}
	return io.EOF
}

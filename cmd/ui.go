package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"strings"

	db "github.com/moohalem/WeightTracker/internal/db"
)

func MainMenu(database *sql.DB, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)

	for {
		fmt.Fprintln(output, "\nWEIGHT TRACKER")
		fmt.Fprintln(output, "==============")
		fmt.Fprintln(output, "1. Add measurement")
		fmt.Fprintln(output, "2. View measurements")
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
				if err == io.EOF {
					return nil
				}
				fmt.Fprintf(output, "Could not save measurement: %v\n", err)
			}
		case 2:
			if err := db.SelectData(database); err != nil {
				fmt.Fprintf(output, "Could not load measurements: %v\n", err)
			}
		case 0:
			fmt.Fprintln(output, "Goodbye.")
			return nil
		default:
			fmt.Fprintln(output, "Please enter 1, 2, or 0.")
		}
	}
}

func addMeasurement(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	fmt.Fprint(output, "Name: ")
	if !scanner.Scan() {
		return scannerError(scanner)
	}
	name := strings.TrimSpace(scanner.Text())
	if name == "" {
		fmt.Fprintln(output, "Name cannot be empty.")
		return nil
	}

	fmt.Fprint(output, "Weight (kg): ")
	if !scanner.Scan() {
		return scannerError(scanner)
	}
	weight, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64)
	if err != nil || weight <= 0 {
		fmt.Fprintln(output, "Enter a valid weight greater than zero.")
		return nil
	}

	return db.AddData(database, name, weight)
}

func scannerError(scanner *bufio.Scanner) error {
	if err := scanner.Err(); err != nil {
		return err
	}
	return io.EOF
}

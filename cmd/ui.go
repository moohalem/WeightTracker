// Package main starts the WeightTracker command-line application.
package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	db "github.com/moohalem/WeightTracker/internal/db"
)

const (
	reset       = "\033[0m"
	bold        = "\033[1m"
	dim         = "\033[2m"
	red         = "\033[91m"
	green       = "\033[92m"
	yellow      = "\033[93m"
	cyan        = "\033[96m"
	white       = "\033[97m"
	bannerWidth = 44
)

func MainMenu(database *sql.DB, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)

	for {
		fmt.Fprint(output, "\033[2J\033[H")
		printBanner(output)
		printSeparator(output)
		printMenuItem(output, "1", "Add measurement")
		printMenuItem(output, "2", "View latest 10 measurements")
		printMenuItem(output, "3", "View measurements by name")
		printMenuItem(output, "m", "[M]anage measurements")
		printMenuItem(output, "e", "[E]xit")
		fmt.Fprintf(output, "\n  %s▸%s Choose an option: ", cyan, reset)

		if !scanner.Scan() {
			return scanner.Err()
		}

		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
		switch choice {
		case "1":
			if err := addMeasurement(scanner, output, database); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				fmt.Fprintf(output, "  %s✗ Could not save measurement: %v%s\n", red, err, reset)
			}
		case "2":
			measurements, err := db.RecentMeasurements(database)
			if err != nil {
				fmt.Fprintf(output, "  %s✗ Could not load measurements: %v%s\n", red, err, reset)
			} else {
				renderMeasurements(output, "LATEST 10 MEASUREMENTS", measurements)
				if err := pauseForMenu(scanner, output); err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
			}
		case "3":
			if err := viewMeasurementsByName(scanner, output, database); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				fmt.Fprintf(output, "  %s✗ Could not load measurements: %v%s\n", red, err, reset)
			}
		case "m":
			if err := manageMeasurements(scanner, output, database); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				fmt.Fprintf(output, "  %s✗ Could not manage measurements: %v%s\n", red, err, reset)
			}
		case "e":
			fmt.Fprintf(output, "  %s✓ Goodbye.%s\n", green, reset)
			return nil
		default:
			fmt.Fprintf(output, "  %s✗ Please enter 1, 2, 3, m, or 0.%s\n", red, reset)
		}
	}
}

func addMeasurement(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	names, err := db.Names(database)
	if err != nil {
		return err
	}

	var name string
	if len(names) == 0 {
		name, err = readLine(scanner, output, "Name: ")
	} else {
		var canceled bool
		name, canceled, err = chooseMeasurementName(scanner, output, names)
		if canceled {
			return nil
		}
	}
	if err != nil {
		return err
	}
	if name == "" {
		fmt.Fprintf(output, "  %s✗ Name cannot be empty.%s\n", red, reset)
		return nil
	}

	weightText, err := readLine(scanner, output, "Weight (kg): ")
	if err != nil {
		return err
	}
	weight, err := strconv.ParseFloat(weightText, 64)
	if err != nil || weight <= 0 {
		fmt.Fprintf(output, "  %s✗ Enter a valid weight greater than zero.%s\n", red, reset)
		return nil
	}

	if err := db.AddData(database, name, weight); err != nil {
		return err
	}
	fmt.Fprintf(output, "  %s✓ Measurement saved.%s\n", green, reset)
	return nil
}

func chooseMeasurementName(scanner *bufio.Scanner, output io.Writer, names []string) (string, bool, error) {
	for {
		fmt.Fprint(output, "\033[2J\033[H")
		fmt.Fprintf(output, "\n  %sChoose an existing name or enter a new one:%s\n", bold+white, reset)
		for index, name := range names {
			fmt.Fprintf(output, "  %s[%s%d%s]%s  %s\n", dim, reset+bold+cyan, index+1, reset+dim, reset, name)
		}
		fmt.Fprintf(output, "  %s[a]%s  Add a new name\n", cyan, reset)
		fmt.Fprintf(output, "  %s[c]%s  Cancel\n", cyan, reset)

		choiceText, err := readLine(scanner, output, "Name option: ")
		if err != nil {
			return "", false, err
		}
		switch strings.ToLower(choiceText) {
		case "a":
			name, err := readLine(scanner, output, "Name: ")
			return name, false, err
		case "c":
			return "", true, nil
		}

		choice, err := strconv.Atoi(choiceText)
		if err != nil || choice < 1 || choice > len(names) {
			fmt.Fprintf(output, "  %s✗ Please choose a number from the list.%s\n", red, reset)
			continue
		}
		return names[choice-1], false, nil
	}
}

func viewMeasurementsByName(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	fmt.Fprint(output, "\033[2J\033[H")
	names, err := db.Names(database)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintf(output, "  %s No names found. Add a measurement first.%s\n", cyan, reset)
		return nil
	}

	for {
		fmt.Fprint(output, "\033[2J\033[H")
		fmt.Fprintf(output, "\n  %sChoose a name:%s\n", bold+white, reset)
		for index, name := range names {
			fmt.Fprintf(output, "  %s[%s%d%s]%s  %s\n", dim, reset+bold+cyan, index+1, reset+dim, reset, name)
		}
		fmt.Fprintf(output, "  %s[0]%s  Back to main menu\n", cyan, reset)

		choiceText, err := readLine(scanner, output, "Name number: ")
		if err != nil {
			return err
		}
		choice, err := strconv.Atoi(choiceText)
		if err != nil || choice < 0 || choice > len(names) {
			fmt.Fprint(output, "\033[2J\033[H")
			fmt.Fprintf(output, "  %s✗ Please choose a number from the list.%s\n", red, reset)
			time.Sleep(time.Second)
			continue
		}
		if choice == 0 {
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
}

func manageMeasurements(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	for {
		fmt.Fprint(output, "\033[2J\033[H")
		printBanner(output)
		printSeparator(output)
		fmt.Fprintf(output, "  %sMANAGE MEASUREMENTS%s\n\n", bold+white, reset)
		printMenuItem(output, "u", "Update measurement")
		printMenuItem(output, "d", "Delete measurement")
		printMenuItem(output, "c", "Cancel")

		choice, err := readLine(scanner, output, "Choose an option: ")
		if err != nil {
			return err
		}
		switch strings.ToLower(choice) {
		case "u":
			if err := updateMeasurement(scanner, output, database); err != nil {
				return err
			}
		case "d":
			if err := deleteMeasurement(scanner, output, database); err != nil {
				return err
			}
		case "c":
			return nil
		default:
			fmt.Fprintf(output, "  %s✗ Choose u, d, or c.%s\n", red, reset)
			if _, err := waitForManagementContinue(scanner, output); err != nil {
				return err
			}
		}
	}
}

func updateMeasurement(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	name, canceled, err := selectManagementName(scanner, output, database)
	if err != nil || canceled {
		return err
	}

	measurements, err := db.MeasurementsByName(database, name)
	if err != nil {
		return err
	}
	if len(measurements) == 0 {
		return showManagementNotice(scanner, output, "No measurements found for "+name+".")
	}

	measurement, canceled, err := selectManagementMeasurement(scanner, output, measurements, yellow)
	if err != nil || canceled {
		return err
	}

	fmt.Fprint(output, "\033[2J\033[H")
	fmt.Fprintf(output, "\n  %sUPDATE MEASUREMENT%s\n", bold+white, reset)
	fmt.Fprintf(output, "  %sWeight:%s %.2f kg\n", bold, reset, measurement.Weight)
	fmt.Fprintf(output, "  %sMeasured:%s %s\n\n", bold, reset, formatMeasurementTime(measurement.MeasuringTime))
	for {
		weightText, err := readLine(scanner, output, "New weight in kg (c to cancel): ")
		if err != nil {
			return err
		}
		if strings.EqualFold(weightText, "c") {
			return nil
		}
		weight, err := strconv.ParseFloat(weightText, 64)
		if err != nil || weight <= 0 {
			fmt.Fprintf(output, "  %s✗ Enter a valid weight greater than zero.%s\n", red, reset)
			continue
		}
		if err := db.UpdateMeasurementWeight(database, measurement.ID, weight); err != nil {
			return err
		}
		return showManagementNotice(scanner, output, "Measurement weight updated.")
	}
}

func deleteMeasurement(scanner *bufio.Scanner, output io.Writer, database *sql.DB) error {
	name, canceled, err := selectManagementName(scanner, output, database)
	if err != nil || canceled {
		return err
	}

	deleteAll, canceled, err := chooseDeleteScope(scanner, output, name)
	if err != nil || canceled {
		return err
	}
	if deleteAll {
		confirmed, err := confirmManagementAction(scanner, output, "Delete all measurements for "+name+"?")
		if err != nil || !confirmed {
			return err
		}
		if err := db.DeleteMeasurementsByName(database, name); err != nil {
			return err
		}
		return showManagementNotice(scanner, output, "All measurements for "+name+" were deleted.")
	}

	measurements, err := db.MeasurementsByName(database, name)
	if err != nil {
		return err
	}
	if len(measurements) == 0 {
		return showManagementNotice(scanner, output, "No measurements found for "+name+".")
	}
	measurement, canceled, err := selectManagementMeasurement(scanner, output, measurements, green)
	if err != nil || canceled {
		return err
	}
	confirmed, err := confirmManagementAction(
		scanner,
		output,
		fmt.Sprintf("Delete the %.2f kg measurement from %s?", measurement.Weight, formatMeasurementTime(measurement.MeasuringTime)),
	)
	if err != nil || !confirmed {
		return err
	}
	if err := db.DeleteMeasurement(database, measurement.ID); err != nil {
		return err
	}
	return showManagementNotice(scanner, output, "Measurement deleted.")
}

func selectManagementName(scanner *bufio.Scanner, output io.Writer, database *sql.DB) (string, bool, error) {
	names, err := db.Names(database)
	if err != nil {
		return "", false, err
	}
	if len(names) == 0 {
		return "", true, showManagementNotice(scanner, output, "No measurements to manage.")
	}

	fmt.Fprint(output, "\033[2J\033[H")
	fmt.Fprintf(output, "\n  %sCHOOSE A NAME%s\n", bold+white, reset)
	for index, name := range names {
		fmt.Fprintf(output, "  %s[%s%d%s]%s  %s\n", dim, reset+bold+cyan, index+1, reset+dim, reset, name)
	}
	fmt.Fprintf(output, "  %s[c]%s  Cancel\n", cyan, reset)
	for {
		choiceText, err := readLine(scanner, output, "Name number: ")
		if err != nil {
			return "", false, err
		}
		if strings.EqualFold(choiceText, "c") {
			return "", true, nil
		}
		choice, err := strconv.Atoi(choiceText)
		if err != nil || choice < 1 || choice > len(names) {
			fmt.Fprintf(output, "  %s✗ Choose a number from the list or c to cancel.%s\n", red, reset)
			continue
		}
		return names[choice-1], false, nil
	}
}

func selectManagementMeasurement(scanner *bufio.Scanner, output io.Writer, measurements []db.Measurement, optionColor string) (db.Measurement, bool, error) {
	fmt.Fprint(output, "\033[2J\033[H")
	fmt.Fprintf(output, "\n  %sCHOOSE A MEASUREMENT%s\n", bold+white, reset)
	for index, measurement := range measurements {
		fmt.Fprintf(output, "  %s[%d]%s  %.2f kg  %s\n",
			optionColor, index+1, reset,
			measurement.Weight, formatMeasurementTime(measurement.MeasuringTime))
	}
	fmt.Fprintf(output, "  %s[c]%s  Cancel\n", cyan, reset)
	for {
		choiceText, err := readLine(scanner, output, "Measurement number: ")
		if err != nil {
			return db.Measurement{}, false, err
		}
		if strings.EqualFold(choiceText, "c") {
			return db.Measurement{}, true, nil
		}
		choice, err := strconv.Atoi(choiceText)
		if err != nil || choice < 1 || choice > len(measurements) {
			fmt.Fprintf(output, "  %s✗ Choose a number from the list or c to cancel.%s\n", red, reset)
			continue
		}
		return measurements[choice-1], false, nil
	}
}

func chooseDeleteScope(scanner *bufio.Scanner, output io.Writer, name string) (bool, bool, error) {
	fmt.Fprint(output, "\033[2J\033[H")
	fmt.Fprintf(output, "\n  %sDELETE MEASUREMENTS FOR %s%s\n", bold+white, name, reset)
	fmt.Fprintf(output, "  %s[a]%s  Delete [a]ll measurements for this name\n", red, reset)
	fmt.Fprintf(output, "  %s[s]%s  Delete [s]ingle measurement\n", red, reset)
	fmt.Fprintf(output, "  %s[c]%s  Cancel\n", cyan, reset)
	for {
		choice, err := readLine(scanner, output, "Choose an option: ")
		if err != nil {
			return false, false, err
		}
		switch strings.ToLower(choice) {
		case "a":
			return true, false, nil
		case "s":
			return false, false, nil
		case "c":
			return false, true, nil
		default:
			fmt.Fprintf(output, "  %s✗ Choose a, s, or c.%s\n", red, reset)
		}
	}
}

func confirmManagementAction(scanner *bufio.Scanner, output io.Writer, prompt string) (bool, error) {
	fmt.Fprint(output, "\033[2J\033[H")
	fmt.Fprintf(output, "\n  %s%s%s\n", bold+white, prompt, reset)
	fmt.Fprintf(output, "  %s[y]%s  Confirm\n", cyan, reset)
	fmt.Fprintf(output, "  %s[c]%s  Cancel\n", cyan, reset)
	for {
		choice, err := readLine(scanner, output, "Choose an option: ")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(choice) {
		case "y":
			return true, nil
		case "c":
			return false, nil
		default:
			fmt.Fprintf(output, "  %s✗ Choose y or c.%s\n", red, reset)
		}
	}
}

func showManagementNotice(scanner *bufio.Scanner, output io.Writer, message string) error {
	fmt.Fprint(output, "\033[2J\033[H")
	fmt.Fprintf(output, "\n  %s%s%s\n", cyan, message, reset)
	_, err := waitForManagementContinue(scanner, output)
	return err
}

func waitForManagementContinue(scanner *bufio.Scanner, output io.Writer) (bool, error) {
	choice, err := readLine(scanner, output, "Press Enter or c to return: ")
	if err != nil {
		return false, err
	}
	return strings.EqualFold(choice, "c"), nil
}

func renderMeasurements(output io.Writer, title string, measurements []db.Measurement) {
	fmt.Fprint(output, "\033[2J\033[H")
	printBanner(output)
	printSeparator(output)
	fmt.Fprintf(output, "  %s%s%s\n\n", bold+white, title, reset)

	if len(measurements) == 0 {
		fmt.Fprintf(output, "  %sℹ No measurements found.%s\n", cyan, reset)
		return
	}

	rows := make([][]string, 0, len(measurements)+1)
	rows = append(rows, []string{"ID", "NAME", "WEIGHT", "MEASURED AT"})
	for _, measurement := range measurements {
		rows = append(rows, []string{
			strconv.FormatInt(measurement.ID, 10),
			measurement.Name,
			fmt.Sprintf("%.2f kg", measurement.Weight),
			formatMeasurementTime(measurement.MeasuringTime),
		})
	}
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		for column, value := range row {
			if width := utf8.RuneCountInString(value); width > widths[column] {
				widths[column] = width
			}
		}
	}
	printTableBorder(output, widths, "┌", "┬", "┐")
	for rowIndex, row := range rows {
		fmt.Fprint(output, "  "+dim+"│"+reset)
		for column, value := range row {
			style := ""
			if rowIndex == 0 {
				style = bold + white
			} else if column == 0 {
				style = yellow
			} else if column == 2 {
				style = green
			}
			if column == 0 || column == 2 {
				fmt.Fprintf(output, " %s%s%s %s", style, padLeft(value, widths[column]), reset, dim+"│"+reset)
			} else {
				fmt.Fprintf(output, " %s%s%s %s", style, padRight(value, widths[column]), reset, dim+"│"+reset)
			}
		}
		fmt.Fprintln(output)
		if rowIndex == 0 {
			printTableBorder(output, widths, "├", "┼", "┤")
		}
	}
	printTableBorder(output, widths, "└", "┴", "┘")
}

func printBanner(output io.Writer) {
	const title = "W E I G H T   T R A C K E R"
	fmt.Fprintln(output)
	fmt.Fprintf(output, "  %s╔%s╗%s\n", bold+cyan, strings.Repeat("═", bannerWidth), reset)
	padding := (bannerWidth - utf8.RuneCountInString(title)) / 2
	rightPadding := bannerWidth - padding - utf8.RuneCountInString(title)
	fmt.Fprintf(output, "  %s║%s%s%s%s%s║%s\n",
		bold+cyan, reset, strings.Repeat(" ", padding), bold+yellow+title+reset+bold+cyan,
		strings.Repeat(" ", rightPadding), bold+cyan, reset)
	fmt.Fprintf(output, "  %s╚%s╝%s\n\n", bold+cyan, strings.Repeat("═", bannerWidth), reset)
}

func printSeparator(output io.Writer) {
	fmt.Fprintf(output, "  %s%s%s\n\n", dim, strings.Repeat("─", bannerWidth+2), reset)
}

func printMenuItem(output io.Writer, key, label string) {
	fmt.Fprintf(output, "  %s[%s%s%s]%s  %s\n", dim, reset+bold+cyan, key, reset+dim, reset, label)
}

func printTableBorder(output io.Writer, widths []int, left, middle, right string) {
	fmt.Fprintf(output, "  %s%s", dim, left)
	for column, width := range widths {
		fmt.Fprint(output, strings.Repeat("─", width+2))
		if column < len(widths)-1 {
			fmt.Fprint(output, middle)
		}
	}
	fmt.Fprintf(output, "%s%s\n", right, reset)
}

func padLeft(value string, width int) string {
	return strings.Repeat(" ", width-utf8.RuneCountInString(value)) + value
}

func padRight(value string, width int) string {
	return value + strings.Repeat(" ", width-utf8.RuneCountInString(value))
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
	fmt.Fprintf(output, "  %s▸%s %s", cyan, reset, prompt)
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

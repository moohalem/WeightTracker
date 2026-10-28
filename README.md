# WeightTracker

A small interactive command-line app for recording and reviewing weight measurements. Data is stored locally in SQLite.

## Features

- Record a person's name and weight in kilograms.
- View the 10 most recent measurements.
- Select a stored name and view that person's measurements.
- Display measurement results in a table.

## Requirements

- Go 1.27.3 or later.

## Run

From the project root, run:

```sh
go run ./cmd
```

On first run, the app creates `app.db` in the current working directory and initializes its schema.

## Menu

- `1` Add a measurement. Enter a name and a weight greater than zero.
- `2` View the 10 most recent measurements.
- `3` Choose a name from the names already recorded and view that person's measurements.
- `0` Exit.

Measurement lists are ordered from newest to oldest. Press Enter after viewing a list to return to the menu.

## Test

Run the Go package tests with:

```sh
go test ./...
```

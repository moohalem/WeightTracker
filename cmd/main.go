package main

import (
	"log"

	db "github.com/moohalem/WeightTracker/internal/db"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	// var name string
	// var weight float64
	// fmt.Println("Welcome to the Weight Tracker")
	// fmt.Println("Enter your name: ")
	// fmt.Scanln(&name)
	// fmt.Printf("Hi %s, enter your weight: ", name)
	// fmt.Scanln(&weight)
}

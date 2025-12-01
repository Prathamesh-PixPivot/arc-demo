package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify a command: 'up' or 'review-defaults'")
	}

	switch os.Args[1] {
	case "up":
		runMigrations()
	case "review-defaults":
		setReviewDefaults()
	default:
		log.Fatalf("Unknown command: %s. Use 'up' or 'review-defaults'", os.Args[1])
	}
}

func runMigrations() {
	RunMigrations()
}

func setReviewDefaults() {
	review()
}


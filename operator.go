package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: operator <sqlite_url> <postgres_url>")
		os.Exit(1)
	}

	sqliteURL := os.Args[1]
	postgresURL := os.Args[2]

	// Set environment variables
	os.Setenv("EXAMPLE_VAR1", sqliteURL)
	os.Setenv("EXAMPLE_VAR2", postgresURL)

	// Simulate performing tasks on the instances
	fmt.Printf("Performing tasks on SQLite instance at %s and Postgres instance at %s...\n", sqliteURL, postgresURL)

	// Output the variables for debugging
	fmt.Printf("EXAMPLE_VAR1: %s\n", os.Getenv("EXAMPLE_VAR1"))
	fmt.Printf("EXAMPLE_VAR2: %s\n", os.Getenv("EXAMPLE_VAR2"))
}

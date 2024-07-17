package main

import (
	"os"
	"testing"
)

func TestEnvironmentVariables(t *testing.T) {
	var1 := os.Getenv("EXAMPLE_VAR1")
	var2 := os.Getenv("EXAMPLE_VAR2")

	if var1 == "" {
		t.Fatalf("Expected EXAMPLE_VAR1 to be set, but it was empty")
	}
	if var2 == "" {
		t.Fatalf("Expected EXAMPLE_VAR2 to be set, but it was empty")
	}
}

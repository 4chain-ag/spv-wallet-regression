package utils

import (
	"fmt"
	"os"
	"regexp"
)

var (
	StdErr               = os.Stderr
	StdOut               = os.Stdout
	explicitHTTPURLRegex = regexp.MustCompile(`^https?://`)
)

// IsValidURL checks if a string is a valid URL with http/https.
func IsValidURL(rawURL string) bool {
	return explicitHTTPURLRegex.MatchString(rawURL)
}

// AddPrefixIfNeeded ensures the URL has an https:// prefix.
func AddPrefixIfNeeded(url string) string {
	if !IsValidURL(url) {
		return "https://" + url
	}
	return url
}

// GetEnv returns the value of an environment variable.
func GetEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is not set", key)
	}
	return value, nil
}

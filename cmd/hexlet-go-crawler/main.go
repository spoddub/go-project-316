package main

import (
	"code/crawler"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error:")
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/hexlet-go-crawler https://example.com")
		os.Exit(1)
	}

	rootURL := os.Args[1]

	report, err := crawler.Analyze(rootURL, crawler.Options{
		Client: http.DefaultClient,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package brave

import (
	"context"
	"os"
	"testing"
)

func TestSearch(t *testing.T) {
	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		t.Skip("BRAVE_API_KEY not set")
	}

	client := New(apiKey)
	results, err := client.Search(context.Background(), "barbados real estate")

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected results, got none")
	}

	for _, r := range results {
		t.Logf("URL: %s", r.URL)
	}

}

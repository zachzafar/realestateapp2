package scraper

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/zachzafar/realestate-app-2/internal/brave"
)

func TestScraper(t *testing.T) {
	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		t.Skip("BRAVE_API_KEY not set")
	}

	client := brave.New(apiKey)
	results, err := client.Search(context.Background(), "barbados real estate")
	var wg sync.WaitGroup

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	for _, r := range results {

		scraper, err := New(r.URL)

		if err != nil {
			t.Fatalf("scraper failed with error: %v for url %s", err, r.URL)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			scraper.StartScrape()
		}()

	}
	wg.Wait()
	log.Printf("Scrape finished")
}

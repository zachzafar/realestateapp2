package main

import (
	"log"

	scraper "github.com/zachzafar/realestate-app-2/internal/scrapers"
)

func main() {
	scrapers := []scraper.Scraper{scraper.Terra{}, scraper.TerraLuxury{}}
	for _, scraper := range scrapers {
		properties, err := scraper.Scrape()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("scraped %d properties", len(properties))
	}

}

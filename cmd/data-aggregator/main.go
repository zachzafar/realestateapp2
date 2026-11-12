package main

import (
	"log"

	scraper "github.com/zachzafar/realestate-app-2/internal/scrapers"
)

func main() {
	scraper := scraper.TerraLuxury{}
	properties, err := scraper.Scrape()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("scraped %d properties", len(properties))
}

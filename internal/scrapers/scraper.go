package scraper

import (
	"github.com/zachzafar/realestate-app-2/internal/property"
)

type Scraper interface {
	Scrape() (error, []property.Property)
}

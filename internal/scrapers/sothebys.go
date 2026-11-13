package scraper

import (
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/zachzafar/realestate-app-2/internal/property"
)

type Sothebys struct {
}

func (s Sothebys) Scrape() ([]property.Property, error) {
	_ = os.RemoveAll("path")
	var properties []property.Property

	u := launcher.New().UserDataDir("path").Headless(false).MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := stealth.MustPage(browser)

	page = page.Timeout(300 * time.Second)

	pageNumber := 1
	var links []string
	for {
		url := fmt.Sprintf("https://www.barbadossothebysrealty.com/sales?page=%v", pageNumber)

		if err := page.Navigate(url); err != nil {
			return properties, fmt.Errorf("navigate: %w", err)
		}

		if _, err := page.Timeout(100*time.Millisecond).ElementR(`p.mt-16.text-center.text-gray-500`, `No properties found`); err == nil {
			break
		}

		// page.MustWaitLoad()
		// page.MustWaitStable()

		els, err := page.Elements(`a[href^="/sales/"]`)

		if err != nil {
			return properties, fmt.Errorf("elements: %w", err)
		}

		for _, a := range els {
			href, err := a.Attribute("href")
			if err != nil || href == nil || *href == "" {
				continue
			}

			links = append(links, *href)

		}

		pageNumber += 1
	}

	fmt.Print(links)

	return properties, nil
}

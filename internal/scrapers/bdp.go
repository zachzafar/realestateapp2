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

// Barbados dream properties scraper

type BDP struct{}

func (b BDP) Scrape() ([]property.Property, error) {
	_ = os.RemoveAll("path")

	var properties []property.Property

	u := launcher.New().UserDataDir("path").Headless(false).MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := stealth.MustPage(browser)

	page = page.Timeout(60 * time.Second)

	url := "https://www.barbadosdreamproperties.com/real-estate-sales/search"

	if err := page.Navigate(url); err != nil {
		return properties, fmt.Errorf("navigate: %w", err)
	}

	page.MustWaitLoad()
	page.MustWaitStable()

	var links []string
	prevCount := 0
	for {
		page.MustEval(`() => window.scrollTo(0, document.body.scrollHeight)`)

		wait := page.WaitRequestIdle(2*time.Second, nil, nil, nil)
		wait()
		property_cards, _ := page.Elements(`article.pl-article`)

		if len(property_cards) == prevCount {
			break
		}

		prevCount = len(property_cards)
	}

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

	fmt.Print(links)

	return properties, nil

}

package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/smegg99/human"

	"github.com/zachzafar/realestate-app-2/internal/property"
)

type Terra struct {
}

func (t Terra) Scrape() ([]property.Property, error) {
	_ = os.RemoveAll("path")
	var properties []property.Property
	countries := []string{"barbados", "saint-lucia", "grenada", "trinidad"}

	u := launcher.New().
		UserDataDir("path").
		Headless(false).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := stealth.MustPage(browser)

	page = page.Timeout(60 * time.Second)

	// 3. Navigate
	// cursor := human.New(page, human.Casual())

	for _, country := range countries {
		url := fmt.Sprintf("https://terracaribbean.com/%s/properties/?limit=10000", country)
		// doc, err := fetchPageWithAllListings(url)

		if err := page.Navigate(url); err != nil {
			return properties, fmt.Errorf("navigate: %w", err)
		}

		page.MustWaitLoad()
		page.MustWaitStable()

		els, err := page.Elements(`.ip-search-main-result-item a`)

		if err != nil {
			return properties, fmt.Errorf("elements: %w", err)
		}

		var links []string

		for _, a := range els {
			href, err := a.Attribute("href")
			if err != nil || href == nil {
				continue
			}

			links = append(links, *href)
		}

		for _, link := range links {
			p := page.Timeout(60 * time.Second)

			p.Navigate(link)
			time.Sleep(time.Duration(rand.Intn(500)+300) * time.Millisecond)
		}
		fmt.Print(links)
		time.Sleep(time.Duration(10 * time.Second))
	}
	return properties, nil
}

type TerraLuxury struct {
}

func (t TerraLuxury) Scrape() ([]property.Property, error) {

	properties := []property.Property{}
	_ = os.RemoveAll("path")

	u := launcher.New().
		UserDataDir("path").
		Headless(false).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := stealth.MustPage(browser)

	p := page.Timeout(30 * time.Second)

	// 3. Navigate
	cursor := human.New(p, human.Casual())

	// doc, err := fetchPageWithAllListings(url)

	if err := p.Navigate("https://terraluxury.com/our-properties/"); err != nil {
		return properties, fmt.Errorf("navigate: %w", err)
	}

	const selectorToWaitFor = `.ip-list-view-all a`
	el, err := p.Element(selectorToWaitFor)
	if err != nil {
		return properties, fmt.Errorf("element not found: %w", err)
	}

	el.MustScrollIntoView()

	time.Sleep(time.Duration(rand.Intn(500)+300) * time.Millisecond)

	if err := cursor.Click(el); err != nil {
		return properties, fmt.Errorf("click failed: %w", err)
	}

	// get all 'a' tags
	els, err := p.Elements(`.ip-search-result-item-inner a`)

	if err != nil {
		return properties, fmt.Errorf("elements: %w", err)
	}

	var links []string

	for _, a := range els {
		href, err := a.Attribute("href")
		if err != nil || href == nil {
			continue
		}

		links = append(links, *href)
	}

	for _, link := range links {

		// if strings.Contains(link, "property-enquiry") {
		// 	continue
		// }
		p := page.Timeout(60 * time.Second)

		p.Navigate(link)

		if err := p.WaitLoad(); err != nil {
			return properties, err
		}

		time.Sleep((time.Duration(rand.Intn(4000)+2000) * time.Millisecond))

		info, _ := p.Info()
		log.Printf("link=%s url=%s title=%q", link, info.URL, info.Title)
		// Get name
		var name string
		if ok, h1, _ := p.Has("h1"); ok {
			name, _ = h1.Text()
			name = strings.TrimSpace(name)
		}
		if name == "" {
			log.Printf("no h1 on %s (title %q), skipping", link, info.Title)
			continue
		}

		// Get images
		var images []string
		imgEls, err := page.Elements(`a[href*="cloudfront.net"]`)

		if err != nil {
			return properties, err
		}

		for _, a := range imgEls {
			if href, err := a.Attribute("href"); err == nil && href != nil {
				images = append(images, *href)
			}
		}

		// Get Description
		var description string
		if el, err := page.Element(`meta[name="description"]`); err == nil {
			if c, _ := el.Attribute("content"); c != nil {
				description = *c
			}
		}

		// Features / Agent Insights (XPath)
		features := listByXPath(page, `//h2[normalize-space()='Features']/following-sibling::ul[1]/li`)
		// insights := listByXPath(page, `//h2[normalize-space()='Agent Insights']/following-sibling::ul[1]/li`)

		// Price and specs: XPath, or rod's built-in regex matcher
		priceStr := textByXPath(page, `//*[contains(text(),'Price:')]`)
		bedroomsStr := textByXPath(page, `//*[contains(text(),'Bedrooms')]`)
		landAreaStr := textByXPath(page, `//*[contains(text(),'Land Area')]`)

		priceStr = regexp.MustCompile(`[\d,]+(?:\.\d+)?`).FindString(priceStr)
		bedroomsStr = regexp.MustCompile(`\d+`).FindString(bedroomsStr)
		landAreaStr = regexp.MustCompile(`[\d,]+(?:\.\d+)?`).FindString(landAreaStr)

		bedrooms, _ := strconv.Atoi(bedroomsStr)
		landArea, _ := strconv.ParseFloat(strings.ReplaceAll(landAreaStr, ",", ""), 32)
		price, _ := strconv.ParseFloat(strings.ReplaceAll(priceStr, ",", ""), 32)

		property := property.Property{
			Name:        name,
			Images:      images,
			Description: description,
			Price:       float32(price),
			Features:    features,
			Bedrooms:    bedrooms,
			Landarea:    float32(landArea),
		}

		propertyJSON, err := json.MarshalIndent(property, "", "  ")

		if err != nil {
			return properties, err
		}

		fmt.Println(string(propertyJSON))

		properties = append(properties, property)

	}

	return properties, nil
}

func listByXPath(page *rod.Page, xpath string) []string {
	els, err := page.ElementsX(xpath)
	if err != nil {
		return nil
	}
	var out []string
	for _, li := range els {
		if t, err := li.Text(); err == nil {
			out = append(out, strings.TrimSpace(t))
		}
	}
	return out
}

func textByXPath(page *rod.Page, xpath string) string {
	els, err := page.ElementsX(xpath)
	if err != nil || len(els) == 0 {
		return ""
	}
	t, _ := els[len(els)-1].Text() // last match is the deepest element
	return strings.TrimSpace(t)
}

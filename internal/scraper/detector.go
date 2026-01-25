package scraper

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PageSignals struct {
	HasPrice         bool
	HasBedrooms      bool
	HasBathrooms     bool
	HasSquareFootage bool
	HasImageGallery  bool
	HasAddress       bool
	PropertyCards    int // multiple = listing page
	PriceCount       int // multiple prices = listing page
	Title            string
}

var (
	pricePattern   = regexp.MustCompile(`(?i)(\$|BBD|USD|US\$)\s?[\d,]+`)
	bedroomPattern = regexp.MustCompile(`(?i)(\d+)\s*(bed|bedroom|br|bd)s?`)
	bathPattern    = regexp.MustCompile(`(?i)(\d+)\s*(bath|bathroom|ba)s?`)
	sqftPattern    = regexp.MustCompile(`(?i)([\d,]+)\s*(sq\.?\s*ft|sqft|square feet|sq\s*m)`)
	addressPattern = regexp.MustCompile(`(?i)(st\.\s*\w+|christ church|st\.\s*james|bridgetown|parish|barbados)`)
)

func DetectPageSignals(doc *goquery.Document) PageSignals {
	bodyText := doc.Find("body").Text()

	signals := PageSignals{}

	// Text-based signals
	signals.HasPrice = pricePattern.MatchString(bodyText)
	signals.HasBedrooms = bedroomPattern.MatchString(bodyText)
	signals.HasBathrooms = bathPattern.MatchString(bodyText)
	signals.HasSquareFootage = sqftPattern.MatchString(bodyText)
	signals.HasAddress = addressPattern.MatchString(bodyText)

	// Count prices (many = listing page, one = detail page)
	signals.PriceCount = len(pricePattern.FindAllString(bodyText, -1))

	// Check for image gallery
	signals.HasImageGallery = doc.Find(
		".gallery, .carousel, .slider, .lightbox, "+
			"[class*='gallery'], [class*='carousel'], [class*='slider'], "+
			".property-images, .listing-images",
	).Length() > 0

	// Count property card elements (indicates listing page)
	signals.PropertyCards = doc.Find(
		".property-card, .listing-card, .property-item, .listing-item, " +
			"[class*='property-card'], [class*='listing-card'], " +
			".search-result, .property-result",
	).Length()

	// Get the main title
	signals.Title = strings.TrimSpace(doc.Find("h1").First().Text())

	return signals
}

func IsPropertyDetailPage(doc *goquery.Document, url string) bool {
	lower := strings.ToLower(url)

	// Definite non-property pages
	skipList := []string{
		"/about", "/contact", "/terms", "/privacy", "/team",
		"/blog", "/news", "/faq", "/login", "/register", "/search",
		"/agents", "/careers", "/sitemap",
	}
	for _, skip := range skipList {
		if strings.Contains(lower, skip) {
			return false
		}
	}

	signals := DetectPageSignals(doc)

	// Listing page indicators (NOT a detail page)
	if signals.PropertyCards > 3 {
		return false
	}
	if signals.PriceCount > 5 {
		return false // too many prices = listing page
	}

	// Detail page scoring
	score := 0

	// URL patterns (strong signals)
	if regexp.MustCompile(`(?i)/property/`).MatchString(url) {
		score += 3
	}
	if regexp.MustCompile(`(?i)/listing/\d+`).MatchString(url) {
		score += 3
	}
	if regexp.MustCompile(`(?i)[?&](id|property_id)=\d+`).MatchString(url) {
		score += 2
	}

	// Content signals
	if signals.HasPrice && signals.PriceCount <= 3 {
		score += 2
	}
	if signals.HasBedrooms {
		score++
	}
	if signals.HasBathrooms {
		score++
	}
	if signals.HasSquareFootage {
		score++
	}
	if signals.HasImageGallery {
		score += 2
	}
	if signals.HasAddress {
		score++
	}
	if signals.Title != "" && len(signals.Title) > 10 {
		score++
	}

	return score >= 4
}

package scraper

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Scraper struct {
	base_url *url.URL
	links    map[string]bool
	pd_links []string
	mu       sync.Mutex
	wg       sync.WaitGroup
}

func (s *Scraper) LinkSeen(link string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[link] = true
}

func (s *Scraper) AddLink(link string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[link] = false
}

func (s *Scraper) CheckLink(link string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, exists := s.links[link]

	if !exists {
		return false
	}

	return value
}

func New(url_string string) (*Scraper, error) {
	base, err := url.Parse(url_string)

	if err != nil {
		return nil, err
	}
	log.Printf("Created scraper for %s", url_string)
	return &Scraper{
		base_url: base,
		links:    make(map[string]bool),
	}, nil
}

func (s *Scraper) StartScrape() {
	base := s.base_url.String()
	s.AddLink(base)
	s.wg.Add(1)
	s.ExtractLinks(base)
	s.wg.Wait()
	for _, result := range s.pd_links {
		log.Printf("Detail Page: %s", result)
	}
}

func (s *Scraper) ExtractLinks(url_string string) {
	defer s.wg.Done()
	resp, err := http.Get(url_string)
	//log.Printf("exploring link %s", url_string)
	s.LinkSeen(url_string)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return
	}

	if value := IsPropertyDetailPage(doc, url_string); value {
		s.mu.Lock()
		s.pd_links = append(s.pd_links, url_string)
		s.mu.Unlock()
	}

	doc.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")

		if !exists || href == "" {
			return
		}

		// Skip non-http links
		if strings.HasPrefix(href, "javascript:") ||
			strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "#") {
			return
		}

		//resolve relative URLs
		parsed, err := url.Parse(href)

		if err != nil {
			return
		}

		resolved := s.base_url.ResolveReference(parsed)

		if resolved.Host != s.base_url.Host {
			return
		}
		resolvedStr := resolved.String()
		if seen := s.CheckLink(resolvedStr); !seen {
			//log.Printf("Found %s", resolved)
			s.AddLink(resolvedStr)
			s.wg.Add(1)
			go s.ExtractLinks(resolvedStr)
		}

	})
}

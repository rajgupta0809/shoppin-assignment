package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const RequestPerSec = 5

var productPatterns = []string{
	`/product/`,
	`/p/`,
	`/item/`,
	`/sku/`,
	`/detail/`,
}

var categoryPatterns = []string{
	"/category/",
	"/c/",
	"/collections/",
	"/gaming/",
}

type DomainList struct {
	Domains []string `json:"domains"`
}

func LoadDomains(filename string) ([]string, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var domainList DomainList
	if err := json.Unmarshal(file, &domainList); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return domainList.Domains, nil
}

func isCategoryURL(url string) bool {
	for _, pattern := range categoryPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}

func ExtractCategoryLinks(homepage string) ([]string, error) {
	resp, err := http.Get(homepage)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch page: status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var categoryLinks []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && isCategoryURL(href) {
			categoryLinks = append(categoryLinks, href)
		}
	})

	return categoryLinks, nil
}

func extractProductLinks(html string) []string {
	var links []string
	for _, pattern := range productPatterns {
		if strings.Contains(html, pattern) {
			links = append(links, pattern)
		}
	}
	return links
}

func isProductURL(url string) bool {
	for _, pattern := range productPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}

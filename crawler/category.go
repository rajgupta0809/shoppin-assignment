package crawler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var categoryPatterns = []string{
	"/category/",
	"/c/",
	"/collections/",
	"/gaming/",
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

func isCategoryURL(url string) bool {
	for _, pattern := range categoryPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}

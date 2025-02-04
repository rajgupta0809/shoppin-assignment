package crawler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

type ProductData struct {
	Domain      string   `json:"domain"`
	ProductURLs []string `json:"product_urls"`
}

func FetchAndParseHTML(pageURL string, wg *sync.WaitGroup, ch chan<- string, limiter <-chan time.Time) {
	defer wg.Done()

	<-limiter

	resp, err := http.Get(pageURL)
	if err != nil {
		fmt.Println("Failed to fetch page:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to fetch page, status code:", resp.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Failed to parse HTML:", err)
		return
	}

	doc.Find("a").Each(func(index int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if exists && isProductURL(href) {
			ch <- href // Send product URL to channel
		}
	})
}

func FetchJavaScriptRenderedPage(pageURL string, wg *sync.WaitGroup, ch chan<- string, limiter <-chan time.Time) {
	defer wg.Done()

	<-limiter

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		fmt.Println("Failed to load page:", err)
		return
	}

	productLinks := extractProductLinks(htmlContent)
	for _, link := range productLinks {
		fmt.Println("links")
		ch <- link
	}
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

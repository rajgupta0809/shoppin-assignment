package main

import (
	"assignment/crawler"
	"assignment/storage"
	"assignment/utils"
	"fmt"
	"log"
	"sync"
	"time"
)

const requestPerSec = 5

func main() {
	domains, err := utils.LoadDomains("domains.json")
	if err != nil {
		log.Fatalf("Failed to load domains: %v", err)
	}

	var wg sync.WaitGroup
	limiter := time.NewTicker(time.Second / requestPerSec).C

	ch := make(chan string)

	for _, domain := range domains {
		wg.Add(1)

		go func(domain string) {
			defer wg.Done()

			sitemaps, err := crawler.FetchRobotsTXT(domain)
			if err != nil {
				log.Printf("Error fetching robots.txt for %s: %v", domain, err)
				return
			}

			if len(sitemaps) > 0 {
				var innerWG sync.WaitGroup
				for _, sitemapURL := range sitemaps {
					innerWG.Add(1)
					go crawler.FetchAndParseSitemap(sitemapURL, &innerWG, ch, limiter)
				}

				innerWG.Wait()
			} else {
				categoryLinks, err := crawler.ExtractCategoryLinks(domain)
				if err != nil {
					log.Printf("Error extracting category links from %s: %v", domain, err)
					return
				}

				for _, categoryURL := range categoryLinks {
					crawler.FetchAndParseHTML(categoryURL, &wg, ch, limiter)
				}

				for _, categoryURL := range categoryLinks {
					crawler.FetchJavaScriptRenderedPage(categoryURL, &wg, ch, limiter)
				}
			}

		}(domain)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var productData []storage.ProductData

	for url := range ch {
		productData = append(productData, storage.ProductData{
			Domain:      "snapdeal.com",
			ProductURLs: []string{url},
		})
	}

	if err := storage.SaveToJSON(productData, "output.json"); err != nil {
		log.Fatalf("Error saving to JSON: %v", err)
	}

	if err := storage.SaveToCSV(productData, "output.csv"); err != nil {
		log.Fatalf("Error saving to CSV: %v", err)
	}

	fmt.Println("Product URLs have been successfully saved.")
}

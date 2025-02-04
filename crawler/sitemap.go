package crawler

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc string `xml:"loc"`
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Urls    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

var productPatterns = []string{
	`/product/`,
	`/p/`,
	`/item/`,
	`/sku/`,
	`/detail/`,
}

func FetchAndParseSitemap(sitemapURL string, wg *sync.WaitGroup, ch chan<- string, limiter <-chan time.Time) {
	defer wg.Done()

	<-limiter

	resp, err := http.Get(sitemapURL)
	if err != nil {
		fmt.Println("Failed to fetch sitemap:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to fetch sitemap, status code:", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return
	}

	var index SitemapIndex
	err = xml.Unmarshal(body, &index)
	if err == nil && len(index.Sitemaps) > 0 {
		fmt.Println("Detected sitemap index. Fetching nested sitemaps...")
		var innerWG sync.WaitGroup
		for _, sitemap := range index.Sitemaps {
			innerWG.Add(1)
			go FetchAndParseSitemap(sitemap.Loc, &innerWG, ch, limiter)
		}
		wg.Wait()
		return
	}

	var urlset URLSet
	err = xml.Unmarshal(body, &urlset)
	if err == nil && len(urlset.Urls) > 0 {
		fmt.Println("Detected product URLs sitemap.")
		for _, url := range urlset.Urls {
			if len(ch) > 10 {
				return
			}
			if isProductURL(url.Loc) {
				ch <- url.Loc
			}
		}
		return
	}
	fmt.Println("Unknown sitemap format:", sitemapURL)
}

func FetchRobotsTXT(domain string) ([]string, error) {
	url := domain + "/robots.txt"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch robots.txt, status code: %d", resp.StatusCode)
	}

	var sitemaps []string
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Sitemap:") {
			sitemapURL := strings.TrimSpace(strings.TrimPrefix(line, "Sitemap: "))
			sitemaps = append(sitemaps, sitemapURL)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading robots.txt: %w", err)
	}

	return sitemaps, nil
}

func FetchSitemap(domain string) (string, error) {
	url := domain + "/sitemap.xml"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch sitemap, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

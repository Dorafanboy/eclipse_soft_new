package requester

import (
	"eclipse/internal/logger"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ScraperConfig struct {
	MaxURLs              int
	DelayBetweenRequests time.Duration
	OutputFile           string
	SearchTerms          []string
}

func getRandomSearchTerm() string {
	searchTerms := []string{
		"digital art",
		"abstract art",
		"nft collection",
		"crypto art",
		"generative art",
		"pixel art",
		"3d art",
		"modern art",
		"surreal art",
		"minimalist art",
		"fantasy art",
		"space art",
		"cyberpunk art",
		"geometric art",
		"futuristic art",
	}

	rand.Seed(time.Now().UnixNano())
	return searchTerms[rand.Intn(len(searchTerms))]
}

func GetOneRandomImage(client http.Client) string {
	searchTerm := getRandomSearchTerm()
	logger.Info("Searching images for: %s", searchTerm)

	urls := scrapeGoogleImages(client, searchTerm)

	if len(urls) > 0 {
		rand.Seed(time.Now().UnixNano())
		randomUrl := urls[rand.Intn(len(urls))]
		return randomUrl
	}

	logger.Info("No images found, using fallback URL")
	return "https://ssl.gstatic.com/gb/images/bar/al-icon.png"
}

func scrapeGoogleImages(client http.Client, searchTerm string) []string {
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&tbm=isch",
		strings.ReplaceAll(searchTerm, " ", "+"))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		logger.Error("Error creating request: %v", err)
		return nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error making request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response: %v", err)
		return nil
	}

	imageURLs := extractImageURLs(string(body))

	return imageURLs
}

func extractImageURLs(html string) []string {
	imgRegex := regexp.MustCompile(`https?:\/\/[^"']*?(?:png|jpg|jpeg|gif|webp)`)
	matches := imgRegex.FindAllString(html, -1)

	uniqueURLs := make(map[string]bool)
	var result []string

	for _, url := range matches {
		if !uniqueURLs[url] {
			uniqueURLs[url] = true
			result = append(result, url)
		}
	}

	return result
}

package relay

import "net/http"

func SetRequestHeaders(req *http.Request, refererURL string) {
	headers := map[string]string{
		"accept":             "application/json, text/plain, */*",
		"accept-language":    "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		"cache-control":      "no-cache",
		"content-type":       "application/json",
		"origin":             "https://relay.link",
		"pragma":             "no-cache",
		"priority":           "u=1, i",
		"referer":            refererURL,
		"sec-ch-ua":          `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

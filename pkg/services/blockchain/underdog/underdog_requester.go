package underdog

import (
	"bytes"
	"eclipse/internal/logger"
	"encoding/json"
	"io"
	"net/http"
)

type CollectionData struct {
	Account      string `json:"account"`
	Name         string `json:"name"`
	Image        string `json:"image"`
	Description  string `json:"description"`
	ExternalUrl  string `json:"externalUrl"`
	Soulbound    bool   `json:"soulbound"`
	Transferable bool   `json:"transferable"`
	Burnable     bool   `json:"burnable"`
}

func CreateCollection(client http.Client, data CollectionData) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error(err.Error())
	}

	req, err := http.NewRequest("POST", "https://eclipse.underdogprotocol.com/api/collections", bytes.NewReader(jsonData))
	if err != nil {
		logger.Error(err.Error())
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://eclipse.underdogprotocol.com")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://eclipse.underdogprotocol.com/collections/create")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err.Error())

	}
	
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err.Error())

	}

	type Response struct {
		Transaction string `json:"transaction"`
		Message     string `json:"message"`
	}

	var response Response
	if err := json.Unmarshal(bodyText, &response); err != nil {
		logger.Error(err.Error())

	}

	return response.Transaction
}

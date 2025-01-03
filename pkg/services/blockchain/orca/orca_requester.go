package orca

import (
	"bytes"
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"io"
	"net/url"
	"strconv"
)

type SwapQuoteParams struct {
	FromToken            string
	ToToken              string
	Amount               string
	IsLegacy             bool
	AmountIsInput        bool
	IncludeData          bool
	IncludeComputeBudget bool
	MaxTxSize            int
	WalletAddress        string
}

type RequestConfig struct {
	URL            string
	AcceptHeader   string
	AcceptLanguage string
	CacheControl   string
	Origin         string
	Pragma         string
	Priority       string
	Referer        string
	UserAgent      string
	SecChUa        string
	SecChUaMobile  string
	SecChPlatform  string
	SecFetchDest   string
	SecFetchMode   string
	SecFetchSite   string
}

type SwapResponse struct {
	Data struct {
		Request struct {
			AmountIsInput string `json:"amountIsInput"`
		} `json:"request"`
		Swap struct {
			InputAmount  string     `json:"inputAmount"`
			OutputAmount string     `json:"outputAmount"`
			Split        [][]SwapOp `json:"split"`
		} `json:"swap"`
	} `json:"data"`
}

type SwapOp struct {
	Pool   string `json:"pool"`
	Input  Token  `json:"input"`
	Output Token  `json:"output"`
}

type Token struct {
	Mint   string `json:"mint"`
	Amount string `json:"amount"`
}

type SwapInstructions struct {
	Data struct {
		Instructions []struct {
			ProgramID []byte        `json:"programId"`
			Accounts  []AccountMeta `json:"accounts"`
			Data      []byte        `json:"data"`
		} `json:"instructions"`
		LookupTableAccounts [][]byte `json:"lookupTableAccounts"`
		Signers             [][]byte `json:"signers"`
	} `json:"data"`
	Meta struct {
		Cursor    interface{} `json:"cursor"`
		Slot      int64       `json:"slot"`
		BlockTime int64       `json:"block_time"`
	} `json:"meta"`
}

type AccountMeta struct {
	Pubkey     []byte `json:"pubkey"`
	IsSigner   bool   `json:"isSigner"`
	IsWritable bool   `json:"isWritable"`
}

type SwapRequestBody struct {
	AmountIsInput bool `json:"amountIsInput"`
	Swap          struct {
		InputAmount  string     `json:"inputAmount"`
		OutputAmount string     `json:"outputAmount"`
		Split        [][]SwapOp `json:"split"`
	} `json:"swap"`
	Wallet   string `json:"wallet"`
	Slippage string `json:"slippage"`
}

func DefaultOrcaConfig(path string) RequestConfig {
	return RequestConfig{
		URL:            fmt.Sprintf("https://pools-api-eclipse.mainnet.orca.so/%s", path),
		AcceptHeader:   "application/json, text/plain, */*",
		AcceptLanguage: "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		CacheControl:   "no-cache",
		Origin:         "https://www.orca.so",
		Pragma:         "no-cache",
		Priority:       "u=1, i",
		Referer:        "https://www.orca.so/",
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		SecChUa:        `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		SecChUaMobile:  "?0",
		SecChPlatform:  `"Windows"`,
		SecFetchDest:   "empty",
		SecFetchMode:   "cors",
		SecFetchSite:   "same-site",
	}
}

func (c RequestConfig) getHeaders() http.Header {
	return http.Header{
		"accept":             {c.AcceptHeader},
		"accept-language":    {c.AcceptLanguage},
		"cache-control":      {c.CacheControl},
		"origin":             {c.Origin},
		"pragma":             {c.Pragma},
		"priority":           {c.Priority},
		"referer":            {c.Referer},
		"user-agent":         {c.UserAgent},
		"sec-ch-ua":          {c.SecChUa},
		"sec-ch-ua-mobile":   {c.SecChUaMobile},
		"sec-ch-ua-platform": {c.SecChPlatform},
		"sec-fetch-dest":     {c.SecFetchDest},
		"sec-fetch-mode":     {c.SecFetchMode},
		"sec-fetch-site":     {c.SecFetchSite},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"cache-control",
			"origin",
			"pragma",
			"priority",
			"referer",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"user-agent",
		},
	}
}

func GetOrcaSwapQuote(params SwapQuoteParams, proxy string) (*SwapResponse, error) {
	config := DefaultOrcaConfig("swap-quote")

	query := url.Values{}
	query.Add("from", params.FromToken)
	query.Add("to", params.ToToken)
	query.Add("amount", params.Amount)
	query.Add("isLegacy", strconv.FormatBool(params.IsLegacy))
	query.Add("amountIsInput", strconv.FormatBool(params.AmountIsInput))
	query.Add("includeData", strconv.FormatBool(params.IncludeData))
	query.Add("includeComputeBudget", strconv.FormatBool(params.IncludeComputeBudget))
	query.Add("maxTxSize", strconv.Itoa(params.MaxTxSize))
	query.Add("wallet", params.WalletAddress)

	fullURL := config.URL + "?" + query.Encode()

	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_110),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithProxyUrl(proxy),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header = config.getHeaders()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("неуспешный статус ответа: %d, тело: %s", resp.StatusCode, string(body))
	}

	var response SwapResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &response, nil
}

func PrepareSwapInstructions(swapResponse *SwapResponse, walletAddress string, proxy string) (*SwapInstructions, error) {
	config := DefaultOrcaConfig("swap-prepare-instructions")

	requestBody := SwapRequestBody{
		AmountIsInput: true,
		Swap:          swapResponse.Data.Swap,
		Wallet:        walletAddress,
		Slippage:      "0.005",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_110),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithProxyUrl(proxy),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента: %v", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://pools-api-eclipse.mainnet.orca.so/swap-prepare-instructions",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header = config.getHeaders()
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("неуспешный статус ответа: %d, тело: %s", resp.StatusCode, string(body))
	}

	var swapInstructions SwapInstructions
	if err := json.Unmarshal(body, &swapInstructions); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %v", err)
	}

	return &swapInstructions, nil
}

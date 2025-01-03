package solar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"io"
	"net/http"
)

type SolarSwapResponse struct {
	ID      string   `json:"id"`
	Success bool     `json:"success"`
	Version string   `json:"version"`
	Data    SwapData `json:"data"`
}

type SwapData struct {
	SwapType             string      `json:"swapType"`
	InputMint            string      `json:"inputMint"`
	InputAmount          string      `json:"inputAmount"`
	OutputMint           string      `json:"outputMint"`
	OutputAmount         string      `json:"outputAmount"`
	OtherAmountThreshold string      `json:"otherAmountThreshold"`
	SlippageBps          int         `json:"slippageBps"`
	PriceImpactPct       float64     `json:"priceImpactPct"`
	ReferrerAmount       string      `json:"referrerAmount"`
	RoutePlan            []RoutePlan `json:"routePlan"`
}

type RoutePlan struct {
	PoolID            string   `json:"poolId"`
	InputMint         string   `json:"inputMint"`
	OutputMint        string   `json:"outputMint"`
	FeeMint           string   `json:"feeMint"`
	FeeRate           int      `json:"feeRate"`
	FeeAmount         string   `json:"feeAmount"`
	RemainingAccounts []string `json:"remainingAccounts"`
}

type SwapParams struct {
	Amount    string
	FromToken solana.PublicKey
	ToToken   solana.PublicKey
}

func GetSolarSwapCompute(client http.Client, params SwapParams) (*SolarSwapResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.solarstudios.co/compute/swap-base-in?inputMint=%s&outputMint=%s&amount=%s&slippageBps=50&txVersion=V0", params.FromToken.String(), params.ToToken.String(), params.Amount), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	headers := map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		"Cache-Control":      "no-cache",
		"Connection":         "keep-alive",
		"Origin":             "https://eclipse.solarstudios.co",
		"Pragma":             "no-cache",
		"Referer":            "https://eclipse.solarstudios.co/",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"sec-ch-ua":          `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var swapResponse SolarSwapResponse
	if err := json.Unmarshal(bodyText, &swapResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &swapResponse, nil
}

type SwapRequest struct {
	Wallet                        string            `json:"wallet"`
	ComputeUnitPriceMicroLamports string            `json:"computeUnitPriceMicroLamports"`
	SwapResponse                  SolarSwapResponse `json:"swapResponse"`
	TxVersion                     string            `json:"txVersion"`
	WrapSol                       bool              `json:"wrapSol"`
	UnwrapSol                     bool              `json:"unwrapSol"`
	InputAccount                  string            `json:"inputAccount"`
}

type TransactionResponse struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Success bool   `json:"success"`
	Data    []struct {
		Transaction string `json:"transaction"`
	} `json:"data"`
}

func CreateSwapTransaction(client http.Client, wallet, outputAccount string, swapResponse *SolarSwapResponse) (*TransactionResponse, error) {
	swapRequest := SwapRequest{
		Wallet:                        wallet,
		ComputeUnitPriceMicroLamports: "200000",
		SwapResponse:                  *swapResponse,
		TxVersion:                     "LEGACY",
		WrapSol:                       false,
		UnwrapSol:                     true,
		InputAccount:                  outputAccount,
	}

	jsonData, err := json.Marshal(swapRequest)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST",
		"https://api.solarstudios.co/transaction/swap-base-in",
		bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	headers := map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		"Cache-Control":      "no-cache",
		"Connection":         "keep-alive",
		"Content-Type":       "application/json",
		"Origin":             "https://eclipse.solarstudios.co",
		"Pragma":             "no-cache",
		"Referer":            "https://eclipse.solarstudios.co/",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"sec-ch-ua":          `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var txResponse TransactionResponse
	if err := json.Unmarshal(bodyText, &txResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	if !txResponse.Success || len(txResponse.Data) == 0 {
		return nil, fmt.Errorf("invalid response from server")
	}

	return &txResponse, nil
}

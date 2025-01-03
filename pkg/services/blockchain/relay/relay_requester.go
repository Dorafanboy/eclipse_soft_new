package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RelayRequest struct {
	User                 string `json:"user"`
	OriginChainId        int    `json:"originChainId"`
	DestinationChainId   int    `json:"destinationChainId"`
	OriginCurrency       string `json:"originCurrency"`
	DestinationCurrency  string `json:"destinationCurrency"`
	Recipient            string `json:"recipient"`
	TradeType            string `json:"tradeType"`
	Amount               string `json:"amount"`
	Referrer             string `json:"referrer"`
	UseExternalLiquidity bool   `json:"useExternalLiquidity"`
}

type TransactionData struct {
	From                 string `json:"from"`
	To                   string `json:"to"`
	Data                 string `json:"data"`
	Value                string `json:"value"`
	ChainId              int    `json:"chainId"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
}

type RelayResponse struct {
	Steps []struct {
		Items []struct {
			Data TransactionData `json:"data"`
		} `json:"items"`
	} `json:"steps"`
}

func GetRelayData(client http.Client, request RelayRequest) (*TransactionData, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.relay.link/quote", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	refererURL := fmt.Sprintf(
		"https://relay.link/bridge/eclipse?toCurrency=%s&fromChainId=%d&fromCurrency=%s",
		request.DestinationCurrency,
		request.OriginChainId,
		request.OriginCurrency,
	)

	SetRequestHeaders(req, refererURL)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	var relayResponse RelayResponse
	if err := json.NewDecoder(resp.Body).Decode(&relayResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(relayResponse.Steps) == 0 || len(relayResponse.Steps[0].Items) == 0 {
		return nil, fmt.Errorf("no transaction data in response")
	}

	log.Println("Успешно получил данные для выполнение Relay бриджа")
	return &relayResponse.Steps[0].Items[0].Data, nil
}

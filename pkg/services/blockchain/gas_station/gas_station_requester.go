package gas_station

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type SwapRequest struct {
	User              string `json:"user"`
	SourceMint        string `json:"sourceMint"`
	Amount            int    `json:"amount"`
	SlippingTolerance int    `json:"slippingTolerance"`
}

type Response struct {
	Status       string `json:"status"`
	Transaction  string `json:"transaction"`
	Quote        Quote  `json:"quote"`
	MessageToken string `json:"messageToken"`
}

type Quote struct {
	EstimatedAmountIn      string      `json:"estimatedAmountIn"`
	EstimatedAmountOut     string      `json:"estimatedAmountOut"`
	EstimatedEndTickIndex  int         `json:"estimatedEndTickIndex"`
	EstimatedEndSqrtPrice  string      `json:"estimatedEndSqrtPrice"`
	EstimatedFeeAmount     string      `json:"estimatedFeeAmount"`
	TransferFee            TransferFee `json:"transferFee"`
	Amount                 string      `json:"amount"`
	AmountSpecifiedIsInput bool        `json:"amountSpecifiedIsInput"`
	AToB                   bool        `json:"aToB"`
	OtherAmountThreshold   string      `json:"otherAmountThreshold"`
	SqrtPriceLimit         string      `json:"sqrtPriceLimit"`
	TickArray0             string      `json:"tickArray0"`
	TickArray1             string      `json:"tickArray1"`
	TickArray2             string      `json:"tickArray2"`
}

type TransferFee struct {
	DeductingFromEstimatedAmountIn string `json:"deductingFromEstimatedAmountIn"`
	DeductedFromEstimatedAmountOut string `json:"deductedFromEstimatedAmountOut"`
}

func GetTxData(client http.Client, request SwapRequest) (*Response, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://octane-server-alpha.vercel.app/api/buildWhirlpoolsSwap", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://app.eclipse.xyz")
	req.Header.Set("referer", "https://app.eclipse.xyz/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

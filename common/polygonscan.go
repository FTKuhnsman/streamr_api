package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PolygonScanResponse struct to map the entire JSON response
type PolygonScanResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"` // ABI is a JSON-encoded string
}

// FetchContractABI makes an HTTP GET request to the PolygonScan API to fetch a contract's ABI
func FetchContractABI(contractAddress string) (string, error) {
	url := fmt.Sprintf("https://api.polygonscan.com/api?module=contract&action=getabi&address=%s", contractAddress)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var jsonResponse PolygonScanResponse
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", err
	}

	if jsonResponse.Status != "1" {
		return "", fmt.Errorf("error fetching contract ABI: %s", jsonResponse.Message)
	}

	return jsonResponse.Result, nil
}

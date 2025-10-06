package solana

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gagliardetto/solana-go"
	bin "github.com/gagliardetto/binary"
)

type JupiterClient struct {
	apiURL     string
	httpClient *http.Client
}

func NewJupiterClient(apiURL string) *JupiterClient {
	return &JupiterClient{
		apiURL:     apiURL,
		httpClient: &http.Client{},
	}
}

type QuoteRequest struct {
	InputMint            string
	OutputMint           string
	Amount               uint64
	SlippageBps          int
	OnlyDirectRoutes     bool
	AsLegacyTransaction  bool
}

type QuoteResponse struct {
	InputMint            string   `json:"inputMint"`
	InAmount             string   `json:"inAmount"`
	OutputMint           string   `json:"outputMint"`
	OutAmount            string   `json:"outAmount"`
	OtherAmountThreshold string   `json:"otherAmountThreshold"`
	SwapMode             string   `json:"swapMode"`
	SlippageBps          int      `json:"slippageBps"`
	PriceImpactPct       string   `json:"priceImpactPct"`
	RoutePlan            []Route  `json:"routePlan"`
}

type Route struct {
	SwapInfo SwapInfo `json:"swapInfo"`
	Percent  int      `json:"percent"`
}

type SwapInfo struct {
	AmmKey     string `json:"ammKey"`
	Label      string `json:"label"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

type SwapRequest struct {
	QuoteResponse        QuoteResponse `json:"quoteResponse"`
	UserPublicKey        string        `json:"userPublicKey"`
	WrapAndUnwrapSol     bool          `json:"wrapAndUnwrapSol"`
	ComputeUnitPriceMicroLamports *int64 `json:"computeUnitPriceMicroLamports,omitempty"`
}

type SwapResponse struct {
	SwapTransaction      string `json:"swapTransaction"`
	LastValidBlockHeight int64  `json:"lastValidBlockHeight"`
}

func (j *JupiterClient) GetQuote(ctx context.Context, req QuoteRequest) (*QuoteResponse, error) {
	params := url.Values{}
	params.Add("inputMint", req.InputMint)
	params.Add("outputMint", req.OutputMint)
	params.Add("amount", strconv.FormatUint(req.Amount, 10))
	params.Add("slippageBps", strconv.Itoa(req.SlippageBps))
	
	if req.OnlyDirectRoutes {
		params.Add("onlyDirectRoutes", "true")
	}
	if req.AsLegacyTransaction {
		params.Add("asLegacyTransaction", "true")
	}

	quoteURL := fmt.Sprintf("%s/quote?%s", j.apiURL, params.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", quoteURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := j.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jupiter API error: %s - %s", resp.Status, string(body))
	}

	var quote QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("failed to decode quote: %w", err)
	}

	return &quote, nil
}

func (j *JupiterClient) GetSwapTransaction(ctx context.Context, req SwapRequest) (*SwapResponse, error) {
	swapURL := fmt.Sprintf("%s/swap", j.apiURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal swap request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", swapURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jupiter swap API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var swapResp SwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&swapResp); err != nil {
		return nil, fmt.Errorf("failed to decode swap response: %w", err)
	}

	return &swapResp, nil
}

func (j *JupiterClient) ExecuteSwap(
	ctx context.Context,
	client *Client,
	inputMint, outputMint string,
	amount uint64,
	slippageBps int,
	priorityFeeMicroLamports int64,
) (solana.Signature, error) {
	// Get quote
	quote, err := j.GetQuote(ctx, QuoteRequest{
		InputMint:   inputMint,
		OutputMint:  outputMint,
		Amount:      amount,
		SlippageBps: slippageBps,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to get quote: %w", err)
	}

	// Get swap transaction
	swapResp, err := j.GetSwapTransaction(ctx, SwapRequest{
		QuoteResponse:                 *quote,
		UserPublicKey:                 client.wallet.PublicKey.String(),
		WrapAndUnwrapSol:              true,
		ComputeUnitPriceMicroLamports: &priorityFeeMicroLamports,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to get swap transaction: %w", err)
	}

	// Decode transaction
	txBytes, err := base64.StdEncoding.DecodeString(swapResp.SwapTransaction)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to decode transaction: %w", err)
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(client.wallet.PublicKey) {
			return &client.wallet.PrivateKey
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	sig, err := client.SendTransaction(ctx, tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig, nil
}

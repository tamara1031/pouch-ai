package proxy

import (
	"io"
	"net/http"
	"pouch-ai/internal/domain/proxy"
)

type ExecutionHandler struct {
	client *http.Client
}

func NewExecutionHandler() *ExecutionHandler {
	return &ExecutionHandler{
		client: &http.Client{},
	}
}

func (h *ExecutionHandler) Handle(req *proxy.Request) (*proxy.Response, error) {
	// 1. Prepare Request
	httpReq, err := req.Provider.PrepareHTTPRequest(req.Context, req.Model, req.RawBody)
	if err != nil {
		return nil, err
	}

	// 2. Execute
	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 3. Buffer Response (for token counting and returning)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 4. Calculate Usage
	inputUsage, _ := req.Provider.EstimateUsage(req.Model, req.RawBody)
	outputTokens, _ := req.Provider.ParseOutputUsage(req.Model, body, req.IsStream)

	pricing, _ := req.Provider.GetPricing(req.Model)
	inputCost := 0.0
	if inputUsage != nil {
		inputCost = inputUsage.TotalCost
	}
	outputCost := float64(outputTokens) / 1000.0 * pricing.Output

	return &proxy.Response{
		StatusCode:   resp.StatusCode,
		Body:         body,
		PromptTokens: inputUsage.InputTokens,
		OutputTokens: outputTokens,
		TotalCost:    inputCost + outputCost,
	}, nil
}

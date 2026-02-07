package infra

import (
	"bytes"
	"io"
	"net/http"
	"pouch-ai/backend/domain"
)

type ExecutionHandler struct {
	client *http.Client
	repo   domain.Repository
}

func NewExecutionHandler(repo domain.Repository) *ExecutionHandler {
	return &ExecutionHandler{
		client: &http.Client{},
		repo:   repo,
	}
}

func (h *ExecutionHandler) Handle(req *domain.Request) (*domain.Response, error) {
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

	// 3. For non-streaming, we still need to read it to count tokens reliably if the provider needs the full body.
	// But let's try to be consistent.
	if !req.IsStream {
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		inputUsage, _ := req.Provider.EstimateUsage(req.Model, req.RawBody)
		outputTokens, _ := req.Provider.ParseOutputUsage(req.Model, body, false)
		pricing, _ := req.Provider.GetPricing(req.Model)

		inputCost := 0.0
		if inputUsage != nil {
			inputCost = inputUsage.TotalCost
		}
		outputCost := float64(outputTokens) / 1000.0 * pricing.Output

		return &domain.Response{
			StatusCode:   resp.StatusCode,
			Header:       resp.Header,
			Body:         io.NopCloser(bytes.NewBuffer(body)),
			PromptTokens: inputUsage.InputTokens,
			OutputTokens: outputTokens,
			TotalCost:    inputCost + outputCost,
		}, nil
	}

	// 4. For streaming, we return the body directly but wrapped in a CountingReader.
	inputUsage, _ := req.Provider.EstimateUsage(req.Model, req.RawBody)
	inputCost := 0.0
	if inputUsage != nil {
		inputCost = inputUsage.TotalCost
	}

	// We wrap the body to count tokens and update usage when it's closed.
	var repo domain.Repository
	var keyID domain.ID
	if req.Key != nil {
		repo = h.repo
		keyID = req.Key.ID
	}

	// Create a wrapper that will update the database on Close()
	return &domain.Response{
		StatusCode:   resp.StatusCode,
		Header:       resp.Header,
		Body:         NewCountingReader(resp.Body, req.Provider, req.Model, repo, keyID, req.Context),
		PromptTokens: inputUsage.InputTokens,
		TotalCost:    inputCost,
	}, nil
}

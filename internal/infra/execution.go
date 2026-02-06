package infra

import (
	"bytes"
	"io"
	"net/http"
	"pouch-ai/internal/domain"
)

type ExecutionHandler struct {
	client *http.Client
}

func NewExecutionHandler() *ExecutionHandler {
	return &ExecutionHandler{
		client: &http.Client{},
	}
}

type readCloserWrapper struct {
	io.Reader
	closer io.Closer
}

func (r *readCloserWrapper) Close() error {
	return r.closer.Close()
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

		// Transform Response
		transformedReader, err := req.Provider.TransformResponse(bytes.NewBuffer(body), false)
		if err != nil {
			return nil, err
		}

		// If transformedReader is a ReadCloser, use it, otherwise wrap it
		var bodyReadCloser io.ReadCloser
		if rc, ok := transformedReader.(io.ReadCloser); ok {
			bodyReadCloser = rc
		} else {
			bodyReadCloser = io.NopCloser(transformedReader)
		}

		return &domain.Response{
			StatusCode:   resp.StatusCode,
			Header:       resp.Header,
			Body:         bodyReadCloser,
			PromptTokens: inputUsage.InputTokens,
			OutputTokens: outputTokens,
			TotalCost:    inputCost + outputCost,
		}, nil
	}

	// 4. For streaming
	inputUsage, _ := req.Provider.EstimateUsage(req.Model, req.RawBody)
	inputCost := 0.0
	if inputUsage != nil {
		inputCost = inputUsage.TotalCost
	}

	// Transform Response
	transformedReader, err := req.Provider.TransformResponse(resp.Body, true)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	// Wrap the transformed reader to ensure the original response body is closed
	wrappedBody := &readCloserWrapper{
		Reader: transformedReader,
		closer: resp.Body,
	}

	return &domain.Response{
		StatusCode:   resp.StatusCode,
		Header:       resp.Header,
		Body:         wrappedBody, // TODO: Wrap in CountingReader
		PromptTokens: inputUsage.InputTokens,
		TotalCost:    inputCost,
	}, nil
}

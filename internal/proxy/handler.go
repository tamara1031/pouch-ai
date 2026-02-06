package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"pouch-ai/internal/token"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	Token   *token.Counter
	Pricing *Pricing
	Target  *url.URL
	creds   *CredentialsManager

	// Optional callback to track usage (e.g. for app keys)
	UsageCallback func(c echo.Context, cost float64)
}

func NewHandler(t *token.Counter, p *Pricing, targetStr string, creds *CredentialsManager) (*Handler, error) {
	target, err := url.Parse(targetStr)
	if err != nil {
		return nil, err
	}
	return &Handler{
		Token:   t,
		Pricing: p,
		Target:  target,
		creds:   creds,
	}, nil
}

type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	MaxTokens int  `json:"max_tokens,omitempty"`
	Stream    bool `json:"stream,omitempty"`
}

// Handle is the Echo handler for proxying requests.
func (h *Handler) Handle(c echo.Context) error {
	// 1. Parse Request Body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read body")
	}
	// Restore body for proxy
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req OpenAIRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON body")
	}

	// 2. Estimate Cost
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096 // Default fallback safety
	}

	// Combine messages for input token count
	var inputBuilder strings.Builder
	for _, m := range req.Messages {
		inputBuilder.WriteString(m.Content)
	}
	inputStr := inputBuilder.String()

	price, err := h.Pricing.GetPrice(req.Model)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("Unsupported model: %s", req.Model))
	}

	inputTokens, err := h.Token.CountTokens(req.Model, inputStr)
	if err != nil {
		log.Printf("Token count error: %v", err)
		inputTokens = len(inputStr) / 4 // Very rough fallback
	}

	estimatedInputCost := float64(inputTokens) / 1000.0 * price.Input
	estimatedOutputCost := float64(req.MaxTokens) / 1000.0 * price.Output
	maxCost := estimatedInputCost + estimatedOutputCost

	// 3. Deposit (Reserve) - REMOVED (Handled by strict pre-check or just post-payment in this version)
	// if err := h.Budget.Reserve(maxCost); err != nil {
	// 	return echo.NewHTTPError(http.StatusPaymentRequired, fmt.Sprintf("Budget exceeded. Required: $%.4f", maxCost))
	// }

	// Prepare for Refund calculations
	startTime := time.Now()
	var actualOutputTokens int

	// Define Reverse Proxy
	proxy := httputil.NewSingleHostReverseProxy(h.Target)
	originalDirector := proxy.Director

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		// No refund needed as we charge after usage now.
		// Or if we implemented reservation, we would refund here.
		w.WriteHeader(http.StatusBadGateway)
	}

	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = h.Target.Host

		// Injection: Get Key
		key, err := h.creds.GetAPIKey("openai", "") // Using default inside
		if err == nil && key != "" {
			req.Header.Set("Authorization", "Bearer "+key)
		} else {
			log.Printf("Warning: No API Key found for openai")
		}
	}

	// Custom ResponseWriter to intercept usage
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode != 200 {
			// Refund handled in Handle flow if we return here?
			// Actually ModifyResponse is called before copying body.
			// If we return nil, it proceeds to copy body.
			// We should let it copy, but our RW will see the non-200.
			return nil
		}
		return nil
	}

	rw := &CostTrackingResponseWriter{
		ResponseWriter: c.Response().Writer,
		Mode:           "stream", // Default assume stream or detect
		Req:            req,
		Token:          h.Token,
	}
	if !req.Stream {
		rw.Mode = "json"
	}

	proxy.ServeHTTP(rw, c.Request())

	// Calculate actual usage
	rw.CalculateTokens()

	// 4. Refund Logic (Post-Process)
	actualOutputTokens = rw.OutputTokens

	// If request failed (non-200), we should refund fully?
	// How do we know result status?
	// We can't easily access the status code from ReverseProxy execution here unless RW captured it?
	// But RW wraps ResponseWriter.

	// Optimization: If rw.OutputTokens is 0 and it was a success, maybe it was really 0?
	// But usually there is some tokens.

	// Calculate final cost
	actualOutputCost := float64(actualOutputTokens) / 1000.0 * price.Output
	finalCost := estimatedInputCost + actualOutputCost

	// 4. Refund Logic - REMOVED
	// refundAmount := maxCost - finalCost
	// if refundAmount > 0 {
	// 	if err := h.Budget.Refund(refundAmount); err != nil {
	// 		log.Printf("Failed to refund: %v", err)
	// 	}
	// }
	refundAmount := 0.0

	log.Printf("Req: %s | max: $%.4f | actual: $%.4f | refund: $%.4f | duration: %v | out_tok: %d",
		req.Model, maxCost, finalCost, refundAmount, time.Since(startTime), actualOutputTokens)

	if h.UsageCallback != nil {
		h.UsageCallback(c, finalCost)
	}

	return nil
}

// CostTrackingResponseWriter wraps http.ResponseWriter to capture output tokens.
type CostTrackingResponseWriter struct {
	http.ResponseWriter
	Mode         string
	Req          OpenAIRequest
	Token        *token.Counter
	OutputTokens int
	BodyBuffer   bytes.Buffer
}

func (w *CostTrackingResponseWriter) Write(b []byte) (int, error) {
	w.BodyBuffer.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *CostTrackingResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *CostTrackingResponseWriter) CalculateTokens() {
	respStr := w.BodyBuffer.String()

	if w.Mode == "json" {
		var resp struct {
			Usage struct {
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		// Try to parse usage from standard response
		if err := json.Unmarshal([]byte(respStr), &resp); err == nil && resp.Usage.CompletionTokens > 0 {
			w.OutputTokens = resp.Usage.CompletionTokens
			return
		}

		// Fallback: Parse content from choices (if no usage field provided by some models)
		// ...

	} else {
		// Stream Parsing
		// Format: data: {"choices":[{"delta":{"content":"..."}}]} ...

		var fullContent strings.Builder
		lines := strings.Split(respStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			dataStr := strings.TrimPrefix(line, "data: ")
			if dataStr == "[DONE]" {
				continue
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(dataStr), &chunk); err == nil {
				if len(chunk.Choices) > 0 {
					fullContent.WriteString(chunk.Choices[0].Delta.Content)
				}
			}
		}

		finalText := fullContent.String()
		if len(finalText) > 0 {
			if count, err := w.Token.CountTokens(w.Req.Model, finalText); err == nil {
				w.OutputTokens = count
				return
			}
		}
	}

	// Final Fallback if parsing failed
	// Estimate 4 chars per token? length / 4
	w.OutputTokens = len(respStr) / 4
}

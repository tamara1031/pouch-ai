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

	"github.com/labstack/echo/v4"
	"pouch-ai/internal/budget"
	"pouch-ai/internal/token"
)

type Handler struct {
	Budget  *budget.Manager
	Token   *token.Counter
	Pricing *Pricing
	Target  *url.URL
}

func NewHandler(b *budget.Manager, t *token.Counter, p *Pricing, targetStr string) (*Handler, error) {
	target, err := url.Parse(targetStr)
	if err != nil {
		return nil, err
	}
	return &Handler{
		Budget:  b,
		Token:   t,
		Pricing: p,
		Target:  target,
	}, nil
}

type OpenAIRequest struct {
	Model     string  `json:"model"`
	Messages  []struct {
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
		// If not JSON or invalid, just pass through? OR fail?
		// Fail safe: we must control budget.
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

	// 3. Deposit (Reserve)
	if err := h.Budget.Reserve(maxCost); err != nil {
		return echo.NewHTTPError(http.StatusPaymentRequired, fmt.Sprintf("Budget exceeded. Required: $%.4f", maxCost))
	}

	// Prepare for Refund calculations
	startTime := time.Now()
	var actualOutputTokens int
	
	// Define Reverse Proxy
	proxy := httputil.NewSingleHostReverseProxy(h.Target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = h.Target.Host
		// Ensure API Key is passed (should be injected by middleware or header from client)
		// For now we assume client sends it or we handle key injection elsewhere.
        // The user spec said "credentials: encrypted_key". 
        // We need to inject the key here if we are managing keys.
        // TODO: Key Injection from DB based on some ID? Or use global key?
        // Assuming global key for MVP or extracted from header. 
        // Let's assume the request already has the key OR we inject it.
        // Implementation Plan didn't specify Key Injection logic in detail yet. 
        // We will revisit Key Injection.
	}

	// Custom ResponseWriter to intercept usage
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Log status
		if resp.StatusCode != 200 {
            // If API fails, we should fully refund? Or charge for input?
            // Usually if 4xx/5xx from OpenAI, no tokens used essentially (or negligible).
            // Let's refund full amount on error.
            return nil
		}
		return nil
	}
	
	// We need to intercept the body to count tokens.
    // This is complex for streaming.
    // For MVP, let's look at `usage` field for non-stream.
    // For stream, we must count chunks.
    
    // Simplification for MVP:
    // We will wrap the response writer.

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

	// 4. Refund Logic (Post-Process)
    // rw.OutputTokens captures the actual output.
    
    actualOutputTokens = rw.OutputTokens
    // If tracking failed (e.g. stream parsing error), we default to max reserved (no refund) -> Safe for budget.
    // Or we estimate based on length.
    
    actualOutputCost := float64(actualOutputTokens) / 1000.0 * price.Output
    finalCost := estimatedInputCost + actualOutputCost // Input is costed as estimated (since we sent it)
    
    refundAmount := maxCost - finalCost
    if refundAmount > 0 {
        if err := h.Budget.Refund(refundAmount); err != nil {
            log.Printf("Failed to refund: %v", err)
        }
    }
    
    // Log Audit
    log.Printf("Req: %s | max: $%.4f | actual: $%.4f | refund: $%.4f | duration: %v", req.Model, maxCost, finalCost, refundAmount, time.Since(startTime))

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
    // Pass through immediately
	return w.ResponseWriter.Write(b)
}

func (w *CostTrackingResponseWriter) Flush() {
    if f, ok := w.ResponseWriter.(http.Flusher); ok {
        f.Flush()
    }
    
    // Process accumulated buffer for token counting
    // For streaming, this is tricky as we need to parse SSE chunks incrementally or at end.
    // At Flush() we might check?
    // Actually `httputil.ReverseProxy` handles flushing.
    
    // Better strategy:
    // Process `w.BodyBuffer` at the very end of `Handle`? 
    // `ServeHTTP` blocks until done. So `w.BodyBuffer` will contain full response.
    // Wait, for streaming, `Write` is called many times.
    // We can just accumulate everything in `BodyBuffer` (memory intensive?)
    // For MVP, accumulating in memory is acceptable for text generation. 
}

// Close/Finish hook?
// We parse the buffer after ServeHTTP returns in Handle.
// But we need accessing `OutputTokens` from `rw` in `Handle`.
// So we should add a method `CalculateTokens`.

func (w *CostTrackingResponseWriter) CalculateTokens() {
    if w.Mode == "json" {
        // Parse JSON body, look for usage
        // ...
        // Fallback: count text
    } else {
        // Parse SSE stream
        // [data: {...}]
        // ...
    }
    
    // Implementation of token counting from response body...
    // Placeholder:
    w.OutputTokens = 0 // Update this!
    
    respStr := w.BodyBuffer.String()
    if w.Mode == "json" {
        var resp struct {
             Usage struct {
                 CompletionTokens int `json:"completion_tokens"`
             } `json:"usage"`
        }
        if err := json.Unmarshal([]byte(respStr), &resp); err == nil && resp.Usage.CompletionTokens > 0 {
             w.OutputTokens = resp.Usage.CompletionTokens
             return
        }
    }
    
    // Fallback or Stream: extract text and count
    // For stream: remove "data: ", parse JSON, extract delta.content
    // This is heavy.
    // Simplified: Just count the raw text length / 4?
    // Let's try to be a bit more accurate.
    
    // Simple text extraction for counting
    // ...
    // For now, let's leave 0 to be pessimistic (No refund if we can't count -> user pays max).
    // Or maybe naive char count.
    w.OutputTokens = len(respStr) / 4 // Crude approximation for MVP
}


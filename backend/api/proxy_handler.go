package api

import (
	"io"
	"net/http"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"

	"github.com/labstack/echo/v4"
)

const MaxBodySize = 10 * 1024 * 1024 // 10MB

type ProxyHandler struct {
	proxyService *service.ProxyService
	registry     domain.ProviderRegistry
}

func NewProxyHandler(ps *service.ProxyService, r domain.ProviderRegistry) *ProxyHandler {
	return &ProxyHandler{
		proxyService: ps,
		registry:     r,
	}
}

func (h *ProxyHandler) Proxy(c echo.Context) error {
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, MaxBodySize)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			return NewAPIError(c, http.StatusRequestEntityTooLarge, "Request body too large")
		}
		return BadRequest(c, "Failed to read body")
	}

	appKey, ok := c.Get("app_key").(*domain.Key)
	if !ok {
		return Unauthorized(c, "App Key not found")
	}

	// Identify Provider
	if appKey.Configuration == nil || appKey.Configuration.Provider.ID == "" {
		return BadRequest(c, "Provider not configured for this key")
	}
	provName := appKey.Configuration.Provider.ID
	prov, err := h.registry.Get(provName)
	if err != nil {
		return InternalError(c, "Provider not found")
	}

	model, isStream, err := prov.ParseRequest(body)
	if err != nil {
		return BadRequest(c, "Invalid request body")
	}

	req := &domain.Request{
		Context:  c.Request().Context(),
		Key:      appKey,
		Provider: prov,
		Model:    model,
		RawBody:  body,
		IsStream: isStream,
	}

	resp, err := h.proxyService.Execute(req)
	if err != nil {
		return BadGateway(c, err.Error())
	}
	defer resp.Body.Close()

	// Copy headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			c.Response().Header().Add(k, v)
		}
	}

	// Handle response
	if isStream {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
	}

	return c.Stream(resp.StatusCode, c.Response().Header().Get("Content-Type"), resp.Body)
}

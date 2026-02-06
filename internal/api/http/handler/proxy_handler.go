package handler

import (
	"io"
	"net/http"
	"pouch-ai/internal/app"
	"pouch-ai/internal/domain/key"
	"pouch-ai/internal/domain/provider"
	"pouch-ai/internal/domain/proxy"

	"github.com/labstack/echo/v4"
)

type ProxyHandler struct {
	proxyService *app.ProxyService
	registry     provider.Registry
}

func NewProxyHandler(ps *app.ProxyService, r provider.Registry) *ProxyHandler {
	return &ProxyHandler{
		proxyService: ps,
		registry:     r,
	}
}

func (h *ProxyHandler) Proxy(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read body")
	}

	appKey, ok := c.Get("app_key").(*key.Key)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "App Key not found")
	}

	// Identify Provider (MVP: always openai)
	prov, err := h.registry.Get("openai")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Provider not found")
	}

	model, isStream, err := prov.ParseRequest(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	req := &proxy.Request{
		Context:  c.Request().Context(),
		Key:      appKey,
		Provider: prov,
		Model:    model,
		RawBody:  body,
		IsStream: isStream,
	}

	resp, err := h.proxyService.Execute(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	// Handle response
	// If it was a stream, we should probably set some headers
	if isStream {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
	}

	return c.Blob(resp.StatusCode, c.Response().Header().Get("Content-Type"), resp.Body)
}

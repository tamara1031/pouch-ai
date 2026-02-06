package handler

import (
	"net/http"
	"pouch-ai/internal/app"
	"strconv"

	"github.com/labstack/echo/v4"
)

type KeyHandler struct {
	service *app.KeyService
}

func NewKeyHandler(s *app.KeyService) *KeyHandler {
	return &KeyHandler{service: s}
}

func (h *KeyHandler) ListKeys(c echo.Context) error {
	keys, err := h.service.ListKeys(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, keys)
}

func (h *KeyHandler) CreateKey(c echo.Context) error {
	var req struct {
		Name         string  `json:"name"`
		Provider     string  `json:"provider"`
		ExpiresAt    *int64  `json:"expires_at"`
		BudgetLimit  float64 `json:"budget_limit"`
		BudgetPeriod string  `json:"budget_period"`
		IsMock       bool    `json:"is_mock"`
		MockConfig   string  `json:"mock_config"`
		RateLimit    int     `json:"rate_limit"`
		RatePeriod   string  `json:"rate_period"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.Provider == "" {
		req.Provider = "openai"
	}

	raw, _, err := h.service.CreateKey(c.Request().Context(), req.Name, req.Provider, req.ExpiresAt, req.BudgetLimit, req.BudgetPeriod, req.IsMock, req.MockConfig, req.RateLimit, req.RatePeriod)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"key": raw,
	})
}

func (h *KeyHandler) UpdateKey(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var req struct {
		Name        string  `json:"name"`
		Provider    string  `json:"provider"`
		BudgetLimit float64 `json:"budget_limit"`
		IsMock      bool    `json:"is_mock"`
		MockConfig  string  `json:"mock_config"`
		RateLimit   int     `json:"rate_limit"`
		RatePeriod  string  `json:"rate_period"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = h.service.UpdateKey(c.Request().Context(), id, req.Name, req.Provider, req.BudgetLimit, req.IsMock, req.MockConfig, req.RateLimit, req.RatePeriod)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *KeyHandler) DeleteKey(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	if err := h.service.DeleteKey(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

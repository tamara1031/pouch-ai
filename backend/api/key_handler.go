package api

import (
	"errors"
	"fmt"
	"net/http"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

type KeyHandler struct {
	service *service.KeyService
}

func NewKeyHandler(s *service.KeyService) *KeyHandler {
	return &KeyHandler{service: s}
}

func (h *KeyHandler) ListKeys(c echo.Context) error {
	keys, err := h.service.ListKeys(c.Request().Context())
	if err != nil {
		return InternalError(c, err.Error())
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
		return BadRequest(c, err.Error())
	}

	input := service.CreateKeyInput{
		Name:         req.Name,
		Provider:     req.Provider,
		ExpiresAt:    req.ExpiresAt,
		BudgetLimit:  req.BudgetLimit,
		BudgetPeriod: req.BudgetPeriod,
		RateLimit:    req.RateLimit,
		RatePeriod:   req.RatePeriod,
		IsMock:       req.IsMock,
		MockConfig:   req.MockConfig,
	}

	raw, _, err := h.service.CreateKey(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrProviderNotFound) {
			return BadRequest(c, fmt.Sprintf("Invalid provider: %v", err))
		}
		return InternalError(c, err.Error())
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"key": raw,
	})
}

func (h *KeyHandler) UpdateKey(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return BadRequest(c, "Invalid ID")
	}

	var req struct {
		Name        string  `json:"name"`
		Provider    string  `json:"provider"`
		BudgetLimit float64 `json:"budget_limit"`
		IsMock      bool    `json:"is_mock"`
		MockConfig  string  `json:"mock_config"`
		RateLimit   int     `json:"rate_limit"`
		RatePeriod  string  `json:"rate_period"`
		ExpiresAt   *int64  `json:"expires_at"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, err.Error())
	}

	input := service.UpdateKeyInput{
		ID:          id,
		Name:        req.Name,
		Provider:    req.Provider,
		BudgetLimit: req.BudgetLimit,
		RateLimit:   req.RateLimit,
		RatePeriod:  req.RatePeriod,
		IsMock:      req.IsMock,
		MockConfig:  req.MockConfig,
		ExpiresAt:   req.ExpiresAt,
	}

	err = h.service.UpdateKey(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrProviderNotFound) {
			return BadRequest(c, fmt.Sprintf("Invalid provider: %v", err))
		}
		return InternalError(c, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *KeyHandler) DeleteKey(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return BadRequest(c, "Invalid ID")
	}

	if err := h.service.DeleteKey(c.Request().Context(), id); err != nil {
		return InternalError(c, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *KeyHandler) GetProviderUsage(c echo.Context) error {
	usage, err := h.service.GetProviderUsage(c.Request().Context())
	if err != nil {
		return InternalError(c, err.Error())
	}
	return c.JSON(http.StatusOK, usage)
}

func (h *KeyHandler) ListProviders(c echo.Context) error {
	providers, err := h.service.ListProviders(c.Request().Context())
	if err != nil {
		return InternalError(c, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{
		"providers": providers,
	})
}

package api

import (
	"net/http"
	"pouch-ai/internal/service"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
)

var keyNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)

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

	if len(req.Name) > 50 {
		return BadRequest(c, "Key name is too long (max 50 characters)")
	}
	if !keyNameRegex.MatchString(req.Name) {
		return BadRequest(c, "Key name contains invalid characters")
	}

	raw, _, err := h.service.CreateKey(c.Request().Context(), req.Name, req.Provider, req.ExpiresAt, req.BudgetLimit, req.BudgetPeriod, req.IsMock, req.MockConfig, req.RateLimit, req.RatePeriod)
	if err != nil {
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

	if len(req.Name) > 50 {
		return BadRequest(c, "Key name is too long (max 50 characters)")
	}
	if !keyNameRegex.MatchString(req.Name) {
		return BadRequest(c, "Key name contains invalid characters")
	}

	err = h.service.UpdateKey(c.Request().Context(), id, req.Name, req.Provider, req.BudgetLimit, req.IsMock, req.MockConfig, req.RateLimit, req.RatePeriod, req.ExpiresAt)
	if err != nil {
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

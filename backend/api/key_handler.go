package api

import (
	"errors"
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

type KeyResponse struct {
	ID            int64                    `json:"id"`
	Name          string                   `json:"name"`
	Prefix        string                   `json:"prefix"`
	ExpiresAt     *int64                   `json:"expires_at"`
	BudgetUsage   float64                  `json:"budget_usage"`
	CreatedAt     int64                    `json:"created_at"`
	Configuration *domain.KeyConfiguration `json:"configuration"`
}

func mapKeyToResponse(k *domain.Key) KeyResponse {
	resp := KeyResponse{
		ID:            int64(k.ID),
		Name:          k.Name,
		Prefix:        k.Prefix,
		BudgetUsage:   k.BudgetUsage,
		CreatedAt:     k.CreatedAt.Unix(),
		Configuration: k.Configuration,
	}

	if k.ExpiresAt != nil {
		ts := k.ExpiresAt.Unix()
		resp.ExpiresAt = &ts
	}
	return resp
}

func (h *KeyHandler) ListKeys(c echo.Context) error {
	keys, err := h.service.ListKeys(c.Request().Context())
	if err != nil {
		return InternalError(c, err.Error())
	}

	resp := make([]KeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = mapKeyToResponse(k)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *KeyHandler) CreateKey(c echo.Context) error {
	var req struct {
		Name        string                `json:"name"`
		Provider    string                `json:"provider"`
		ExpiresAt   *int64                `json:"expires_at"`
		MockConfig  string                `json:"mock_config"`
		Middlewares []domain.PluginConfig `json:"middlewares"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, err.Error())
	}

	input := service.CreateKeyInput{
		Name:        req.Name,
		Provider:    req.Provider,
		ExpiresAt:   req.ExpiresAt,
		MockConfig:  req.MockConfig,
		Middlewares: req.Middlewares,
	}

	raw, _, err := h.service.CreateKey(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrProviderNotFound) || domain.IsValidationError(err) {
			return BadRequest(c, err.Error())
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
		Name        string                `json:"name"`
		Provider    string                `json:"provider"`
		MockConfig  string                `json:"mock_config"`
		ExpiresAt   *int64                `json:"expires_at"`
		Middlewares []domain.PluginConfig `json:"middlewares"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, err.Error())
	}

	input := service.UpdateKeyInput{
		ID:          id,
		Name:        req.Name,
		Provider:    req.Provider,
		MockConfig:  req.MockConfig,
		ExpiresAt:   req.ExpiresAt,
		Middlewares: req.Middlewares,
	}

	err = h.service.UpdateKey(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrProviderNotFound) || domain.IsValidationError(err) {
			return BadRequest(c, err.Error())
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

func (h *KeyHandler) ListMiddlewares(c echo.Context) error {
	mws, err := h.service.ListMiddlewares(c.Request().Context())
	if err != nil {
		return InternalError(c, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{
		"middlewares": mws,
	})
}

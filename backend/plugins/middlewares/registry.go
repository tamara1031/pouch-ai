package middlewares

import (
	"pouch-ai/backend/domain"
)

type MiddlewareBuiltin struct {
	Info    domain.MiddlewareInfo
	Factory func(config map[string]any) domain.Middleware
}

func GetBuiltins() []MiddlewareBuiltin {
	return []MiddlewareBuiltin{
		{
			Info:    GetInfo(),
			Factory: NewRateLimitMiddleware,
		},
	}
}

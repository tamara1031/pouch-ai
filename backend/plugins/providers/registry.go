package providers

import (
	"pouch-ai/backend/domain"
)

func GetBuilders() []domain.ProviderBuilder {
	return []domain.ProviderBuilder{
		&OpenAIBuilder{},
		&MockBuilder{},
	}
}

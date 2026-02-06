package infra_test

import (
	"bytes"
	"fmt"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/infra"
	"testing"
)

// Reusing mockCounter from openai_provider_test.go if it's in the same package (infra_test)
// Assuming openai_provider_test.go defines mockCounter in package infra_test

func BenchmarkParseOutputUsage_Stream(b *testing.B) {
	// Construct a large streaming response
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		chunk := fmt.Sprintf(`data: {"id":"chatcmpl-%d","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":" word%d"},"finish_reason":null}]}`+"\n\n", i, i)
		buf.WriteString(chunk)
	}
	buf.WriteString("data: [DONE]\n\n")
	respBody := buf.Bytes()

	p := infra.NewOpenAIProvider("test-key", "", nil, &mockCounter{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseOutputUsage(domain.Model("gpt-3.5-turbo"), respBody, true)
		if err != nil {
			b.Fatalf("ParseOutputUsage failed: %v", err)
		}
	}
}

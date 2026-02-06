package token

import (
	"github.com/pkoukk/tiktoken-go"
)

// Counter handles token counting logic.
type Counter struct {
	// Cache codecs to avoid re-initializing
	codecs map[string]*tiktoken.Tiktoken
}

func NewCounter() *Counter {
	return &Counter{
		codecs: make(map[string]*tiktoken.Tiktoken),
	}
}

func (c *Counter) getCodec(model string) (*tiktoken.Tiktoken, error) {
	// Simple caching strategy. Not thread-safe for writes but fine if pre-warmed or mutexed.
	// For MVP we just load it every time or optimize later. 
	// tiktoken-go does internal caching usually.
    // Let's rely on tiktoken.EncodingForModel which is efficient.
	
	// OpenAI models often share encodings. 
	// gpt-4, gpt-3.5-turbo -> cl100k_base
	
    // Handle model aliases or prefixes if necessary
    encoding, err := tiktoken.EncodingForModel(model)
    if err != nil {
        // Fallback to cl100k_base for newer unknown models
        return tiktoken.GetEncoding("cl100k_base")
    }
    return encoding, nil
}

// CountTokens counts tokens in a text string for a given model.
func (c *Counter) CountTokens(model, text string) (int, error) {
	tkm, err := c.getCodec(model)
	if err != nil {
		return 0, err
	}
	
	tokenized := tkm.Encode(text, nil, nil)
	return len(tokenized), nil
}

// EstimateRequestCost estimates the token usage for a request.
// returns input_tokens, estimated_total_tokens
func (c *Counter) EstimateCost(model string, input string, maxTokens int) (int, int, error) {
	inputCount, err := c.CountTokens(model, input)
	if err != nil {
		return 0, 0, err
	}
    
    // Safety buffer?
    // Usually total = input + max_tokens
    
    return inputCount, inputCount + maxTokens, nil
}

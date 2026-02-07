package providers

import (
	"github.com/pkoukk/tiktoken-go"
)

type TiktokenCounter struct{}

func NewTiktokenCounter() *TiktokenCounter {
	return &TiktokenCounter{}
}

func (c *TiktokenCounter) Count(model string, text string) (int, error) {
	encoding, err := tiktoken.EncodingForModel(model)
	if err != nil {
		encoding, _ = tiktoken.GetEncoding("cl100k_base")
	}
	tokenized := encoding.Encode(text, nil, nil)
	return len(tokenized), nil
}

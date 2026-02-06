package proxy

import (
	"context"
	"pouch-ai/internal/domain/key"
	"pouch-ai/internal/domain/provider"
)

type Request struct {
	Context  context.Context
	Key      *key.Key
	Provider provider.Provider
	Model    provider.Model
	RawBody  []byte
	IsStream bool
}

type Response struct {
	StatusCode   int
	Body         []byte
	PromptTokens int
	OutputTokens int
	TotalCost    float64
}

type Handler interface {
	Handle(req *Request) (*Response, error)
}

type Middleware interface {
	Execute(req *Request, next Handler) (*Response, error)
}

type MiddlewareFunc func(req *Request, next Handler) (*Response, error)

func (f MiddlewareFunc) Execute(req *Request, next Handler) (*Response, error) {
	return f(req, next)
}

type Chain struct {
	middlewares []Middleware
	final       Handler
}

func NewChain(final Handler, middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
		final:       final,
	}
}

func (c *Chain) Handle(req *Request) (*Response, error) {
	return c.execute(req, 0)
}

func (c *Chain) execute(req *Request, index int) (*Response, error) {
	if index < len(c.middlewares) {
		return c.middlewares[index].Execute(req, HandlerFunc(func(r *Request) (*Response, error) {
			return c.execute(r, index+1)
		}))
	}
	return c.final.Handle(req)
}

type HandlerFunc func(req *Request) (*Response, error)

func (f HandlerFunc) Handle(req *Request) (*Response, error) {
	return f(req)
}

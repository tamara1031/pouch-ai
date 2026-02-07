package domain

import (
	"context"
	"io"
	"net/http"
)

type UsageCommitter interface {
	CommitUsage(ctx context.Context, keyID ID, reserved, actual float64) error
}

type Request struct {
	Context      context.Context
	Key          *Key
	Provider     Provider
	Model        Model
	RawBody      []byte
	IsStream     bool
	ReservedCost float64
	Committer    UsageCommitter
}

type Response struct {
	StatusCode   int
	Header       http.Header
	Body         io.ReadCloser
	PromptTokens int
	OutputTokens int
	TotalCost    float64
}

type Handler interface {
	Handle(req *Request) (*Response, error)
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

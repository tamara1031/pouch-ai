package domain

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	Context  context.Context
	Key      *Key
	Provider Provider
	Model    Model
	RawBody  []byte
	IsStream bool
}

type Response struct {
	StatusCode   int
	Header       http.Header
	Body         io.ReadCloser
	PromptTokens int
	OutputTokens int
	TotalCost    float64
}

type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeSelect  FieldType = "select"
)

type FieldRole string

const (
	FieldRoleLimit  FieldRole = "limit"
	FieldRolePeriod FieldRole = "period"
)

type FieldSchema struct {
	Type        FieldType `json:"type"`
	DisplayName string    `json:"display_name,omitempty"`
	Default     string    `json:"default,omitempty"`
	Description string    `json:"description,omitempty"`
	Options     []string  `json:"options,omitempty"`
	Role        FieldRole `json:"role,omitempty"`
}

type MiddlewareSchema map[string]FieldSchema

type MiddlewareInfo struct {
	ID     string           `json:"id"`
	Schema MiddlewareSchema `json:"schema"`
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

type MiddlewareRegistry interface {
	Register(id string, factory func(config map[string]string) Middleware, schema MiddlewareSchema)
	Get(id string, config map[string]string) (Middleware, error)
	List() []MiddlewareInfo
}

type DefaultMiddlewareRegistry struct {
	plugins map[string]regEntry
}

type regEntry struct {
	factory func(config map[string]string) Middleware
	schema  MiddlewareSchema
}

func NewMiddlewareRegistry() MiddlewareRegistry {
	return &DefaultMiddlewareRegistry{plugins: make(map[string]regEntry)}
}

func (r *DefaultMiddlewareRegistry) Register(id string, factory func(config map[string]string) Middleware, schema MiddlewareSchema) {
	r.plugins[id] = regEntry{factory: factory, schema: schema}
}

func (r *DefaultMiddlewareRegistry) Get(id string, config map[string]string) (Middleware, error) {
	entry, ok := r.plugins[id]
	if !ok {
		return nil, fmt.Errorf("middleware not found: %s", id)
	}
	return entry.factory(config), nil
}

func (r *DefaultMiddlewareRegistry) List() []MiddlewareInfo {
	var infos []MiddlewareInfo
	for id, entry := range r.plugins {
		infos = append(infos, MiddlewareInfo{
			ID:     id,
			Schema: entry.schema,
		})
	}
	return infos
}

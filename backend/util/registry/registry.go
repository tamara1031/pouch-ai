package registry

import (
	"fmt"
	"sync"
)

type Registry[T any] interface {
	Register(name string, item T)
	Get(name string) (T, error)
	List() []T
	ListKeys() []string
}

type DefaultRegistry[T any] struct {
	mu    sync.RWMutex
	items map[string]T
}

func NewRegistry[T any]() *DefaultRegistry[T] {
	return &DefaultRegistry[T]{
		items: make(map[string]T),
	}
}

func (r *DefaultRegistry[T]) Register(name string, item T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[name] = item
}

var ErrNotFound = fmt.Errorf("item not found in registry")

func (r *DefaultRegistry[T]) Get(name string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[name]
	if !ok {
		var zero T
		return zero, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return item, nil
}

func (r *DefaultRegistry[T]) List() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]T, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items
}

func (r *DefaultRegistry[T]) ListKeys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.items))
	for k := range r.items {
		keys = append(keys, k)
	}
	return keys
}

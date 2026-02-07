package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"pouch-ai/backend/domain"
)

type PluginManager struct {
	registry  domain.MiddlewareRegistry
	pluginDir string
}

func NewPluginManager(registry domain.MiddlewareRegistry, pluginDir string) *PluginManager {
	return &PluginManager{
		registry:  registry,
		pluginDir: pluginDir,
	}
}

func (m *PluginManager) LoadPlugins() error {
	// Ensure directory exists
	if _, err := os.Stat(m.pluginDir); os.IsNotExist(err) {
		return os.MkdirAll(m.pluginDir, 0755)
	}

	files, err := filepath.Glob(filepath.Join(m.pluginDir, "*.so"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := m.loadPlugin(file); err != nil {
			fmt.Printf("Error loading plugin %s: %v\n", file, err)
		}
	}

	return nil
}

func (m *PluginManager) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("could not open plugin: %w", err)
	}

	symbol, err := p.Lookup("GetFactory")
	if err != nil {
		return fmt.Errorf("could not find GetFactory symbol: %w", err)
	}

	factory, ok := symbol.(*func(config map[string]any) domain.Middleware)
	if !ok {
		return fmt.Errorf("GetFactory symbol has wrong type: expected *func(map[string]any) domain.Middleware")
	}

	// Optional: lookup schema
	var schema domain.MiddlewareSchema
	schemaSymbol, err := p.Lookup("GetSchema")
	if err == nil {
		if getSchema, ok := schemaSymbol.(*func() domain.MiddlewareSchema); ok {
			schema = (*getSchema)()
		}
	} else {
		schema = domain.MiddlewareSchema{}
	}

	// The ID of the middleware is the filename without extension
	id := filepath.Base(path)
	id = id[:len(id)-len(filepath.Ext(id))]

	m.registry.Register(domain.MiddlewareInfo{
		ID:     id,
		Schema: schema,
	}, *factory)
	fmt.Printf("Registered plugin middleware: %s (with schema: %v)\n", id, len(schema) > 0)

	return nil
}

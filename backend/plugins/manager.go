package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"pouch-ai/backend/config"
	"pouch-ai/backend/domain"
	"pouch-ai/backend/plugins/middlewares"
	"pouch-ai/backend/plugins/providers"
	"pouch-ai/backend/util/logger"
)

type PluginManager struct {
	mwRegistry domain.MiddlewareRegistry
	pRegistry  domain.ProviderRegistry
	cfg        *config.Config
	pluginDir  string
}

func NewPluginManager(mwRegistry domain.MiddlewareRegistry, pRegistry domain.ProviderRegistry, cfg *config.Config, pluginDir string) *PluginManager {
	return &PluginManager{
		mwRegistry: mwRegistry,
		pRegistry:  pRegistry,
		cfg:        cfg,
		pluginDir:  pluginDir,
	}
}

// InitializeBuiltins registers all built-in middlewares and providers.
func (m *PluginManager) InitializeBuiltins() error {
	// 1. Initialize Middlewares
	for _, builtin := range middlewares.GetBuiltins() {
		m.mwRegistry.Register(builtin.Info.ID, domain.MiddlewareEntry{
			Info:    builtin.Info,
			Factory: builtin.Factory,
		})
	}

	// 2. Initialize Providers via Builders
	ctx := context.Background()
	for _, b := range providers.GetBuilders() {
		p, err := b.Build(ctx, m.cfg)
		if err != nil {
			return fmt.Errorf("failed to build provider: %w", err)
		}
		if p != nil {
			m.pRegistry.Register(p.Name(), p)
			logger.L.Info("Registered built-in provider", "name", p.Name())
		}
	}

	return nil
}

func (m *PluginManager) LoadPlugins() error {
	// Ensure directory exists
	if _, err := os.Stat(m.pluginDir); os.IsNotExist(err) {
		if err := os.MkdirAll(m.pluginDir, 0755); err != nil {
			return err
		}
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

	sym, err := p.Lookup("GetPlugin")
	if err != nil {
		return fmt.Errorf("could not find GetPlugin symbol: %w", err)
	}

	getPlugin, ok := sym.(func() domain.MiddlewareEntry)
	if !ok {
		return fmt.Errorf("GetPlugin symbol has wrong type: expected func() domain.MiddlewareEntry")
	}

	entry := getPlugin()
	if entry.Info.ID == "" {
		// Fallback to filename if ID is missing
		id := filepath.Base(path)
		entry.Info.ID = id[:len(id)-len(filepath.Ext(id))]
	}

	m.mwRegistry.Register(entry.Info.ID, entry)
	fmt.Printf("Registered plugin middleware: %s\n", entry.Info.ID)

	return nil
}

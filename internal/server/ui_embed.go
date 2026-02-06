package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

// UIAssets holds the embedded UI files.
// We expect the build output to be in ui/dist
// The embed directive must be in the package or main. 
// Ideally we pass the fs.FS from main to this package.
// But for simplicity, let's assume we pass it in.

func (s *Server) SetupStaticAssets(assets fs.FS) {
	// Root should serve index.html
	// We need to serve internal files from 'dist' probably.
	
	assetHandler := http.FileServer(http.FS(assets))
	
	// Serve static files
	s.echo.GET("/*", echo.WrapHandler(assetHandler))
    
    // Fallback for SPA routing if needed (Astro SSG usually doesn't need wildcard fallback unless using client router which we might not be)
}

// Custom handler to strip prefix if needed?
// If we embed "ui/dist", the root is "", so it's fine.

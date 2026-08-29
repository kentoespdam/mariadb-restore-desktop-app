package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/baguspdam/mariadb-restore-desktop-app/src/backend/app"
)

//go:embed all:src/frontend/dist
var assets embed.FS

func main() {
	exeDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "exeDir:", err)
		os.Exit(1)
	}
	if filepath.Base(exeDir) == "build" {
		exeDir = filepath.Dir(exeDir)
	}

	// Build with a background ctx; the real wails ctx replaces the
	// event bus during OnStartup.
	a, err := app.New(context.Background(), exeDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap:", err)
		os.Exit(1)
	}
	defer a.Close()

	err = wails.Run(&options.App{
		Title:  "Portable MariaDB Restore & Backup",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			// The platform/events package stashes the wails ctx in a
			// package var; after this call, events.Default(ctx) and
			// the recovery bus route through wails.
			a.RebindCtx(ctx)
		},
		OnShutdown: func(_ context.Context) { a.Close() },
		Bind: []interface{}{
			a,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "wails:", err)
	}
}

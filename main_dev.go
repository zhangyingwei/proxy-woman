//go:build dev

package main

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 开发模式下的额外日志
	log.Println("🚀 Starting ProxyWoman in development mode...")

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ProxyWoman (Dev)",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup: func(ctx context.Context) {
			log.Println("🎯 App startup in development mode")
			app.startup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			log.Println("🛑 App shutdown in development mode")
			app.Shutdown(ctx)
		},
		Bind: []interface{}{
			app,
		},
		// 开发模式特定选项
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		log.Printf("❌ Error: %v\n", err)
		fmt.Printf("Error: %v\n", err)
	}
}

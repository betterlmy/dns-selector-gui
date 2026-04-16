package main

import (
	"embed"

	"dns-selector-gui/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// 嵌入前端构建产物
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建应用主服务实例
	app := backend.NewAppService()

	// 配置并启动 Wails 应用
	err := wails.Run(&options.App{
		Title:     "DNS Selector v0.1.0", // 窗口标题
		Width:     1280,                  // 默认窗口宽度（16:9）
		Height:    720,                   // 默认窗口高度（16:9）
		MinWidth:  800,                   // 最小窗口宽度
		MinHeight: 450,                   // 最小窗口高度（16:9）
		AssetServer: &assetserver.Options{
			Assets: assets, // 嵌入的前端静态资源
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1}, // 默认背景色（白色）
		OnStartup:        app.OnStartup,                               // 应用启动回调
		Bind: []interface{}{
			app, // 绑定 AppService，前端可调用其所有公开方法
		},
	})

	if err != nil {
		println("启动失败:", err.Error())
	}
}

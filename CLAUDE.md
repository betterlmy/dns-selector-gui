# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目简介

基于 Wails v2 的跨平台桌面 GUI，是 `dns-selector` CLI 的图形化版本。用于测试 DNS 服务器（UDP / DoT / DoH）性能，并将最优结果应用到系统。

技术栈：Go 1.21+ 后端，React 18 + TypeScript + Zustand + Recharts 前端，Wails v2 作为桥接层。

## 常用命令

```bash
make dev                      # Wails 开发模式（热重载）
make build                    # 构建当前平台
make test                     # go test ./backend/... -v -count=1
make deps                     # npm install + go mod download + 拷贝图标到 build/
make frontend                 # wails generate module + 构建 frontend/dist
make clean

# 跨平台构建：make build-{windows,mac,linux}-{amd64,arm64}
# macOS 打包：make dmg-amd64 / dmg-arm64  （需要 create-dmg）
```

运行单个测试 / 过滤测试：

```bash
go test ./backend/ -v -run TestSomething
go test ./backend/ -v -run TestProperty        # 仅运行基于 rapid 的属性测试
go test ./backend/ -v -short                   # 跳过集成测试
```

所有 `wails build` 目标都依赖 `make deps` —— 它会把 `assets/appicon.png`、`assets/icons.icns`、`assets/icon.ico` 拷贝到 `build/`，Wails 构建需要这些文件。

macOS 构建使用 `wails build -nopackage`，然后手工组装 `.app` bundle 并执行 ad-hoc `codesign -`（见 Makefile 中的 `wails_build_mac`）。**不要**在 darwin 上改回普通的 `wails build` —— 会破坏 DMG 打包流程和签名。

## 架构

**Wails 桥接。** `main.go` 通过 `go:embed` 嵌入 `frontend/dist`，并向 JS 侧只绑定一个 `*backend.AppService`。前端能调用的每个后端方法，都对应 `AppService`（或其内嵌服务）上的一个导出方法。新增后端调用时：在这里加方法，然后重新执行 `wails generate module`（通过 `make frontend`）以刷新 `frontend/wailsjs/` 下的 TypeScript 绑定。

**后端服务组织**（`backend/`）：

- `app_service.go` —— 顶层服务，组合下面几个子服务；这是 Wails 实际绑定的对象。
- `benchmark_service.go` —— 通过外部 `dns-selector/selector` SDK 执行 DNS 测试。
- `preset_service.go` + `presets.go` —— 内置 CN / Global 预设方案。
- `config_service.go` —— 将用户配置持久化到 `os.UserConfigDir()/dns-selector-gui/`。
- `dns_config_service.go` —— 读写系统 DNS；委托给 `PlatformProvider`。
- `validation.go`、`models.go` —— 输入校验 + 共享 DTO（同时也决定了 TS 绑定的形状）。

**平台抽象。** `platform.go` 定义 `PlatformProvider`；`platform_{windows,darwin,linux}.go` 通过 build tag 选择实现。Windows 使用注册表 + `iphlpapi.dll`，macOS 使用 `networksetup`，Linux 使用 `resolvectl` / `/etc/resolv.conf`。修改系统 DNS 需要管理员/root 权限 —— 应用假定宿主已完成提权。任何新增的平台相关行为都应作为 `PlatformProvider` 的方法加进来并补齐三端实现，**不要**在 service 里用 `runtime.GOOS` 分支处理。

**测试。** 除常规 `*_test.go` 外，多个文件使用 `github.com/flyingmutant/rapid` 做属性测试（`*_property_test.go`）。涉及网络或真实系统 DNS 的集成测试通过 `testing.Short()` 开关控制 —— 新增此类测试时也要遵守 `-short` 约定。

**前端。** `frontend/src/` —— `components/` 是 UI，`store/` 是 Zustand 状态切片，还有 `hooks/`、`styles/`、`types/`。Wails 自动生成的绑定位于 `frontend/wailsjs/`（不要手改）。主题跟随系统亮/暗色。

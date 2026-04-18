# DNS Selector GUI

<p align="center">
  <img src="assets/icon.svg" width="128" alt="DNS Selector GUI">
</p>

跨平台 DNS 择优器桌面应用 —— 开源 CLI 项目 [dns-selector](https://github.com/betterlmy/dns-selector) 的图形化版本。

面向 Windows / macOS / Linux 的 DNS 择优桌面应用，通过图形界面测试多个 DNS 服务器（UDP / DoT / DoH）的响应速度、稳定性和正确性，并在可转换为系统 DNS 地址时帮助你把结果应用到系统网络配置。

## 功能特性

- **多协议测速**：同时测试 UDP、DNS-over-TLS、DNS-over-HTTPS
- **内置预设方案**：CN（中国大陆）和 Global（全球）两套预设，开箱即用
- **综合评分排名**：基于中位延迟、成功率、抖动等指标自动推荐最优 DNS
- **答案验证**：多服务器共识验证，识别返回异常结果的 DNS 服务器
- **可视化图表**：Score 柱状图对比，鼠标悬停查看详细数据
- **系统 DNS 修改**：对可转换为系统 DNS 地址的结果，直接在应用内修改系统网络适配器的 DNS 设置
- **配置持久化**：自定义服务器、域名、测试参数自动保存，支持 JSON 导入导出
- **预设切换不丢数据**：切换 CN / Global 预设时保留自定义服务器和域名
- **浅色/深色主题**：自动跟随系统主题设置

> 说明：DoT/DoH 项都可以参与测速，但并不是所有 DoT/DoH 结果都能直接写入系统 DNS。当前实现仅在后端能解析出可用系统 DNS 地址时才允许“应用到系统”。

## 平台支持

以下平台具备实现代码。当前仓库内已经有 GitHub Actions 发布工作流，但持续验证仍以本地实测和手动验证为主。

| 平台 | 架构 | 安装包格式 | DNS 修改机制 |
|------|------|-----------|-------------|
| Windows 10/11 | amd64, arm64 | `.exe` | 注册表 + `iphlpapi.dll` |
| macOS 11+ | amd64 (Intel), arm64 (Apple Silicon) | `.dmg` | `networksetup` |
| Linux (GTK3) | amd64, arm64 | 二进制 | `resolvectl` / `/etc/resolv.conf` |

当前验证状态：

- macOS 本机已实测通过：`make build`、`make build-mac-amd64`、`make build-mac-arm64`、`make build-mac-universal`、`make dmg-amd64`、`make dmg-arm64`
- macOS 本机已实测通过：`make build-windows-amd64`、`make build-windows-arm64`
- macOS 本机不支持 Wails Linux 交叉编译；`make build-linux-*` 会明确失败并提示改到 Linux 主机或 CI 上执行

## 技术栈

| 层级 | 技术 |
|------|------|
| GUI 框架 | [Wails v2](https://wails.io/) |
| 后端 | Go 1.25.0 |
| 前端 | React 18 + TypeScript 5 |
| 状态管理 | Zustand |
| 图表 | Recharts |
| DNS 引擎 | [dns-selector/selector](https://github.com/betterlmy/dns-selector) SDK |
| 属性测试 | [rapid](https://github.com/flyingmutant/rapid) |

## 项目结构

```
dns-selector-win/
├── main.go                        # Wails 应用入口
├── wails.json                     # Wails 项目配置
├── Makefile                       # 跨平台构建脚本
├── backend/                       # Go 后端服务
│   ├── platform.go                # PlatformProvider 跨平台接口定义
│   ├── platform_windows.go        # Windows 平台实现 (build tag: windows)
│   ├── platform_darwin.go         # macOS 平台实现 (build tag: darwin)
│   ├── platform_linux.go          # Linux 平台实现 (build tag: linux)
│   ├── models.go                  # 数据模型定义
│   ├── app_service.go             # 主服务，Wails 绑定入口
│   ├── benchmark_service.go       # DNS 测试服务
│   ├── config_service.go          # 配置文件读写（os.UserConfigDir）
│   ├── dns_config_service.go      # DNS 配置服务（委托 PlatformProvider）
│   ├── preset_service.go          # 预设方案管理
│   ├── presets.go                 # CN / Global 预设数据
│   ├── validation.go              # 输入验证
│   ├── *_test.go                  # 单元测试
│   └── *_property_test.go         # 属性测试 (PBT)
├── frontend/                      # React 前端
│   └── src/
│       ├── components/            # UI 组件
│       ├── store/                 # Zustand 状态管理
│       ├── hooks/                 # 自定义 Hook
│       ├── styles/                # 主题和全局样式
│       └── types/                 # TypeScript 类型定义
```

## 环境要求

- **Go** 1.25.0
- **Node.js** 18+
- **Wails CLI** v2

平台特定依赖：

| 平台 | 额外依赖 |
|------|---------|
| Windows | WebView2 运行时（Win10 1803+ / Win11 已内置） |
| macOS | Xcode Command Line Tools |
| Linux | `libgtk-3-dev` `libwebkit2gtk-4.0-dev` |

## 快速开始

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# 克隆并安装依赖
git clone <仓库地址> && cd dns-selector-win
make deps

# 开发模式（热重载）
make dev

# 构建当前平台
make build
```

## 构建

```bash
make build                  # 当前平台
make build-all              # 当前宿主机支持的全部目标

# 单平台构建
make build-windows-amd64    # Windows x64
make build-windows-arm64    # Windows ARM64
make build-mac-amd64        # macOS Intel
make build-mac-arm64        # macOS Apple Silicon
make build-linux-amd64      # Linux x64
make build-linux-arm64      # Linux ARM64

# macOS DMG 打包（需要 create-dmg）
make dmg-amd64
make dmg-arm64
```

构建产物命名格式：`dns-selector-gui-{version}-{os}-{arch}.{ext}`

说明：

- 在 macOS 主机上，`make build` 会走手工 `go build` 的 Wails 应用构建链路，并显式补齐 `UniformTypeIdentifiers` framework
- 在 macOS 主机上，`make build-all` 当前会构建 4 个已验证目标：`windows/amd64`、`windows/arm64`、`darwin/amd64`、`darwin/arm64`
- `make build-linux-*` 需要在 Linux 主机或 Linux CI runner 上执行
- `make dmg-*` 依赖 `create-dmg` 和 macOS 的 `hdiutil`

## GitHub Release

仓库内提供了基于 tag 的发布工作流：

- 触发方式：push 一个形如 `v0.1.0` 的 tag
- 工作流文件：[`.github/workflows/build.yml`](.github/workflows/build.yml)
- 当前 release 产物矩阵：`windows/amd64`、`windows/arm64`、`darwin/amd64 (.dmg)`、`darwin/arm64 (.dmg)`、`linux/amd64`

说明：

- macOS 工作流已改为和本地一致的手工构建链路：`go build` 产出二进制，再包装 `.app` 和 `.dmg`
- 我已在本地验证过对应的 macOS 与 Windows 产物链路，但没有在当前环境里实际 push tag 到 GitHub 远端执行一次 Actions

## 测试

```bash
make test                                   # 默认运行稳定后端测试（short）
make test-unit                              # 运行后端完整测试集
make test-integration                       # 仅运行集成测试
go test ./backend/... -v -run "TestProperty"  # 仅属性测试
go test ./backend/... -v -short                # 手动跳过集成测试
```

`make test` 默认只跑 `testing.Short()` 下稳定通过的测试；依赖真实网络或系统环境的测试放到 `make test-integration`。

## 运行与行为说明

- 预设切换只影响内置服务器/域名集合，不会清空自定义服务器和自定义域名。
- 测试参数会自动保存，但前端已做 debounce，连续输入不会每击键落盘。
- 最近一次测速结果会持久化为 `last_results.json`，下次启动时自动加载。

## 使用说明

1. 选择预设方案（CN / Global），可选添加自定义 DNS 服务器和域名
2. 点击"开始测试"，等待测试完成
3. 查看结果表格和 Score 柱状图
4. 从结果中选择最优 DNS，点击应用到系统

> 修改系统 DNS 可能需要额外权限。Windows 通常需要管理员权限；Linux 取决于系统 DNS 管理方式；macOS 代码路径使用 `networksetup`，`CheckAdmin()` 当前返回可用，但真实执行仍可能因系统授权失败而报错。

### 配置文件位置

| 平台 | 配置文件 | 最近结果文件 |
|------|-----------|--------------|
| Windows | `%APPDATA%\dns-selector-gui\config.json` | `%APPDATA%\dns-selector-gui\last_results.json` |
| macOS | `~/Library/Application Support/dns-selector-gui/config.json` | `~/Library/Application Support/dns-selector-gui/last_results.json` |
| Linux | `~/.config/dns-selector-gui/config.json` | `~/.config/dns-selector-gui/last_results.json` |

## 许可证

MIT

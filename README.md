# DNS Selector GUI

<p align="center">
  <img src="assets/icon.svg" width="128" alt="DNS Selector GUI">
</p>

跨平台 DNS 择优器桌面应用 —— 开源 CLI 项目 [dns-selector](https://github.com/betterlmy/dns-selector) 的图形化版本。

支持 Windows / macOS / Linux 三平台（amd64 + arm64），通过图形界面测试多个 DNS 服务器（UDP / DoT / DoH）的响应速度、稳定性和正确性，帮助你选择最优 DNS 并一键应用到系统。

## 功能特性

- **多协议支持**：同时测试 UDP、DNS-over-TLS、DNS-over-HTTPS
- **内置预设方案**：CN（中国大陆）和 Global（全球）两套预设，开箱即用
- **综合评分排名**：基于中位延迟、成功率、抖动等指标自动推荐最优 DNS
- **答案验证**：多服务器共识验证，识别返回异常结果的 DNS 服务器
- **可视化图表**：Score 柱状图对比，鼠标悬停查看详细数据
- **系统 DNS 修改**：直接在应用内修改系统网络适配器的 DNS 设置
- **配置持久化**：自定义服务器、域名、测试参数自动保存，支持 JSON 导入导出
- **浅色/深色主题**：自动跟随系统主题设置

## 平台支持

| 平台 | 架构 | 安装包格式 | DNS 修改机制 |
|------|------|-----------|-------------|
| Windows 10/11 | amd64, arm64 | `.exe` | 注册表 + iphlpapi.dll |
| macOS 11+ | amd64 (Intel), arm64 (Apple Silicon) | `.dmg` | `networksetup` |
| Linux (GTK3) | amd64, arm64 | 二进制 | `resolvectl` / `/etc/resolv.conf` |

## 技术栈

| 层级 | 技术 |
|------|------|
| GUI 框架 | [Wails v2](https://wails.io/) |
| 后端 | Go 1.21+ |
| 前端 | React 18 + TypeScript 5 |
| 状态管理 | Zustand |
| 图表 | Recharts |
| DNS 引擎 | [dns-selector/selector](https://github.com/betterlmy/dns-selector) SDK |
| 属性测试 | [rapid](https://github.com/flyingmutant/rapid) |

## 项目结构

```
dns-selector-gui/
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
└── .github/workflows/build.yml    # CI/CD 多平台构建与发布
```

## 环境要求

- **Go** 1.21+
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
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆并安装依赖
git clone <仓库地址> && cd dns-selector-gui
make deps

# 开发模式（热重载）
make dev

# 构建当前平台
make build
```

## 构建

```bash
make build                  # 当前平台
make build-all              # 所有平台（6 个目标）

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

## 测试

```bash
make test                              # 运行所有后端测试
go test ./backend/ -v -run "TestProperty"  # 仅属性测试
go test ./backend/ -v -short               # 跳过集成测试
```

## 使用说明

1. 选择预设方案（CN / Global），可选添加自定义 DNS 服务器和域名
2. 点击"开始测试"，等待测试完成
3. 查看结果表格和 Score 柱状图
4. 从结果中选择最优 DNS，点击应用到系统

> 修改系统 DNS 需要管理员/root 权限。Windows 需以管理员身份运行，macOS/Linux 需 sudo。

### 配置文件位置

| 平台 | 路径 |
|------|------|
| Windows | `%APPDATA%\dns-selector-gui\` |
| macOS | `~/Library/Application Support/dns-selector-gui/` |
| Linux | `~/.config/dns-selector-gui/` |

## 许可证

MIT

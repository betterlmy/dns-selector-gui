# DNS Selector GUI

Windows DNS 择优器可视化桌面应用 —— 开源 CLI 项目 [dns-selector](https://github.com/betterlmy/dns-selector) 的图形化版本。

通过图形界面测试多个 DNS 服务器（支持 UDP / DoT / DoH 三种协议）的响应速度、稳定性和正确性，帮助你选择最优的 DNS 服务器，并支持一键修改 Windows 系统 DNS 配置。

## 功能特性

- **多协议支持**：同时测试 UDP（传统 DNS）、DNS-over-TLS、DNS-over-HTTPS 三种协议
- **内置预设方案**：CN（中国大陆）和 Global（全球）两套预设，开箱即用
- **综合评分排名**：基于中位延迟、成功率、抖动等指标计算 Score，自动推荐最优 DNS
- **答案验证**：多服务器共识验证，识别返回异常结果的 DNS 服务器
- **可视化图表**：Score 柱状图对比，鼠标悬停查看详细数据
- **系统 DNS 修改**：直接在应用内修改 Windows 网络适配器的 DNS 设置
- **配置持久化**：自定义服务器、域名、测试参数自动保存，支持 JSON 导入导出
- **历史结果保存**：上次测试结果自动保存，下次打开无需重新测试
- **浅色/深色主题**：自动跟随 Windows 系统主题设置

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
├── main.go                      # Wails 应用入口
├── wails.json                   # Wails 项目配置
├── backend/                     # Go 后端服务
│   ├── models.go                # 数据模型定义（DTO 结构体）
│   ├── app_service.go           # 主服务，Wails 绑定入口
│   ├── benchmark_service.go     # DNS 测试服务（调用 selector SDK）
│   ├── config_service.go        # JSON 配置文件读写服务
│   ├── dns_config_service.go    # Windows 系统 DNS 修改服务
│   ├── preset_service.go        # 预设方案管理服务
│   ├── presets.go               # CN / Global 预设数据
│   ├── validation.go            # 输入验证（地址、域名、参数）
│   ├── *_test.go                # 单元测试
│   └── *_property_test.go       # 属性测试（Property-Based Testing）
├── frontend/                    # React 前端
│   ├── src/
│   │   ├── App.tsx              # 根组件
│   │   ├── main.tsx             # 入口文件
│   │   ├── types/index.ts       # TypeScript 类型定义
│   │   ├── store/useAppStore.ts # Zustand 全局状态管理
│   │   ├── hooks/               # 自定义 Hook
│   │   ├── styles/              # 主题和全局样式
│   │   └── components/          # UI 组件
│   │       ├── layout/          # 主窗口布局
│   │       ├── preset/          # 预设方案选择
│   │       ├── servers/         # DNS 服务器列表管理
│   │       ├── domains/         # 测试域名列表管理
│   │       ├── params/          # 测试参数配置
│   │       ├── benchmark/       # 测试控制、结果表格、图表
│   │       ├── dns-config/      # 系统 DNS 配置
│   │       └── config/          # 配置导入导出
│   └── wailsjs/                 # Wails 自动生成的 JS 绑定
└── .kiro/specs/                 # 需求、设计、任务文档
```

## 环境要求

- **操作系统**：Windows 10/11（需要 WebView2 运行时，Win10 1803+ / Win11 已内置）
- **Go**：1.21 或更高版本
- **Node.js**：18 或更高版本
- **Wails CLI**：v2（安装方式见下方）

## 快速开始

### 1. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 2. 克隆项目

```bash
git clone <仓库地址>
cd dns-selector-gui
```

### 3. 安装依赖

```bash
# Go 依赖
go mod download

# 前端依赖
cd frontend && npm install && cd ..
```

### 4. 开发模式运行

```bash
wails dev
```

这会启动热重载的开发服务器，前端修改会自动刷新。

### 5. 构建发布版本

```bash
wails build
```

构建产物为 `build/bin/dns-selector-gui.exe`，单文件可执行，无需额外运行时。

## 使用说明

### 基本流程

1. **选择预设方案**：在左侧面板选择 CN（中国大陆）或 Global（全球）预设
2. **（可选）自定义配置**：添加自定义 DNS 服务器或测试域名，调整测试参数
3. **开始测试**：点击"开始测试"按钮，等待测试完成
4. **查看结果**：在右侧面板查看测试结果表格和 Score 柱状图
5. **应用 DNS**：从结果中选择最优 DNS，点击应用到系统网络适配器

### 测试参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| 查询次数 (queries) | 10 | 每个域名的正式查询次数 |
| 预热次数 (warmup) | 1 | 每个服务器的预热查询次数 |
| 最大并发 (concurrency) | 20 | 同时进行的最大查询数 |
| 超时时间 (timeout) | 2.0 秒 | 单次查询的超时时间 |

### Score 计算公式

```
Score = (1 / 中位延迟秒) × (成功率²) × (中位延迟 / P95延迟)
```

- 有效样本数 < 5 时，省略抖动惩罚因子（中位延迟/P95延迟）
- 所有查询均超时的服务器，Score 为 0

### 修改系统 DNS

修改 Windows 系统 DNS 需要管理员权限。如果应用未以管理员身份运行，会提示你重新以管理员身份启动。

### 配置文件

应用配置自动保存在 `%APPDATA%/dns-selector-gui/config.json`，测试结果保存在 `%APPDATA%/dns-selector-gui/last_results.json`。

支持通过"导入配置"/"导出配置"按钮进行 JSON 配置文件的导入导出。

## 运行测试

```bash
# 运行所有后端测试（包含属性测试和集成测试）
go test ./backend/ -v -count=1

# 仅运行属性测试
go test ./backend/ -v -run "TestProperty"

# 跳过集成测试（不需要网络）
go test ./backend/ -v -short
```

项目包含 9 个属性测试（Property-Based Testing），每个至少运行 100 次随机迭代：

| 属性 | 验证内容 |
|------|---------|
| Property 1 | 预设项不可删除 |
| Property 2 | 服务器地址格式验证 |
| Property 3 | 自定义项删除缩减列表 |
| Property 4 | 添加有效域名扩展列表 |
| Property 5 | 测试参数验证 |
| Property 6 | Score 计算公式正确性 |
| Property 7 | 结果按 Score 降序排列且推荐最高分 |
| Property 8 | 配置文件 JSON 序列化往返一致 |
| Property 9 | 无效配置导入被拒绝 |

## 许可证

MIT

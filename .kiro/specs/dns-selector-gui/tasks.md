# 实现计划：DNS Selector GUI

## 概述

基于 Wails v2 + React + TypeScript + Go 技术栈，复用 `github.com/betterlmy/dns-selector/selector` SDK 构建 Windows DNS 择优器可视化桌面应用。实现按照后端服务层 → 前端组件层 → 集成联调的顺序推进，每个阶段包含对应的属性测试和单元测试。

## Tasks

- [x] 1. 初始化 Wails v2 项目结构和基础框架
  - [x] 1.1 使用 `wails init` 创建项目，模板选择 React + TypeScript
    - 配置 `wails.json`：设置应用名称 "DNS Selector"、窗口最小尺寸 800x600
    - 添加 Go 依赖：`github.com/betterlmy/dns-selector/selector`、`github.com/flyingmutant/rapid`
    - _Requirements: 12.2, 12.3_

  - [x] 1.2 定义后端数据模型和类型
    - 创建 `backend/models.go`，定义所有 DTO 结构体：`AddServerRequest`、`ServerInfo`、`DomainInfo`、`TestParams`、`TestResultItem`、`TestResultsData`、`NetworkAdapterInfo`、`AppConfig`、`CustomServerEntry`、`PersistedResults`
    - 设置 `TestParams` 默认值：queries=10, warmup=1, concurrency=20, timeout=2.0
    - _Requirements: 7.2, 13.2_

  - [x] 1.3 定义前端 TypeScript 类型和 Zustand Store 骨架
    - 创建 `frontend/src/types/index.ts`，定义与后端对应的 TypeScript 接口
    - 创建 `frontend/src/store/useAppStore.ts`，定义 Zustand Store 的 state 和 actions 骨架
    - _Requirements: 12.1_

- [x] 2. 实现预设管理服务（PresetService）
  - [x] 2.1 实现 CN 和 Global 预设数据
    - 创建 `backend/presets.go`，硬编码 CN_Preset（32 个 DNS 服务器 + 29 个测试域名）和 Global_Preset（16 个 DNS 服务器 + 24 个测试域名）
    - 使用 `selector.DNSServer` 结构体定义每个服务器，包含协议类型、地址、TLS 配置
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.3, 4.4_

  - [x] 2.2 实现 PresetService 核心逻辑
    - 创建 `backend/preset_service.go`，实现 `PresetService` 结构体
    - 实现 `SwitchPreset`：切换预设方案，清空自定义内容
    - 实现 `GetMergedServers` / `GetMergedDomains`：合并预设和自定义列表
    - 实现 `IsPresetItem`：判断项是否属于预设（不可删除）
    - 实现 `AddCustomServer` / `RemoveCustomServer` / `AddCustomDomain` / `RemoveCustomDomain`：自定义项的增删，预设项拒绝删除
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 5.7, 6.4_

  - [x] 2.3 属性测试：预设项不可删除（Property 1）
    - **Property 1: 预设项不可删除**
    - 使用 `rapid` 生成器随机选择预设列表中的服务器/域名，验证删除操作返回错误且列表不变
    - **Validates: Requirements 2.6**

  - [x] 2.4 属性测试：自定义项删除缩减列表（Property 3）
    - **Property 3: 自定义项删除缩减列表**
    - 使用 `rapid` 生成随机自定义项列表，随机选择一项删除，验证列表长度减 1 且该项不再存在
    - **Validates: Requirements 5.7, 6.4**

  - [x] 2.5 属性测试：添加有效域名扩展列表（Property 4）
    - **Property 4: 添加有效域名扩展列表**
    - 使用 `rapid` 生成随机有效域名，验证添加后列表长度增 1 且包含该域名
    - **Validates: Requirements 6.2**

  - [x] 2.6 单元测试：预设内容完整性和切换逻辑
    - 验证 CN 预设包含 32 个服务器和 29 个域名
    - 验证 Global 预设包含 16 个服务器和 24 个域名
    - 验证预设切换后列表内容正确替换
    - _Requirements: 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.3, 4.4_

- [x] 3. 实现输入验证和测试参数逻辑
  - [x] 3.1 实现服务器地址验证函数
    - 创建 `backend/validation.go`，实现 `ValidateServerAddress(protocol, address string) error`
    - UDP：验证有效 IPv4 地址
    - DoT：验证有效域名或 "IP@TLSServerName" 格式
    - DoH：验证有效 HTTPS URL 或 "https://IP/path@TLSServerName" 格式
    - 返回具体的格式错误提示信息
    - _Requirements: 5.3, 5.4, 5.5, 5.6_

  - [x] 3.2 实现测试参数验证函数
    - 在 `backend/validation.go` 中实现 `ValidateTestParams(params TestParams) error`
    - 验证 queries、warmup、concurrency 为正整数，timeout 为正数
    - _Requirements: 7.3_

  - [x] 3.3 实现域名格式验证函数
    - 在 `backend/validation.go` 中实现 `ValidateDomain(domain string) error`
    - 验证输入为有效的域名格式
    - _Requirements: 6.3_

  - [x] 3.4 属性测试：服务器地址格式验证（Property 2）
    - **Property 2: 服务器地址格式验证**
    - 使用 `rapid` 生成随机协议类型和随机字符串（含有效/无效格式），验证验证函数的接受/拒绝行为与格式规则一致
    - **Validates: Requirements 5.3, 5.4, 5.5, 5.6**

  - [x] 3.5 属性测试：测试参数验证（Property 5）
    - **Property 5: 测试参数验证**
    - 使用 `rapid` 生成随机整数和浮点数（含负数、零、极大值），验证验证函数的接受/拒绝行为
    - **Validates: Requirements 7.3**

- [x] 4. Checkpoint - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户。

- [x] 5. 实现配置服务（ConfigService）
  - [x] 5.1 实现 ConfigService 核心逻辑
    - 创建 `backend/config_service.go`，实现 `ConfigService` 结构体
    - 实现 `Load`：从 JSON 文件加载配置，文件不存在或损坏时返回默认配置
    - 实现 `Save`：将配置序列化为 JSON 并写入文件
    - 实现 `GetDefaultPath`：返回 `%APPDATA%/dns-selector-gui/config.json`
    - 实现 `UpdateConfig`：更新内存中的配置并自动保存
    - _Requirements: 13.1, 13.2, 13.6, 13.7, 13.8_

  - [x] 5.2 实现测试结果持久化
    - 在 `backend/config_service.go` 中添加 `SaveResults` / `LoadResults` 方法
    - 结果保存路径：`%APPDATA%/dns-selector-gui/last_results.json`
    - _Requirements: 14.1, 14.2, 14.3_

  - [x] 5.3 属性测试：配置文件 JSON 序列化往返一致（Property 8）
    - **Property 8: 配置文件 JSON 序列化往返一致**
    - 使用 `rapid` 生成随机 `AppConfig` 对象，验证序列化再反序列化后与原始对象等价
    - **Validates: Requirements 13.1, 13.2**

  - [x] 5.4 属性测试：无效配置导入被拒绝（Property 9）
    - **Property 9: 无效配置导入被拒绝**
    - 使用 `rapid` 生成随机无效 JSON 字符串，验证导入函数返回非空错误且当前配置不变
    - **Validates: Requirements 13.5**

  - [x] 5.5 单元测试：配置文件回退和持久化
    - 测试配置文件不存在时使用默认配置
    - 测试配置文件损坏时的回退行为
    - 测试结果持久化的读写正确性
    - _Requirements: 13.7, 13.8, 14.1, 14.2, 14.3_

- [x] 6. 实现测试服务（BenchmarkService）
  - [x] 6.1 实现 BenchmarkService 核心逻辑
    - 创建 `backend/benchmark_service.go`，实现 `BenchmarkService` 结构体
    - 实现 `BuildSelector`：根据服务器列表、域名列表和测试参数构建 `selector.Selector` 实例
    - 实现 `RunBenchmark`：在 goroutine 中执行测试，支持 context 取消，通过回调推送进度
    - 实现 `IsRunning`：返回测试运行状态
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

  - [x] 6.2 实现 Score 计算和结果排序逻辑
    - 在 `backend/benchmark_service.go` 中实现结果处理：从 `selector.BenchmarkResult` 提取指标，计算 Score
    - Score 公式：`(1/median_seconds) × (success_rate²) × (median/P95)`，有效样本 < 5 时省略抖动惩罚
    - 全部超时的服务器 Score 设为 0
    - 结果按 Score 降序排列，设置 `BestDNS` 为最高分服务器
    - _Requirements: 8.6, 8.7, 8.8, 8.9, 9.1, 9.2, 9.3, 9.4, 10.2, 10.4_

  - [x] 6.3 属性测试：Score 计算公式正确性（Property 6）
    - **Property 6: Score 计算公式正确性**
    - 使用 `rapid` 生成随机 median/P95/success_rate/样本数，验证 Score 计算结果符合公式
    - **Validates: Requirements 8.7, 8.8, 8.9**

  - [x] 6.4 属性测试：测试结果按 Score 降序排列且推荐最高分（Property 7）
    - **Property 7: 测试结果按 Score 降序排列且推荐最高分**
    - 使用 `rapid` 生成随机 `TestResultItem` 列表，验证排序后降序且 bestDNS 为最高分
    - **Validates: Requirements 10.2, 10.4**

- [x] 7. 实现系统 DNS 配置服务（DNSConfigService）
  - [x] 7.1 实现 DNSConfigService
    - 创建 `backend/dns_config_service.go`，实现 `DNSConfigService` 结构体
    - 实现 `GetAdapters`：通过 PowerShell 获取活动网络适配器及其 DNS 配置
    - 实现 `SetDNS`：通过 PowerShell 设置指定适配器的 DNS 服务器
    - 实现 `ResetToAuto`：通过 PowerShell 恢复 DHCP 自动获取
    - 实现 `CheckAdmin`：检查当前进程是否具有管理员权限
    - _Requirements: 11.1, 11.2, 11.4, 11.7_

- [x] 8. 实现 AppService 主服务和 Wails 绑定
  - [x] 8.1 实现 AppService 并整合所有后端服务
    - 创建 `backend/app_service.go`，实现 `AppService` 结构体
    - 在 `OnStartup` 中初始化所有子服务、加载配置、加载历史测试结果
    - 实现所有 Binding 方法：预设管理、服务器/域名增删、测试参数、测试执行/停止、结果获取、系统 DNS 操作、配置导入导出、主题检测
    - 实现 `StartBenchmark`：启动 goroutine 执行测试，通过 `runtime.EventsEmit` 推送 `benchmark:progress`、`benchmark:complete`、`benchmark:error`、`benchmark:stopped` 事件
    - _Requirements: 1.1-1.6, 5.1, 5.2, 6.1, 7.1, 7.4, 8.4, 8.5, 11.3, 11.5, 11.6, 12.4, 12.5, 13.3, 13.4_

  - [x] 8.2 配置 Wails main.go 入口
    - 在 `main.go` 中创建 `AppService` 实例，配置 Wails 应用选项（窗口标题、尺寸、Bind 列表）
    - 设置窗口最小尺寸 800x600，标题 "DNS Selector vX.Y.Z"
    - _Requirements: 12.2, 12.3_

- [x] 9. Checkpoint - 确保后端所有测试通过
  - 确保所有测试通过，如有问题请询问用户。

- [x] 10. 实现前端基础框架和主题系统
  - [x] 10.1 实现主题系统和全局样式
    - 创建 `frontend/src/styles/theme.ts`，定义浅色和深色主题的 CSS 变量
    - 创建 `frontend/src/styles/global.css`，设置全局样式和主题变量
    - 在 `App.tsx` 中实现主题 Provider，启动时通过 `GetSystemTheme()` 获取系统主题
    - _Requirements: 12.5_

  - [x] 10.2 实现 Zustand Store 完整逻辑
    - 完善 `frontend/src/store/useAppStore.ts`，实现所有 state 和 actions
    - 实现初始化 action：调用后端 Binding 加载预设、服务器列表、域名列表、测试参数、历史结果、网卡信息
    - _Requirements: 5.1, 6.1, 14.2_

  - [x] 10.3 实现 Wails 事件监听 Hook
    - 创建 `frontend/src/hooks/useWailsEvents.ts`，监听 `benchmark:progress`、`benchmark:complete`、`benchmark:error`、`benchmark:stopped` 事件
    - 事件触发时更新 Zustand Store 中的对应状态
    - _Requirements: 8.4_

  - [x] 10.4 实现主窗口布局
    - 创建 `frontend/src/components/layout/MainLayout.tsx`，实现左右分栏布局
    - 左侧：预设选择、服务器列表、域名列表、测试参数
    - 右侧：测试控制、测试结果、当前 DNS 配置
    - _Requirements: 12.1_

- [x] 11. 实现前端配置管理组件
  - [x] 11.1 实现预设方案选择组件
    - 创建 `frontend/src/components/preset/PresetSelector.tsx`
    - 提供 CN / Global 切换控件，切换时调用 `SwitchPreset` 并刷新列表
    - _Requirements: 2.2, 2.5_

  - [x] 11.2 实现 DNS 服务器列表和添加对话框
    - 创建 `frontend/src/components/servers/ServerList.tsx`：展示服务器列表，协议标签区分，预设项禁用删除按钮
    - 创建 `frontend/src/components/servers/AddServerDialog.tsx`：协议选择（UDP/DoT/DoH）、地址输入、可选 TLS/Bootstrap 字段、格式验证和错误提示
    - _Requirements: 1.4, 1.5, 1.6, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

  - [x] 11.3 实现测试域名列表和添加对话框
    - 创建 `frontend/src/components/domains/DomainList.tsx`：展示域名列表，预设项禁用删除按钮
    - 创建 `frontend/src/components/domains/AddDomainDialog.tsx`：域名输入、格式验证和错误提示
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [x] 11.4 实现测试参数配置表单
    - 创建 `frontend/src/components/params/TestParamsForm.tsx`
    - 表单字段：queries、warmup、concurrency、timeout，带输入验证和错误提示
    - _Requirements: 7.1, 7.2, 7.3, 7.4_

  - [x] 11.5 实现配置导入导出工具栏
    - 创建 `frontend/src/components/config/ConfigToolbar.tsx`
    - "导入配置"按钮：调用文件选择对话框 + `ImportConfig`，错误时显示详细提示
    - "导出配置"按钮：调用文件保存对话框 + `ExportConfig`
    - _Requirements: 13.3, 13.4, 13.5_

- [x] 12. 实现前端测试执行和结果展示组件
  - [x] 12.1 实现测试控制组件
    - 创建 `frontend/src/components/benchmark/BenchmarkControl.tsx`
    - "开始测试"/"停止测试"按钮切换，进度条显示测试进度和状态
    - _Requirements: 8.1, 8.4, 8.5_

  - [x] 12.2 实现测试结果表格
    - 创建 `frontend/src/components/benchmark/ResultsTable.tsx`
    - 表格列：服务器名称、协议、中位延迟、P95 延迟、成功率、答案不一致数、Score
    - 按 Score 降序排列，最高分绿色标注，超时/Score=0 红色标注
    - 答案不一致 > 0 时显示警告标识
    - 无历史结果时显示"尚未进行测试"提示
    - _Requirements: 9.5, 10.1, 10.2, 10.3, 14.3_

  - [x] 12.3 实现 Score 柱状图
    - 创建 `frontend/src/components/benchmark/ScoreChart.tsx`
    - 使用 Recharts 绘制柱状图，鼠标悬停显示详细数据 Tooltip
    - _Requirements: 10.5, 10.6_

  - [x] 12.4 实现推荐 DNS 展示组件
    - 创建 `frontend/src/components/benchmark/Recommendation.tsx`
    - 在结果区域顶部显示推荐的 DNS 服务器（Score 最高）
    - _Requirements: 10.4_

- [x] 13. 实现前端系统 DNS 配置组件
  - [x] 13.1 实现当前 DNS 配置展示
    - 创建 `frontend/src/components/dns-config/CurrentDNSDisplay.tsx`
    - 展示所有活动网络适配器及其当前 DNS 配置
    - _Requirements: 11.1_

  - [x] 13.2 实现应用 DNS 对话框
    - 创建 `frontend/src/components/dns-config/ApplyDNSDialog.tsx`
    - 从测试结果选择 DNS 服务器后，弹出网卡选择对话框，调用 `ApplyDNS`
    - 无管理员权限时提示用户以管理员身份重新运行
    - 成功/失败时显示对应提示
    - _Requirements: 11.2, 11.3, 11.4, 11.5, 11.6_

  - [x] 13.3 实现恢复 DHCP 按钮
    - 创建 `frontend/src/components/dns-config/RestoreDHCPButton.tsx`
    - 点击后选择网卡并调用 `RestoreDHCP`
    - _Requirements: 11.7_

- [x] 14. Checkpoint - 确保前后端集成正常
  - 确保所有测试通过，如有问题请询问用户。

- [x] 15. 最终集成和完善
  - [x] 15.1 整合所有组件到 App.tsx
    - 在 `App.tsx` 中组装 MainLayout 和所有子组件
    - 实现应用启动初始化流程：加载配置 → 加载预设 → 加载历史结果 → 检测主题 → 检测管理员权限
    - _Requirements: 12.1, 12.4, 13.7_

  - [x] 15.2 实现自动保存逻辑
    - 在前端 Store 的服务器/域名增删 action 中，调用后端自动保存配置
    - 测试完成后自动保存结果到本地存储
    - _Requirements: 13.6, 14.1_

  - [x] 15.3 集成测试：完整 Benchmark 流程
    - 使用少量服务器和域名执行完整测试流程
    - 验证进度事件推送、结果计算、结果持久化
    - _Requirements: 8.1, 8.4, 8.6, 14.1_

- [x] 16. 最终 Checkpoint - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户。

## Notes

- 标记 `*` 的任务为可选任务，可跳过以加速 MVP 开发
- 每个任务引用了具体的需求编号，确保需求可追溯
- Checkpoint 任务用于阶段性验证，确保增量开发的正确性
- 属性测试验证设计文档中定义的 9 个 Correctness Properties
- 单元测试覆盖具体场景和边界条件
- 后端使用 Go `rapid` 库进行属性测试，前端测试为可选

# DNS Selector GUI 问题清单与修复落地方案

本文档把当前审查发现的问题整理为可执行的修复清单，按优先级排序，并给出影响范围、具体改法、验收标准和建议拆分方式。

## P0

### 1. `DoT/DoH` 结果可测速但不可稳定“应用到系统”

- 优先级：P0
- 影响范围：系统 DNS 应用链路、结果推荐、README 功能承诺
- 问题概述：
  当前“推荐 DNS”与“结果表应用”会直接把测速结果里的 `address` 用于系统 DNS 设置。对于 `DoT` 域名、`DoH URL`、`DoH host` 这类值，很多平台并不能把它们当作系统 DNS 服务器地址稳定接受。现状会导致“测速第一名”不一定能真正应用成功，或应用后行为不符合用户预期。
- 受影响位置：
  - [backend/dns_config_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/dns_config_service.go:40)
  - [backend/platform_darwin.go](/Users/zane/Documents/go_project/dns-selector-win/backend/platform_darwin.go:88)
  - [backend/platform_linux.go](/Users/zane/Documents/go_project/dns-selector-win/backend/platform_linux.go:69)
  - [backend/platform_windows.go](/Users/zane/Documents/go_project/dns-selector-win/backend/platform_windows.go:138)
  - [frontend/src/components/benchmark/Recommendation.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/benchmark/Recommendation.tsx:15)
  - [frontend/src/components/benchmark/ResultsTable.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/benchmark/ResultsTable.tsx:82)
- 建议修复方向：
  - 明确区分“测速 endpoint”和“系统可应用的 DNS 地址”。
  - 在后端新增统一转换逻辑，例如 `ResolveSystemDNSAddress(server DNSServer) ([]string, error)`。
  - 转换规则建议：
    - `UDP`：直接使用 IPv4。
    - `DoT`：
      - 若 `Address` 本身是 IP，可直接使用。
      - 若是域名，优先使用 `BootstrapIPs`。
      - 若无 `BootstrapIPs`，不要直接应用，返回明确错误。
    - `DoH`：
      - 若 URL host 是 IP，可直接使用该 IP。
      - 若 URL host 是域名，优先使用 `BootstrapIPs`。
      - 若无 `BootstrapIPs`，不要直接应用，返回明确错误。
  - 前端对不可应用的结果禁用“应用”按钮，并在 UI 上提示“该结果可测速，但不能直接写入系统 DNS”。
  - README 中把“支持测试 DoT/DoH”与“系统 DNS 写入能力”分开描述，避免误导。
- 验收标准：
  - 对所有内置预设服务器，点击“应用”时行为可预测。
  - 没有可用系统 DNS 地址的 `DoT/DoH` 项不能进入应用流程。
  - 用户能在 UI 上看到为什么不能应用。
  - 至少新增单测覆盖以下输入：
    - `dns.google` + `BootstrapIPs`
    - `https://dns.google/dns-query` + `BootstrapIPs`
    - `https://1.1.1.1/dns-query`
    - 无 bootstrap 的域名型 `DoT/DoH`

### 2. 切换预设会清空并持久化删除所有自定义内容

- 优先级：P0
- 影响范围：用户数据、配置保存、预设切换交互
- 问题概述：
  当前切换 `CN / Global` 时会清空自定义服务器和域名，并立即自动保存。这会导致用户在无提示的情况下永久丢失自定义数据。
- 受影响位置：
  - [backend/preset_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/preset_service.go:30)
  - [backend/app_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/app_service.go:101)
  - [frontend/src/components/preset/PresetSelector.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/preset/PresetSelector.tsx:18)
- 建议修复方向：
  - 调整数据模型：预设切换只切换“当前激活 preset”，不应清空 `customServers/customDomains`。
  - 如果业务上确实需要不同预设绑定不同自定义内容，应改为“按 preset 分桶保存”，而不是切换时删除。
  - 前端在任何会导致数据替换的动作前增加确认弹窗。
  - 配置结构可升级为：
    - 方案 A：全局自定义项始终保留，切换预设仅影响内置列表。
    - 方案 B：`customServersByPreset` / `customDomainsByPreset`。
  - 推荐方案 A，改动更小，用户心智更简单。
- 验收标准：
  - 切换预设前后，自定义服务器/域名不会丢失。
  - 导出配置后再导入，行为与 UI 一致。
  - 为“切换预设不丢自定义内容”补单测。

### 3. 服务器唯一标识设计错误，导致删除、判重、渲染都不可靠

- 优先级：P0
- 影响范围：服务器管理、前端列表渲染、预设判断
- 问题概述：
  当前多处把 `address` 当作服务器唯一标识。但同一地址可对应不同协议，甚至不同 `TLSServerName`。这会导致：
  - React key 冲突
  - 误判为预设项
  - 删除操作删错对象或删不掉
- 受影响位置：
  - [backend/preset_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/preset_service.go:59)
  - [backend/preset_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/preset_service.go:83)
  - [backend/app_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/app_service.go:110)
  - [frontend/src/components/servers/ServerList.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/servers/ServerList.tsx:57)
- 建议修复方向：
  - 为服务器引入稳定 ID。
  - 最小改法：
    - 定义 server key：`protocol + "|" + address + "|" + tlsServerName`
    - 所有判重、删除、预设判断、前端 key 都改用这个 key。
  - 更长期的做法：
    - 在模型层显式增加 `ID string`。
    - 预设项和自定义项都携带 `ID`。
  - `RemoveCustomServer` 不再接收裸 `address`，改接收 `serverKey` 或 `ID`。
- 验收标准：
  - 同一 address 下添加 `UDP` 与 `DoT` 两条记录时，前端列表稳定。
  - 删除某一条时不会误删同地址的另一条。
  - 预设项和自定义项不会因地址相同而互相污染。

## P1

### 4. 配置加载失败会静默回退默认值，容易造成“无声重置”

- 优先级：P1
- 影响范围：启动流程、配置可靠性、用户信任
- 问题概述：
  配置文件不存在、权限错误、内容损坏都被统一处理成“返回默认配置且不报错”。这会掩盖真实故障，也会让用户误以为配置被随机清空。
- 受影响位置：
  - [backend/config_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/config_service.go:52)
  - [backend/app_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/app_service.go:49)
- 建议修复方向：
  - 区分三类场景：
    - 文件不存在：允许回退默认值。
    - 文件存在但 JSON 损坏：返回错误，并让 UI 提示“配置损坏，是否重置”。
    - 文件存在但权限不足/读取失败：返回错误，不要默默覆盖。
  - `Load()` 返回更精确的错误类型。
  - `OnStartup()` 不要忽略错误，应记录并通过前端可见提示展示。
  - 对结果文件 `LoadResults()` 也采用类似策略。
- 验收标准：
  - 配置损坏时，用户能看到明确提示。
  - 权限错误时不会自动恢复默认并继续覆盖原配置。
  - 单测覆盖文件不存在、权限错误、JSON 损坏三类情形。

### 5. 自定义服务器附加字段缺少校验，错误暴露过晚

- 优先级：P1
- 影响范围：添加服务器、导入配置、测速与应用流程
- 问题概述：
  当前只验证主地址，`TLSServerName` 与 `BootstrapIP` 几乎不校验。无效值会进入配置，直到后续使用时才失败。
- 受影响位置：
  - [backend/app_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/app_service.go:141)
  - [backend/config_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/config_service.go:183)
  - [backend/validation.go](/Users/zane/Documents/go_project/dns-selector-win/backend/validation.go:14)
  - [frontend/src/components/servers/AddServerDialog.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/servers/AddServerDialog.tsx:24)
- 建议修复方向：
  - 增加结构化校验函数，例如 `ValidateServerEntry(req AddServerRequest) error`。
  - 规则建议：
    - `BootstrapIP` 必须是合法 IPv4。
    - `TLSServerName` 必须是合法域名。
    - `DoT` 的 `IP@TLSServerName` 格式与单独填写的 `TLSServerName` 不能互相冲突。
    - `DoH` 若 host 是 IP 且要走 TLS 名称校验，可允许 `TLSServerName`；若 host 是域名且又填写了不同 `TLSServerName`，需报错或明确规范。
  - 前端添加即时表单校验，而不是只依赖后端报错。
- 验收标准：
  - 无效 `BootstrapIP` 无法保存。
  - 无效 `TLSServerName` 无法保存。
  - 导入配置时也会执行同样的完整校验。

### 6. 测试参数每次输入都会触发保存，存在不必要的高频 I/O

- 优先级：P1
- 影响范围：前端输入体验、Wails 调用频率、配置文件写入频率
- 问题概述：
  用户在编辑参数输入框时，只要当前值暂时合法，就会立即调用后端保存配置。这会造成频繁状态同步和磁盘写入。
- 受影响位置：
  - [frontend/src/components/params/TestParamsForm.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/params/TestParamsForm.tsx:49)
  - [backend/app_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/app_service.go:198)
- 建议修复方向：
  - 前端改为“本地编辑态 + 显式保存”。
  - 若保留自动保存，至少做 300ms 到 800ms debounce。
  - 后端可考虑对相同值短路，避免重复落盘。
  - 保存失败时应回滚 UI 状态或给出提示。
- 验收标准：
  - 连续输入时不会每击键写一次配置。
  - 保存失败时用户可感知。
  - DevTools 或日志可验证调用次数明显下降。

### 7. 测试依赖真实用户目录和外部环境，不够可复现

- 优先级：P1
- 影响范围：CI、本地测试稳定性、沙箱执行
- 问题概述：
  集成测试会依赖真实 `os.UserConfigDir()` 结果，导致在不同平台或受限环境下表现不一致。我本地实测 `go test ./backend/...` 会失败，而改变 `HOME` 后又能通过。
- 受影响位置：
  - [backend/integration_test.go](/Users/zane/Documents/go_project/dns-selector-win/backend/integration_test.go:72)
  - [backend/config_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/config_service.go:37)
  - [Makefile](/Users/zane/Documents/go_project/dns-selector-win/Makefile:94)
- 建议修复方向：
  - 给 `ConfigService` 注入路径提供者，避免测试写真实用户目录。
  - 将结果路径和配置路径抽成可替换依赖，而不是硬编码调用 `defaultResultsPath()`。
  - 把“需要真实网络”的测试显式标为 integration，并默认不进 `make test`。
  - `make test` 默认跑纯单元测试。
  - 另加 `make test-integration` 跑环境敏感项。
- 验收标准：
  - `go test ./backend/... -short` 在干净环境稳定通过。
  - 默认测试命令不写用户真实配置目录。
  - CI 上无需依赖本机 HOME 结构。

### 8. 前端错误处理大量吞掉，只在控制台打印

- 优先级：P1
- 影响范围：可用性、问题定位、用户反馈
- 问题概述：
  多个关键交互失败后只写 `console.error`，用户看不到失败原因，只会看到空状态或按钮无响应。
- 受影响位置：
  - [frontend/src/store/useAppStore.ts](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/store/useAppStore.ts:82)
  - [frontend/src/components/benchmark/BenchmarkControl.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/benchmark/BenchmarkControl.tsx:10)
  - [frontend/src/components/preset/PresetSelector.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/preset/PresetSelector.tsx:18)
  - [frontend/src/components/domains/DomainList.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/domains/DomainList.tsx:16)
  - [frontend/src/components/servers/ServerList.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/servers/ServerList.tsx:28)
  - [frontend/src/components/dns-config/CurrentDNSDisplay.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/dns-config/CurrentDNSDisplay.tsx:10)
- 建议修复方向：
  - 引入统一的全局反馈机制，例如 toast 或页面级 error banner。
  - 所有异步操作至少做到：
    - 用户可见错误信息
    - 重试入口
    - 空状态与错误状态分离
  - `GetNetworkAdapters` 失败时不应显示“无活动网络适配器”，而应显示“获取失败”。
- 验收标准：
  - 每个失败分支都有可见反馈。
  - 空列表与加载失败可以在 UI 上区分。

## P2

### 9. README 与实际实现存在明显漂移

- 优先级：P2
- 影响范围：新用户接入、贡献者理解、发布预期
- 问题概述：
  文档里有多处信息已经与仓库现状不一致，包括 Go 版本、项目目录、CI 文件、权限说明等。
- 受影响位置：
  - [README.md](/Users/zane/Documents/go_project/dns-selector-win/README.md:35)
  - [README.md](/Users/zane/Documents/go_project/dns-selector-win/README.md:45)
  - [README.md](/Users/zane/Documents/go_project/dns-selector-win/README.md:71)
  - [README.md](/Users/zane/Documents/go_project/dns-selector-win/README.md:141)
  - [go.mod](/Users/zane/Documents/go_project/dns-selector-win/go.mod:3)
- 具体不一致项：
  - README 写 `Go 1.21+`，但 `go.mod` 要求 `1.25.0`。
  - README 示例目录名与当前仓库目录不一致。
  - README 提到了 `.github/workflows/build.yml`，仓库中不存在。
  - README 的 macOS 权限描述与代码实现不一致。
  - README 写“支持 Windows/macOS/Linux 三平台（amd64 + arm64）”，但实际是否都持续构建、测试、打包，文档未给证据。
- 建议修复方向：
  - 以代码为准更新 README。
  - 对尚未完全验证的平台能力明确标注“计划支持”或“已验证平台”。
  - 文档里补充：
    - 默认测试范围
    - 结果文件路径
    - 预设切换行为
    - `DoT/DoH` 的“测速”和“系统应用”能力区别
- 验收标准：
  - README 能被新开发者直接照着跑通。
  - 文档中的文件路径、命令、版本要求、平台说明与仓库现状一致。

### 10. macOS 权限检测与错误提示逻辑不一致

- 优先级：P2
- 影响范围：DNS 修改交互、用户认知
- 问题概述：
  macOS 平台 `CheckAdmin()` 直接返回 `true`，但 `SetDNS()`/`ResetToAuto()` 里又会根据命令失败提示“需要管理员权限”。这两套语义互相冲突。
- 受影响位置：
  - [backend/platform_darwin.go](/Users/zane/Documents/go_project/dns-selector-win/backend/platform_darwin.go:140)
  - [frontend/src/components/dns-config/ApplyDNSDialog.tsx](/Users/zane/Documents/go_project/dns-selector-win/frontend/src/components/dns-config/ApplyDNSDialog.tsx:61)
- 建议修复方向：
  - 明确 macOS 的真实权限模型。
  - 如果普通用户可用，则不要在失败时统一归因为“需要管理员权限”。
  - 如果某些网络服务修改仍需要授权，应引入更准确的错误分类。
  - `IsAdmin` 在 macOS 上可能不该暴露成简单布尔值。
- 验收标准：
  - macOS 上 UI 提示与真实失败原因一致。
  - 不再出现“前端显示已具备权限，但执行时报权限不足”这种认知冲突。

### 11. 配置服务接口设计不利于测试和扩展

- 优先级：P2
- 影响范围：测试、未来功能扩展、代码可维护性
- 问题概述：
  `ConfigService` 同时负责路径决策、序列化、磁盘读写、默认值策略，耦合度偏高。结果是任何测试都难以只替换其中一层能力。
- 受影响位置：
  - [backend/config_service.go](/Users/zane/Documents/go_project/dns-selector-win/backend/config_service.go:11)
- 建议修复方向：
  - 引入小接口拆分：
    - `PathProvider`
    - `ConfigStore`
    - `ResultsStore`
  - 默认实现走文件系统；测试实现走内存。
  - `AppService` 通过构造函数注入依赖，而不是内部写死 `NewConfigService()`。
- 验收标准：
  - 测试无需依赖环境变量篡改路径。
  - `AppService` 可以用 fake config service 单测。

### 12. 默认测试命令没有明确区分“单元测试”和“外部依赖测试”

- 优先级：P2
- 影响范围：开发体验、CI 时长、误报
- 问题概述：
  当前 `make test` 直接跑 `go test ./backend/... -v -count=1`，把所有后端测试都混在一起。对贡献者来说，很难知道哪些测试依赖网络、文件系统或平台环境。
- 受影响位置：
  - [Makefile](/Users/zane/Documents/go_project/dns-selector-win/Makefile:94)
  - [README.md](/Users/zane/Documents/go_project/dns-selector-win/README.md:126)
- 建议修复方向：
  - 拆成：
    - `make test-unit`
    - `make test-property`
    - `make test-integration`
    - `make test`
  - `make test` 默认只跑稳定、无外部依赖的集合。
  - README 同步说明。
- 验收标准：
  - 新开发者无需了解项目内部细节，也知道该先跑什么。
  - 默认测试命令稳定且快速。

## 建议修复顺序

1. 修 `DoT/DoH` 应用链路与 UI 限制。
2. 修预设切换导致的数据丢失。
3. 修服务器唯一标识与删除逻辑。
4. 修配置加载错误处理。
5. 修完整输入校验。
6. 修测试隔离与命令拆分。
7. 修前端错误反馈。
8. 最后统一整理 README 与工程文档。

## 建议拆分为 Issue/PR 的方式

### PR 1

- 标题建议：修复系统 DNS 应用链路对 `DoT/DoH` 的错误处理
- 范围：
  - 后端地址转换逻辑
  - 前端“应用”按钮可用性判断
  - 单元测试补齐

### PR 2

- 标题建议：修复预设切换导致的自定义配置丢失
- 范围：
  - `PresetService`
  - 配置持久化结构
  - 前端切换交互

### PR 3

- 标题建议：为 DNS 服务器引入稳定标识并修复删除/判重逻辑
- 范围：
  - 服务端模型
  - 前端 key 与删除接口
  - 回归测试

### PR 4

- 标题建议：改进配置加载与错误反馈机制
- 范围：
  - `ConfigService`
  - `AppService` 启动流程
  - 前端全局错误提示

### PR 5

- 标题建议：隔离测试环境并重构测试命令
- 范围：
  - 配置路径注入
  - `Makefile`
  - README 测试说明

### PR 6

- 标题建议：补全文档并校正 README 与实现不一致的问题
- 范围：
  - README
  - 发布说明
  - 平台支持说明

## 备注

- 当前最不建议继续放着不管的是：
  - `DoT/DoH` 结果可推荐但不可稳定应用
  - 切换预设直接抹掉用户自定义数据
- 这两个问题都属于“用户看起来像正常操作，结果却丢数据或行为失真”的类型，优先级应高于性能和文档问题。

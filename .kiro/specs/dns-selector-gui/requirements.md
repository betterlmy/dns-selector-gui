# Requirements Document

## Introduction

本文档定义了 Windows DNS 择优器可视化桌面应用（DNS Selector GUI）的需求。该应用是开源 CLI 项目 dns-selector 的 Windows 可视化版本，为用户提供图形化界面，用于测试多个 DNS 服务器（支持 UDP/DoT/DoH 多协议）的响应速度、稳定性和正确性，帮助用户选择最优的 DNS 服务器，并支持直接修改 Windows 系统的 DNS 配置。应用内置 CN 和 Global 两套预设方案，同时支持通过 JSON 配置文件进行完全自定义。

## Glossary

- **DNS_Selector_GUI**: Windows DNS 择优器的可视化桌面应用程序主体
- **DNS_Server**: 域名系统服务器，负责将域名解析为 IP 地址的网络服务，支持 UDP、DoT、DoH 三种协议
- **DNS_Tester**: 应用中负责对 DNS 服务器执行延迟测试、答案验证和评分计算的模块
- **DNS_Server_List**: 应用中维护的待测试 DNS 服务器列表，包含预置和用户自定义的服务器
- **Latency_Result**: DNS 测试返回的响应延迟数据，以毫秒（ms）为单位
- **Network_Adapter**: Windows 系统中的网络适配器（网卡），DNS 设置绑定在网络适配器上
- **DNS_Configurator**: 应用中负责读取和修改 Windows 系统 DNS 设置的模块
- **Test_Domain**: 用于测试 DNS 解析速度的目标域名
- **Preset**: 预设方案，包含一组 DNS 服务器列表和测试域名列表的完整配置
- **CN_Preset**: 面向中国大陆用户的预设方案，包含 32 个 DNS 服务器和 29 个测试域名
- **Global_Preset**: 面向全球用户的预设方案，包含 16 个 DNS 服务器和 24 个测试域名
- **UDP_DNS**: 使用传统 UDP 协议（端口 53）的 DNS 服务器
- **DoT_DNS**: 使用 DNS-over-TLS 协议（端口 853）的 DNS 服务器
- **DoH_DNS**: 使用 DNS-over-HTTPS 协议的 DNS 服务器
- **TLS_Server_Name**: DoT/DoH 连接中用于 TLS 验证的服务器名称
- **Bootstrap_IP**: 用于解析 DoT/DoH 服务器域名的引导 IP 地址
- **Score**: 综合评分，基于中位延迟、成功率和抖动计算的 DNS 服务器质量指标
- **Median_Latency**: 中位延迟，所有查询延迟的中位数
- **P95_Latency**: P95 延迟，95% 的查询延迟低于此值
- **Success_Rate**: 成功率，验证通过的查询占总查询数的比例
- **Answer_Mismatch**: 答案不一致，DNS 服务器返回的解析结果与多数服务器共识不一致的次数
- **Config_File**: JSON 格式的配置文件，用于持久化用户自定义的服务器、域名和测试参数

## Requirements

### Requirement 1: 多协议 DNS 服务器支持

**User Story:** 作为用户，我希望能够测试 UDP、DoT 和 DoH 三种协议的 DNS 服务器，以便全面评估不同协议下的 DNS 性能。

#### Acceptance Criteria

1. THE DNS_Tester SHALL 支持通过 UDP 协议（端口 53）查询 UDP_DNS 服务器
2. THE DNS_Tester SHALL 支持通过 DNS-over-TLS 协议（端口 853）查询 DoT_DNS 服务器
3. THE DNS_Tester SHALL 支持通过 DNS-over-HTTPS 协议查询 DoH_DNS 服务器
4. WHEN 用户添加 DoT_DNS 服务器时, THE DNS_Selector_GUI SHALL 允许用户指定可选的 TLS_Server_Name 和 Bootstrap_IP
5. WHEN 用户添加 DoH_DNS 服务器时, THE DNS_Selector_GUI SHALL 允许用户指定可选的 TLS_Server_Name 和 Bootstrap_IP
6. THE DNS_Selector_GUI SHALL 在服务器列表中以标签形式区分显示 UDP、DoT、DoH 三种协议类型

### Requirement 2: 预设方案管理

**User Story:** 作为用户，我希望能够在 CN 和 Global 两套预设方案之间切换，以便根据所在地区快速选择合适的 DNS 服务器和测试域名。

#### Acceptance Criteria

1. THE DNS_Selector_GUI SHALL 内置 CN_Preset 和 Global_Preset 两套预设方案
2. THE DNS_Selector_GUI SHALL 提供预设方案选择控件，允许用户在 CN_Preset 和 Global_Preset 之间切换
3. WHEN 用户选择 CN_Preset, THE DNS_Selector_GUI SHALL 加载 CN 预设的 32 个 DNS 服务器和 29 个测试域名
4. WHEN 用户选择 Global_Preset, THE DNS_Selector_GUI SHALL 加载 Global 预设的 16 个 DNS 服务器和 24 个测试域名
5. WHEN 用户切换预设方案, THE DNS_Selector_GUI SHALL 替换当前的 DNS_Server_List 和 Test_Domain 列表为所选预设的内容
6. THE DNS_Selector_GUI SHALL 禁止用户删除预设方案中的 DNS 服务器和测试域名


### Requirement 3: CN 预设内容

**User Story:** 作为中国大陆用户，我希望应用内置完整的国内常用 DNS 服务器和热门域名，以便开箱即用地评估 DNS 性能。

#### Acceptance Criteria

1. THE CN_Preset SHALL 包含以下 UDP_DNS 服务器：AliDNS（223.5.5.5, 223.6.6.6）、BaiduDNS（180.76.76.76）、DNSPod（119.28.28.28, 119.29.29.29）、114DNS（114.114.114.114, 114.114.115.115）、114DNS Safe（114.114.114.119, 114.114.115.119）、114DNS Family（114.114.114.110, 114.114.115.110）、Bytedance（180.184.1.1, 180.184.2.2）、Google（8.8.8.8, 8.8.4.4）、Cloudflare（1.1.1.1, 1.0.0.1）、Freenom（80.80.80.80, 80.80.81.81）
2. THE CN_Preset SHALL 包含以下 DoT_DNS 服务器：AliDNS（dns.alidns.com）、DNSPod（dot.pub）、Google（dns.google）、Cloudflare（1.1.1.1, one.one.one.one）
3. THE CN_Preset SHALL 包含以下 DoH_DNS 服务器：AliDNS（dns.alidns.com 的 3 个 DoH 端点）、DNSPod（doh.pub）、Cloudflare（3 个 DoH 端点）、Google（dns.google）
4. THE CN_Preset SHALL 包含以下 29 个测试域名：douyin.com, kuaishou.com, baidu.com, taobao.com, mi.com, aliyun.com, bilibili.com, jd.com, qq.com, ithome.com, hupu.com, feishu.cn, sohu.com, 163.com, sina.com, weibo.com, xiaohongshu.com, douban.com, zhihu.com, youku.com, youdao.com, mp.weixin.qq.com, iqiyi.com, v.qq.com, y.qq.com, www.ctrip.com, autohome.com.cn, apple.com, github.com, bing.com

### Requirement 4: Global 预设内容

**User Story:** 作为全球用户，我希望应用内置国际主流 DNS 服务器和热门域名，以便评估全球范围内的 DNS 性能。

#### Acceptance Criteria

1. THE Global_Preset SHALL 包含以下 UDP_DNS 服务器：Google（8.8.8.8, 8.8.4.4）、Cloudflare（1.1.1.1, 1.0.0.1）、Quad9（9.9.9.9, 149.112.112.112）、OpenDNS（208.67.222.222, 208.67.220.220）
2. THE Global_Preset SHALL 包含以下 DoT_DNS 服务器：Google（dns.google）、Cloudflare（one.one.one.one）、Quad9（dns.quad9.net）
3. THE Global_Preset SHALL 包含以下 DoH_DNS 服务器：Google（dns.google）、Cloudflare（cloudflare-dns.com）、Quad9（dns.quad9.net）
4. THE Global_Preset SHALL 包含以下 24 个测试域名：google.com, youtube.com, github.com, wikipedia.org, cloudflare.com, amazon.com, openai.com, chatgpt.com, apple.com, microsoft.com, reddit.com, netflix.com, bbc.com, nytimes.com, linkedin.com, instagram.com, x.com, tiktok.com, discord.com, zoom.us, dropbox.com, ubuntu.com, mozilla.org, stackoverflow.com

### Requirement 5: DNS 服务器列表管理

**User Story:** 作为用户，我希望能够管理待测试的 DNS 服务器列表，以便灵活添加和移除要评估的 DNS 服务器。

#### Acceptance Criteria

1. THE DNS_Selector_GUI SHALL 在启动时显示当前预设方案的 DNS_Server_List
2. WHEN 用户点击"添加 DNS"按钮, THE DNS_Selector_GUI SHALL 显示添加对话框，允许用户选择协议类型（UDP/DoT/DoH）并输入服务器地址
3. WHEN 用户添加 UDP_DNS 服务器时, THE DNS_Selector_GUI SHALL 验证输入格式为有效的 IPv4 地址
4. WHEN 用户添加 DoT_DNS 服务器时, THE DNS_Selector_GUI SHALL 验证输入格式为有效的域名或 "IP@TLSServerName" 格式
5. WHEN 用户添加 DoH_DNS 服务器时, THE DNS_Selector_GUI SHALL 验证输入格式为有效的 HTTPS URL 或 "https://IP/dns-query@TLSServerName" 格式
6. WHEN 用户输入的服务器地址格式无效, THE DNS_Selector_GUI SHALL 显示格式错误提示信息，并说明正确的格式
7. WHEN 用户选择列表中的自定义 DNS 服务器并点击"删除"按钮, THE DNS_Server_List SHALL 从列表中移除该服务器

### Requirement 6: 测试域名管理

**User Story:** 作为用户，我希望能够管理测试域名列表，以便使用与自身使用场景相关的域名进行测试。

#### Acceptance Criteria

1. THE DNS_Selector_GUI SHALL 在启动时显示当前预设方案的 Test_Domain 列表
2. WHEN 用户点击"添加域名"按钮并输入有效的域名, THE DNS_Selector_GUI SHALL 将该域名添加到 Test_Domain 列表中
3. WHEN 用户输入的域名格式无效, THE DNS_Selector_GUI SHALL 显示格式错误提示信息
4. WHEN 用户选择列表中的自定义 Test_Domain 并点击"删除"按钮, THE DNS_Selector_GUI SHALL 从列表中移除该域名

### Requirement 7: 测试参数配置

**User Story:** 作为用户，我希望能够调整测试参数，以便根据需要控制测试的精度和速度。

#### Acceptance Criteria

1. THE DNS_Selector_GUI SHALL 提供以下可配置的测试参数：每域名正式查询次数（queries）、每服务器预热查询次数（warmup）、最大并发查询数（concurrency）、单次查询超时时间（timeout）
2. THE DNS_Selector_GUI SHALL 为测试参数提供以下默认值：queries 为 10 次、warmup 为 1 次、concurrency 为 20、timeout 为 2 秒
3. WHEN 用户修改测试参数, THE DNS_Selector_GUI SHALL 验证输入值为正整数（timeout 为正数），并在无效时显示错误提示
4. THE DNS_Selector_GUI SHALL 在测试控制区域以表单形式展示测试参数，允许用户在开始测试前调整


### Requirement 8: DNS 延迟测试与评分

**User Story:** 作为用户，我希望能够对 DNS 服务器进行全面的性能测试和评分，以便基于综合指标选择最优的 DNS 服务器。

#### Acceptance Criteria

1. WHEN 用户点击"开始测试"按钮, THE DNS_Tester SHALL 对 DNS_Server_List 中的每个 DNS_Server 执行预热查询（warmup 次数），然后执行正式查询（queries 次数 × Test_Domain 数量）
2. THE DNS_Tester SHALL 以不超过 concurrency 设定值的并发数执行 DNS 查询
3. THE DNS_Tester SHALL 为每次 DNS 查询设置 timeout 参数指定的超时时间
4. WHILE DNS_Tester 正在执行测试, THE DNS_Selector_GUI SHALL 显示测试进度条和当前测试状态
5. WHILE DNS_Tester 正在执行测试, THE DNS_Selector_GUI SHALL 提供"停止测试"按钮以允许用户中断测试
6. WHEN 测试完成, THE DNS_Tester SHALL 为每个 DNS_Server 计算以下指标：Median_Latency（中位延迟）、P95_Latency（P95 延迟）、Success_Rate（验证后成功率）、Answer_Mismatch（答案不一致数）、Score（综合评分）
7. THE DNS_Tester SHALL 使用以下公式计算 Score：Score = (1 / median_seconds) × (success_rate²) × (median / P95)，其中 median_seconds 为中位延迟（秒），success_rate 为验证后成功率，median/P95 为中位延迟与 P95 延迟的比值
8. WHEN 有效样本数少于 5 个, THE DNS_Tester SHALL 在 Score 计算中省略 median/P95 抖动惩罚因子
9. IF 某个 DNS_Server 在测试中所有查询均超时, THEN THE DNS_Tester SHALL 将该服务器的 Score 设为 0 并标记为"超时"

### Requirement 9: 答案验证机制

**User Story:** 作为用户，我希望应用能够验证 DNS 解析结果的正确性，以便排除返回异常结果的 DNS 服务器。

#### Acceptance Criteria

1. THE DNS_Tester SHALL 对所有 DNS_Server 返回的解析结果执行多服务器共识验证
2. WHEN 多个 DNS_Server 对同一 Test_Domain 返回不同的 DNS 指纹, THE DNS_Tester SHALL 仅接受被多个服务器确认的指纹作为有效结果
3. THE DNS_Tester SHALL 将一次性异常指纹（仅被单个服务器返回的指纹）标记为无效并排除
4. THE DNS_Tester SHALL 区分原始成功数和验证后成功数，并在结果中同时展示
5. WHEN 某个 DNS_Server 的 Answer_Mismatch 数量大于 0, THE DNS_Selector_GUI SHALL 在结果中以警告标识突出显示

### Requirement 10: 测试结果展示

**User Story:** 作为用户，我希望能够直观地查看 DNS 测试结果的全面对比，以便快速识别最优的 DNS 服务器。

#### Acceptance Criteria

1. WHEN 测试完成, THE DNS_Selector_GUI SHALL 以表格形式展示每个 DNS_Server 的以下指标：服务器名称、协议类型、Median_Latency、P95_Latency、Success_Rate、Answer_Mismatch、Score
2. WHEN 测试完成, THE DNS_Selector_GUI SHALL 按照 Score 从高到低对测试结果进行排序
3. THE DNS_Selector_GUI SHALL 使用绿色标注 Score 最高的 DNS_Server，使用红色标注超时或 Score 为 0 的 DNS_Server
4. WHEN 测试完成, THE DNS_Selector_GUI SHALL 在测试结果区域顶部显示推荐使用的 DNS_Server（即 Score 最高的服务器）
5. THE DNS_Selector_GUI SHALL 以柱状图形式展示各 DNS_Server 的 Score 对比
6. WHEN 用户将鼠标悬停在柱状图的某个柱体上, THE DNS_Selector_GUI SHALL 显示该 DNS_Server 的详细测试数据（Median_Latency、P95_Latency、Success_Rate、Answer_Mismatch、Score）

### Requirement 11: Windows DNS 设置修改

**User Story:** 作为用户，我希望能够通过应用直接修改 Windows 系统的 DNS 设置，以便快速应用最优的 DNS 配置。

#### Acceptance Criteria

1. THE DNS_Configurator SHALL 读取当前系统所有活动 Network_Adapter 的 DNS 配置并在界面中显示
2. WHEN 用户从测试结果中选择一个 DNS_Server 并点击"应用 DNS"按钮, THE DNS_Configurator SHALL 将该 DNS_Server 设置为指定 Network_Adapter 的首选 DNS
3. WHEN 用户点击"应用 DNS"按钮, THE DNS_Selector_GUI SHALL 显示 Network_Adapter 选择对话框，允许用户选择要修改的网络适配器
4. IF DNS_Configurator 没有管理员权限, THEN THE DNS_Selector_GUI SHALL 提示用户以管理员身份重新运行应用
5. WHEN DNS 设置修改成功, THE DNS_Selector_GUI SHALL 显示成功提示并刷新当前 DNS 配置显示
6. IF DNS 设置修改失败, THEN THE DNS_Selector_GUI SHALL 显示包含失败原因的错误提示
7. WHEN 用户点击"恢复默认 DNS"按钮, THE DNS_Configurator SHALL 将指定 Network_Adapter 的 DNS 设置恢复为自动获取（DHCP）

### Requirement 12: 应用界面与交互

**User Story:** 作为用户，我希望应用界面简洁直观，以便高效地完成 DNS 择优和配置操作。

#### Acceptance Criteria

1. THE DNS_Selector_GUI SHALL 提供一个主窗口，包含以下区域：预设方案选择区、DNS 服务器列表区、测试域名列表区、测试参数配置区、测试控制区、测试结果展示区、当前 DNS 配置区
2. THE DNS_Selector_GUI SHALL 支持窗口大小调整，最小窗口尺寸为 800x600 像素
3. THE DNS_Selector_GUI SHALL 在标题栏显示应用名称"DNS Selector"和当前版本号
4. WHILE 应用执行耗时操作（DNS 测试、DNS 设置修改）, THE DNS_Selector_GUI SHALL 保持界面响应，操作在后台线程中执行
5. THE DNS_Selector_GUI SHALL 支持浅色和深色两种主题模式，并跟随 Windows 系统主题设置

### Requirement 13: JSON 配置文件支持

**User Story:** 作为用户，我希望能够通过 JSON 配置文件自定义 DNS 服务器、测试域名和测试参数，以便在不同场景下快速切换配置。

#### Acceptance Criteria

1. THE DNS_Selector_GUI SHALL 支持读取和写入 JSON 格式的 Config_File
2. THE Config_File SHALL 包含以下可配置项：自定义 DNS 服务器列表（含协议类型和可选的 TLS_Server_Name、Bootstrap_IP）、自定义测试域名列表、测试参数（queries、warmup、concurrency、timeout）
3. WHEN 用户点击"导入配置"按钮并选择有效的 JSON 文件, THE DNS_Selector_GUI SHALL 加载该文件中的配置并更新界面
4. WHEN 用户点击"导出配置"按钮, THE DNS_Selector_GUI SHALL 将当前的自定义服务器、域名和测试参数保存为 JSON 文件
5. IF 导入的 JSON 文件格式无效或包含不合法的配置项, THEN THE DNS_Selector_GUI SHALL 显示详细的错误提示信息，说明哪些配置项存在问题
6. WHEN 用户添加或删除自定义 DNS_Server 或 Test_Domain, THE DNS_Selector_GUI SHALL 自动将更新后的配置保存到默认 Config_File
7. WHEN 应用启动, THE DNS_Selector_GUI SHALL 从默认 Config_File 加载用户自定义配置（如文件存在）
8. IF 默认 Config_File 不存在或损坏, THEN THE DNS_Selector_GUI SHALL 使用默认预设配置启动并创建新的 Config_File

### Requirement 14: 测试结果持久化

**User Story:** 作为用户，我希望应用能够保存上次测试结果，以便下次打开时无需重新测试即可查看历史数据。

#### Acceptance Criteria

1. WHEN 测试完成, THE DNS_Selector_GUI SHALL 将测试结果（包含所有指标）保存到本地存储
2. WHEN 应用启动, THE DNS_Selector_GUI SHALL 从本地存储加载上次测试结果并显示，同时标注测试时间
3. IF 本地存储中无历史测试结果, THEN THE DNS_Selector_GUI SHALL 在测试结果区域显示"尚未进行测试"的提示信息

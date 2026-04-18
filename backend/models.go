package backend

// AddServerRequest 添加自定义 DNS 服务器的请求
type AddServerRequest struct {
	Protocol      string `json:"protocol"`      // "udp" | "dot" | "doh"
	Address       string `json:"address"`       // IP 地址、域名或 URL
	TLSServerName string `json:"tlsServerName"` // 可选，DoT/DoH 的 TLS 服务器名
	BootstrapIP   string `json:"bootstrapIP"`   // 可选，DoT/DoH 的引导 IP
}

// ServerInfo 服务器列表项（前端展示用）
type ServerInfo struct {
	Name             string `json:"name"`
	Address          string `json:"address"`
	Protocol         string `json:"protocol"` // "udp" | "dot" | "doh"
	TLSServerName    string `json:"tlsServerName"`
	IsPreset         bool   `json:"isPreset"`         // 是否为预设项（不可删除）
	CanApplyToSystem bool   `json:"canApplyToSystem"` // 是否可直接写入系统 DNS
}

// DomainInfo 域名列表项（前端展示用）
type DomainInfo struct {
	Domain   string `json:"domain"`
	IsPreset bool   `json:"isPreset"`
}

// TestParams 测试参数
type TestParams struct {
	Queries     int     `json:"queries"`     // 每域名查询次数，默认 10
	Warmup      int     `json:"warmup"`      // 预热查询次数，默认 1
	Concurrency int     `json:"concurrency"` // 最大并发数，默认 20
	Timeout     float64 `json:"timeout"`     // 超时时间（秒），默认 2.0
}

// TestResultItem 单个服务器的测试结果
type TestResultItem struct {
	Name             string  `json:"name"`
	Address          string  `json:"address"`
	Protocol         string  `json:"protocol"`
	MedianLatencyMs  float64 `json:"medianLatencyMs"`
	P95LatencyMs     float64 `json:"p95LatencyMs"`
	SuccessRate      float64 `json:"successRate"`
	RawSuccesses     int     `json:"rawSuccesses"`
	Successes        int     `json:"successes"`
	Total            int     `json:"total"`
	AnswerMismatches int     `json:"answerMismatches"`
	Score            float64 `json:"score"`
	IsTimeout        bool    `json:"isTimeout"`
	CanApplyToSystem bool    `json:"canApplyToSystem"` // 是否可直接写入系统 DNS
}

// TestResultsData 完整测试结果
type TestResultsData struct {
	Items    []TestResultItem `json:"items"`
	TestTime string           `json:"testTime"`
	Preset   string           `json:"preset"`
	BestDNS  string           `json:"bestDNS"`
}

// NetworkAdapterInfo 网络适配器信息
type NetworkAdapterInfo struct {
	Name         string   `json:"name"`
	InterfaceIdx int      `json:"interfaceIdx"`
	Status       string   `json:"status"`
	IPAddresses  []string `json:"ipAddresses"` // 适配器自身的 IPv4 地址列表
	CurrentDNS   []string `json:"currentDNS"`
}

// AppConfig JSON 配置文件的完整结构
type AppConfig struct {
	CurrentPreset string              `json:"currentPreset"`
	CustomServers []CustomServerEntry `json:"customServers"`
	CustomDomains []string            `json:"customDomains"`
	TestParams    TestParams          `json:"testParams"`
}

// CustomServerEntry 自定义服务器配置项
type CustomServerEntry struct {
	Protocol      string `json:"protocol"`
	Address       string `json:"address"`
	TLSServerName string `json:"tlsServerName,omitempty"`
	BootstrapIP   string `json:"bootstrapIP,omitempty"`
}

// PersistedResults 持久化的测试结果
type PersistedResults struct {
	Results TestResultsData `json:"results"`
	Version string          `json:"version"`
}

// ExportConfig 导出配置的完整结构，包含预设内容 + 自定义内容
type ExportConfig struct {
	CurrentPreset  string              `json:"currentPreset"`
	PresetServers  []CustomServerEntry `json:"presetServers"`  // 预设服务器（只读参考）
	PresetDomains  []string            `json:"presetDomains"`  // 预设域名（只读参考）
	CustomServers  []CustomServerEntry `json:"customServers"`  // 自定义服务器
	CustomDomains  []string            `json:"customDomains"`  // 自定义域名
	TestParams     TestParams          `json:"testParams"`
}

// DefaultTestParams 返回默认测试参数
func DefaultTestParams() TestParams {
	return TestParams{
		Queries:     10,
		Warmup:      1,
		Concurrency: 20,
		Timeout:     2.0,
	}
}

// DefaultAppConfig 返回默认应用配置
func DefaultAppConfig() AppConfig {
	return AppConfig{
		CurrentPreset: "cn",
		CustomServers: []CustomServerEntry{},
		CustomDomains: []string{},
		TestParams:    DefaultTestParams(),
	}
}

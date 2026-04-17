package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/betterlmy/dns-selector/selector"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppService 是 Wails 绑定的主服务，负责协调所有后端子服务。
type AppService struct {
	ctx             context.Context
	benchmark       *BenchmarkService
	config          *ConfigService
	dnsConfig       *DNSConfigService
	preset          *PresetService
	platform        PlatformProvider
	testParams      TestParams
	lastResults     *TestResultsData
	cancelBenchmark context.CancelFunc
	mu              sync.RWMutex
}

// NewAppService 创建 AppService 实例，初始化所有子服务。
func NewAppService() *AppService {
	platform := NewPlatformProvider()
	return &AppService{
		benchmark:  NewBenchmarkService(),
		config:     NewConfigService(),
		dnsConfig:  NewDNSConfigService(platform),
		preset:     NewPresetService(),
		platform:   platform,
		testParams: DefaultTestParams(),
	}
}

// OnStartup 由 Wails 框架在应用启动时调用。
// 初始化所有子服务、加载配置文件、加载历史测试结果。
func (a *AppService) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// 从默认路径加载配置
	cfg, _ := a.config.Load(a.config.GetDefaultPath())
	if cfg != nil {
		a.applyConfig(cfg)
	}

	// 加载历史测试结果
	results, _ := a.config.LoadResults()
	if results != nil {
		a.mu.Lock()
		a.lastResults = results
		a.mu.Unlock()
	}
}

// applyConfig 将加载的 AppConfig 应用到预设服务和测试参数。
func (a *AppService) applyConfig(cfg *AppConfig) {
	// 切换预设方案
	_ = a.preset.SwitchPreset(cfg.CurrentPreset)

	// 加载自定义服务器
	customServers := make([]selector.DNSServer, 0, len(cfg.CustomServers))
	for _, cs := range cfg.CustomServers {
		server := selector.DNSServer{
			Name:          cs.Address,
			Address:       cs.Address,
			Protocol:      cs.Protocol,
			TLSServerName: cs.TLSServerName,
		}
		if cs.BootstrapIP != "" {
			server.BootstrapIPs = []string{cs.BootstrapIP}
		}
		customServers = append(customServers, server)
	}
	a.preset.SetCustomServers(customServers)

	// 加载自定义域名
	a.preset.SetCustomDomains(cfg.CustomDomains)

	// 加载测试参数
	a.mu.Lock()
	a.testParams = cfg.TestParams
	a.mu.Unlock()
}

// --- 预设管理 ---

// GetCurrentPreset 返回当前激活的预设方案名称（"cn" 或 "global"）。
func (a *AppService) GetCurrentPreset() string {
	return a.preset.GetCurrentPreset()
}

// SwitchPreset 切换到指定预设方案并自动保存配置。
func (a *AppService) SwitchPreset(name string) error {
	if err := a.preset.SwitchPreset(name); err != nil {
		return err
	}
	return a.autoSaveConfig()
}

// GetServerList 返回合并后的服务器列表（预设 + 自定义），转换为 ServerInfo 格式。
func (a *AppService) GetServerList() []ServerInfo {
	servers := a.preset.GetMergedServers()
	result := make([]ServerInfo, 0, len(servers))
	for _, s := range servers {
		result = append(result, ServerInfo{
			Name:          s.Name,
			Address:       s.Address,
			Protocol:      s.Protocol,
			TLSServerName: s.TLSServerName,
			IsPreset:      a.preset.IsPresetItem(s.Address),
		})
	}
	return result
}

// GetDomainList 返回合并后的域名列表（预设 + 自定义），转换为 DomainInfo 格式。
func (a *AppService) GetDomainList() []DomainInfo {
	domains := a.preset.GetMergedDomains()
	result := make([]DomainInfo, 0, len(domains))
	for _, d := range domains {
		result = append(result, DomainInfo{
			Domain:   d,
			IsPreset: a.preset.IsPresetItem(d),
		})
	}
	return result
}

// --- 自定义服务器/域名管理 ---

// AddCustomServer 验证并添加自定义 DNS 服务器，然后自动保存配置。
func (a *AppService) AddCustomServer(req AddServerRequest) error {
	if err := ValidateServerAddress(req.Protocol, req.Address); err != nil {
		return err
	}

	server := selector.DNSServer{
		Name:          req.Address,
		Address:       req.Address,
		Protocol:      req.Protocol,
		TLSServerName: req.TLSServerName,
	}
	if req.BootstrapIP != "" {
		server.BootstrapIPs = []string{req.BootstrapIP}
	}

	if err := a.preset.AddCustomServer(server); err != nil {
		return err
	}
	return a.autoSaveConfig()
}

// RemoveCustomServer 按地址删除自定义服务器并自动保存配置。
func (a *AppService) RemoveCustomServer(address string) error {
	if err := a.preset.RemoveCustomServer(address); err != nil {
		return err
	}
	return a.autoSaveConfig()
}

// AddCustomDomain 验证并添加自定义域名，然后自动保存配置。
func (a *AppService) AddCustomDomain(domain string) error {
	if err := ValidateDomain(domain); err != nil {
		return err
	}
	if err := a.preset.AddCustomDomain(domain); err != nil {
		return err
	}
	return a.autoSaveConfig()
}

// RemoveCustomDomain 删除自定义域名并自动保存配置。
func (a *AppService) RemoveCustomDomain(domain string) error {
	if err := a.preset.RemoveCustomDomain(domain); err != nil {
		return err
	}
	return a.autoSaveConfig()
}

// --- 测试参数 ---

// GetTestParams 返回当前测试参数。
func (a *AppService) GetTestParams() TestParams {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.testParams
}

// SetTestParams 验证并设置测试参数，然后自动保存配置。
func (a *AppService) SetTestParams(params TestParams) error {
	if err := ValidateTestParams(params); err != nil {
		return err
	}
	a.mu.Lock()
	a.testParams = params
	a.mu.Unlock()
	return a.autoSaveConfig()
}

// --- 测试执行 ---

// StartBenchmark 在 goroutine 中启动 DNS 测试，通过 Wails 事件推送进度。
func (a *AppService) StartBenchmark() error {
	if a.benchmark.IsRunning() {
		return fmt.Errorf("测试正在运行中")
	}

	servers := a.preset.GetMergedServers()
	domains := a.preset.GetMergedDomains()

	a.mu.RLock()
	params := a.testParams
	a.mu.RUnlock()

	if err := a.benchmark.BuildSelector(servers, domains, params); err != nil {
		return err
	}

	// 总查询数 = 服务器数 × 域名数 × 每域名查询次数
	totalQueries := len(servers) * len(domains) * params.Queries
	var completed int64

	ctx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	a.cancelBenchmark = cancel
	a.mu.Unlock()

	go func() {
		// 进度回调：每完成一次查询时调用
		progressCb := func() {
			current := atomic.AddInt64(&completed, 1)
			percent := float64(current) / float64(totalQueries) * 100
			if percent > 100 {
				percent = 100
			}
			wailsRuntime.EventsEmit(a.ctx, "benchmark:progress", map[string]interface{}{
				"completed": current,
				"total":     totalQueries,
				"percent":   percent,
			})
		}

		results, err := a.benchmark.RunBenchmark(ctx, progressCb)
		if err != nil {
			if ctx.Err() == context.Canceled {
				// 用户主动停止测试
				wailsRuntime.EventsEmit(a.ctx, "benchmark:stopped", map[string]interface{}{})
				return
			}
			// 测试出错
			wailsRuntime.EventsEmit(a.ctx, "benchmark:error", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}

		// 处理测试结果
		preset := a.preset.GetCurrentPreset()
		resultsData := ProcessResults(results, preset)

		// 保存测试结果
		a.mu.Lock()
		a.lastResults = resultsData
		a.mu.Unlock()
		_ = a.config.SaveResults(resultsData)

		// 推送测试完成事件
		wailsRuntime.EventsEmit(a.ctx, "benchmark:complete", resultsData)
	}()

	return nil
}

// StopBenchmark 取消正在运行的测试。
func (a *AppService) StopBenchmark() error {
	a.mu.RLock()
	cancel := a.cancelBenchmark
	a.mu.RUnlock()

	if cancel == nil {
		return fmt.Errorf("当前没有正在运行的测试")
	}
	cancel()
	return nil
}

// --- 测试结果 ---

// GetLastResults 返回最近一次测试结果，如果没有则返回 nil。
func (a *AppService) GetLastResults() (*TestResultsData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastResults, nil
}

// --- 系统 DNS ---

// GetNetworkAdapters 返回所有活动网络适配器及其 DNS 配置。
func (a *AppService) GetNetworkAdapters() ([]NetworkAdapterInfo, error) {
	return a.dnsConfig.GetAdapters()
}

// ApplyDNS 将指定 DNS 服务器设置到指定网络适配器，支持首选和备用 DNS。
// primaryDNS 为首选，secondaryDNS 为备用（可为空）。
func (a *AppService) ApplyDNS(adapterName string, primaryDNS string, secondaryDNS string) error {
	return a.dnsConfig.SetDNS(adapterName, primaryDNS, secondaryDNS)
}

// RestoreDHCP 将指定网络适配器的 DNS 恢复为自动获取（DHCP）。
func (a *AppService) RestoreDHCP(adapterName string) error {
	return a.dnsConfig.ResetToAuto(adapterName)
}

// IsAdmin 返回当前进程是否具有管理员权限。
func (a *AppService) IsAdmin() bool {
	return a.platform.CheckAdmin()
}

// --- 配置导入导出 ---

// ImportConfig 读取 JSON 配置文件，验证后应用到当前状态。
func (a *AppService) ImportConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("JSON 格式无效: %w", err)
	}

	if err := ValidateConfig(&cfg); err != nil {
		return err
	}

	a.applyConfig(&cfg)
	return a.autoSaveConfig()
}

// ExportConfig 将当前配置（含预设内容）导出为 JSON 文件。
func (a *AppService) ExportConfig(filePath string) error {
	cfg := a.buildExportConfig()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// --- 主题 ---

// GetSystemTheme 检测系统主题设置。
// 返回 "dark"（深色模式）或 "light"（浅色模式）。
func (a *AppService) GetSystemTheme() string {
	return a.platform.GetSystemTheme()
}

// --- 内部辅助方法 ---

// autoSaveConfig 构建当前配置并保存到默认路径。
func (a *AppService) autoSaveConfig() error {
	cfg := a.buildCurrentConfig()
	return a.config.UpdateConfig(&cfg)
}

// buildCurrentConfig 从当前状态构建 AppConfig 对象（仅自定义内容，用于自动保存）。
func (a *AppService) buildCurrentConfig() AppConfig {
	customServers := a.preset.GetCustomServers()
	entries := make([]CustomServerEntry, 0, len(customServers))
	for _, s := range customServers {
		entry := CustomServerEntry{
			Protocol:      s.Protocol,
			Address:       s.Address,
			TLSServerName: s.TLSServerName,
		}
		if len(s.BootstrapIPs) > 0 {
			entry.BootstrapIP = s.BootstrapIPs[0]
		}
		entries = append(entries, entry)
	}

	a.mu.RLock()
	params := a.testParams
	a.mu.RUnlock()

	return AppConfig{
		CurrentPreset: a.preset.GetCurrentPreset(),
		CustomServers: entries,
		CustomDomains: a.preset.GetCustomDomains(),
		TestParams:    params,
	}
}

// buildExportConfig 构建包含预设内容的完整导出配置。
func (a *AppService) buildExportConfig() ExportConfig {
	presetName := a.preset.GetCurrentPreset()
	preset, _ := GetPreset(presetName)

	// 预设服务器转换为 CustomServerEntry 格式
	presetServers := make([]CustomServerEntry, 0, len(preset.Servers))
	for _, s := range preset.Servers {
		entry := CustomServerEntry{
			Protocol:      s.Protocol,
			Address:       s.Address,
			TLSServerName: s.TLSServerName,
		}
		if len(s.BootstrapIPs) > 0 {
			entry.BootstrapIP = s.BootstrapIPs[0]
		}
		presetServers = append(presetServers, entry)
	}

	// 自定义服务器
	customServers := a.preset.GetCustomServers()
	entries := make([]CustomServerEntry, 0, len(customServers))
	for _, s := range customServers {
		entry := CustomServerEntry{
			Protocol:      s.Protocol,
			Address:       s.Address,
			TLSServerName: s.TLSServerName,
		}
		if len(s.BootstrapIPs) > 0 {
			entry.BootstrapIP = s.BootstrapIPs[0]
		}
		entries = append(entries, entry)
	}

	a.mu.RLock()
	params := a.testParams
	a.mu.RUnlock()

	return ExportConfig{
		CurrentPreset: presetName,
		PresetServers: presetServers,
		PresetDomains: preset.Domains,
		CustomServers: entries,
		CustomDomains: a.preset.GetCustomDomains(),
		TestParams:    params,
	}
}

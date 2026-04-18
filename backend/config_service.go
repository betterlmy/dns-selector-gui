package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConfigService 管理 JSON 配置文件的读写和自动保存。
type ConfigService struct {
	configPath  string     // 配置文件路径
	resultsPath string     // 测试结果文件路径
	config      *AppConfig // 当前内存中的配置
	mu          sync.RWMutex
}

// NewConfigService 创建 ConfigService 实例，使用默认配置和默认路径。
func NewConfigService() *ConfigService {
	return NewConfigServiceWithPaths(defaultConfigPath(), defaultResultsPath())
}

// NewConfigServiceWithPaths 创建 ConfigService 实例，允许测试或调用方注入配置/结果文件路径。
func NewConfigServiceWithPaths(configPath, resultsPath string) *ConfigService {
	defaultCfg := DefaultAppConfig()
	if configPath == "" {
		configPath = defaultConfigPath()
	}
	if resultsPath == "" {
		resultsPath = defaultResultsPath()
	}
	return &ConfigService{
		configPath:  configPath,
		resultsPath: resultsPath,
		config:      &defaultCfg,
	}
}

// defaultConfigPath 返回默认配置文件路径：{configDir}/dns-selector-gui/config.json
// 使用 os.UserConfigDir() 获取跨平台配置目录，失败时回退到当前目录。
func defaultConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "dns-selector-gui", "config.json")
}

// defaultResultsPath 返回默认测试结果文件路径：{configDir}/dns-selector-gui/last_results.json
// 使用 os.UserConfigDir() 获取跨平台配置目录，失败时回退到当前目录。
func defaultResultsPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "dns-selector-gui", "last_results.json")
}

// GetDefaultPath 返回默认配置文件路径。
func (c *ConfigService) GetDefaultPath() string {
	return c.configPath
}

// Load 从指定路径读取 JSON 配置文件并存储到内存。
//
// 错误处理策略：
//   - 文件不存在：返回默认配置，error 为 nil（首次启动的正常情况）
//   - 文件存在但读取失败（权限不足等）：返回 nil + error，不覆盖原文件
//   - 文件存在但 JSON 损坏：返回 nil + error，让调用方决定是否重置
func (c *ConfigService) Load(path string) (*AppConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.configPath = path

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在属于正常情况（首次运行），返回默认配置
			defaultCfg := DefaultAppConfig()
			c.config = &defaultCfg
			return c.config, nil
		}
		// 文件存在但无法读取（权限不足等），不静默处理
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// JSON 损坏，不自动覆盖，让调用方决策
		return nil, fmt.Errorf("配置文件格式损坏，请检查或删除 %s: %w", path, err)
	}

	c.config = &cfg
	return c.config, nil
}

// Save 将当前配置序列化为格式化 JSON 并写入指定路径。
// 如果父目录不存在会自动创建。
func (c *ConfigService) Save(path string) error {
	c.mu.RLock()
	cfg := c.config
	c.mu.RUnlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录 %s 失败: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件 %s 失败: %w", path, err)
	}

	return nil
}

// GetConfig 返回当前内存中的配置（读锁保护）。
func (c *ConfigService) GetConfig() *AppConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// UpdateConfig 更新内存中的配置并自动保存到 configPath。
func (c *ConfigService) UpdateConfig(config *AppConfig) error {
	c.mu.Lock()
	c.config = config
	path := c.configPath
	c.mu.Unlock()

	return c.Save(path)
}

// SaveResults 将测试结果保存到 {configDir}/dns-selector-gui/last_results.json。
func (c *ConfigService) SaveResults(results *TestResultsData) error {
	if results == nil {
		return fmt.Errorf("测试结果不能为空")
	}

	persisted := PersistedResults{
		Results: *results,
		Version: "1.0.0",
	}

	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化测试结果失败: %w", err)
	}

	path := c.resultsPath
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建结果目录 %s 失败: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入结果文件 %s 失败: %w", path, err)
	}

	return nil
}

// LoadResults 从 {configDir}/dns-selector-gui/last_results.json 加载测试结果。
// 如果文件不存在，返回 nil（不报错）。
func (c *ConfigService) LoadResults() (*TestResultsData, error) {
	path := c.resultsPath

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取结果文件 %s 失败: %w", path, err)
	}

	var persisted PersistedResults
	if err := json.Unmarshal(data, &persisted); err != nil {
		return nil, fmt.Errorf("解析结果文件 %s 失败: %w", path, err)
	}

	return &persisted.Results, nil
}

// ValidateConfig 验证 AppConfig 的合法性，检查测试参数、自定义服务器地址和自定义域名。
// 用于配置导入时的验证。
func ValidateConfig(config *AppConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证测试参数
	if err := ValidateTestParams(config.TestParams); err != nil {
		return fmt.Errorf("测试参数无效: %w", err)
	}

	// 验证自定义服务器地址
	for i, server := range config.CustomServers {
		if err := ValidateServerEntry(AddServerRequest{
			Protocol:      server.Protocol,
			Address:       server.Address,
			TLSServerName: server.TLSServerName,
			BootstrapIP:   server.BootstrapIP,
		}); err != nil {
			return fmt.Errorf("第 %d 个自定义服务器无效: %w", i+1, err)
		}
	}

	// 验证自定义域名
	for i, domain := range config.CustomDomains {
		if err := ValidateDomain(domain); err != nil {
			return fmt.Errorf("第 %d 个自定义域名无效: %w", i+1, err)
		}
	}

	// 验证预设名称
	preset := config.CurrentPreset
	if preset != "cn" && preset != "global" {
		return fmt.Errorf("无效的预设名称: %q，必须为 \"cn\" 或 \"global\"", preset)
	}

	return nil
}

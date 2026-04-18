package backend

import (
	"fmt"

	"github.com/betterlmy/dns-selector/selector"
)

// PresetService 管理预设方案和自定义服务器/域名列表。
type PresetService struct {
	currentPreset string // "cn" 或 "global"
	customServers []selector.DNSServer
	customDomains []string
}

// NewPresetService 创建 PresetService 实例，默认使用 "cn" 预设，自定义列表为空。
func NewPresetService() *PresetService {
	return &PresetService{
		currentPreset: "cn",
		customServers: []selector.DNSServer{},
		customDomains: []string{},
	}
}

// GetCurrentPreset 返回当前激活的预设方案名称（"cn" 或 "global"）。
func (p *PresetService) GetCurrentPreset() string {
	return p.currentPreset
}

// SwitchPreset 切换到指定预设方案（"cn" 或 "global"）。
// 自定义服务器和域名不受影响，始终保留。
func (p *PresetService) SwitchPreset(name string) error {
	if name != "cn" && name != "global" {
		return fmt.Errorf("未知的预设方案: %q，有效值为 \"cn\" 和 \"global\"", name)
	}
	p.currentPreset = name
	return nil
}

// GetMergedServers 返回预设服务器 + 自定义服务器的合并列表。
func (p *PresetService) GetMergedServers() []selector.DNSServer {
	preset, _ := GetPreset(p.currentPreset)
	merged := make([]selector.DNSServer, 0, len(preset.Servers)+len(p.customServers))
	merged = append(merged, preset.Servers...)
	merged = append(merged, p.customServers...)
	return merged
}

// GetMergedDomains 返回预设域名 + 自定义域名的合并列表。
func (p *PresetService) GetMergedDomains() []string {
	preset, _ := GetPreset(p.currentPreset)
	merged := make([]string, 0, len(preset.Domains)+len(p.customDomains))
	merged = append(merged, preset.Domains...)
	merged = append(merged, p.customDomains...)
	return merged
}

// serverKey 返回服务器的复合唯一标识：protocol|address|tlsServerName。
func serverKey(protocol, address, tlsServerName string) string {
	return protocol + "|" + address + "|" + tlsServerName
}

// IsPresetServer 判断给定服务器是否属于当前预设（属于预设的项不可删除）。
func (p *PresetService) IsPresetServer(protocol, address, tlsServerName string) bool {
	preset, _ := GetPreset(p.currentPreset)
	key := serverKey(protocol, address, tlsServerName)
	for _, s := range preset.Servers {
		if serverKey(s.Protocol, s.Address, s.TLSServerName) == key {
			return true
		}
	}
	return false
}

// IsPresetDomain 判断给定域名是否属于当前预设。
func (p *PresetService) IsPresetDomain(domain string) bool {
	preset, _ := GetPreset(p.currentPreset)
	for _, d := range preset.Domains {
		if d == domain {
			return true
		}
	}
	return false
}

// IsPresetItem 判断给定的地址或域名是否属于当前预设（向后兼容，仅用于域名判断）。
func (p *PresetService) IsPresetItem(address string) bool {
	return p.IsPresetDomain(address)
}

// AddCustomServer 向自定义服务器列表追加一个 DNS 服务器。
func (p *PresetService) AddCustomServer(server selector.DNSServer) error {
	p.customServers = append(p.customServers, server)
	return nil
}

// RemoveCustomServer 按复合标识（protocol+address+tlsServerName）删除一个自定义服务器。
// 如果该服务器属于预设，返回错误。
func (p *PresetService) RemoveCustomServer(protocol, address, tlsServerName string) error {
	if p.IsPresetServer(protocol, address, tlsServerName) {
		return fmt.Errorf("无法删除预设服务器")
	}
	key := serverKey(protocol, address, tlsServerName)
	for i, s := range p.customServers {
		if serverKey(s.Protocol, s.Address, s.TLSServerName) == key {
			p.customServers = append(p.customServers[:i], p.customServers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("未找到自定义服务器 %q", address)
}

// AddCustomDomain 向自定义域名列表追加一个域名。
func (p *PresetService) AddCustomDomain(domain string) error {
	p.customDomains = append(p.customDomains, domain)
	return nil
}

// RemoveCustomDomain 删除一个自定义域名。
// 如果该域名属于预设，返回错误。
func (p *PresetService) RemoveCustomDomain(domain string) error {
	if p.IsPresetItem(domain) {
		return fmt.Errorf("无法删除预设域名")
	}
	for i, d := range p.customDomains {
		if d == domain {
			p.customDomains = append(p.customDomains[:i], p.customDomains[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("未找到自定义域名 %q", domain)
}

// SetCustomServers 替换自定义服务器列表（加载配置时使用）。
func (p *PresetService) SetCustomServers(servers []selector.DNSServer) {
	p.customServers = servers
}

// SetCustomDomains 替换自定义域名列表（加载配置时使用）。
func (p *PresetService) SetCustomDomains(domains []string) {
	p.customDomains = domains
}

// GetCustomServers 返回当前自定义服务器列表。
func (p *PresetService) GetCustomServers() []selector.DNSServer {
	return p.customServers
}

// GetCustomDomains 返回当前自定义域名列表。
func (p *PresetService) GetCustomDomains() []string {
	return p.customDomains
}

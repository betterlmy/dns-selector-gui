package backend

import (
	"fmt"
	"net"
	"strings"
)

// DNSConfigService 通过 PlatformProvider 接口读取、修改系统 DNS 配置。
// 不包含任何平台特定代码，所有平台操作委托给 PlatformProvider 实现。
type DNSConfigService struct {
	platform PlatformProvider
}

// NewDNSConfigService 创建 DNSConfigService 实例。
func NewDNSConfigService(platform PlatformProvider) *DNSConfigService {
	return &DNSConfigService{platform: platform}
}

// GetAdapters 返回所有活动网络适配器及其 DNS 和 IP 配置。
func (d *DNSConfigService) GetAdapters() ([]NetworkAdapterInfo, error) {
	return d.platform.GetAdapters()
}

// SetDNS 将指定 DNS 服务器设置到指定网络适配器。
func (d *DNSConfigService) SetDNS(adapterName string, primaryDNS string, secondaryDNS string) error {
	return d.platform.SetDNS(adapterName, primaryDNS, secondaryDNS)
}

// ResetToAuto 将指定网络适配器的 DNS 恢复为自动获取（DHCP）。
func (d *DNSConfigService) ResetToAuto(adapterName string) error {
	return d.platform.ResetToAuto(adapterName)
}

// CheckAdmin 检查当前进程是否具有修改系统 DNS 配置的权限。
func (d *DNSConfigService) CheckAdmin() bool {
	return d.platform.CheckAdmin()
}

// extractDNSIP 从 DNS 地址中提取可用于系统设置的 IP 地址。
// - 普通 IPv4（8.8.8.8）→ 直接返回
// - DoH URL（https://223.6.6.6/dns-query）→ 提取 host 部分
// - DoH URL（https://dns.google/dns-query）→ 返回域名（Windows 10 2004+ 支持）
// - DoT（dns.google 或 8.8.8.8@dns.google）→ 提取 IP 或域名
func extractDNSIP(address string) (string, error) {
	if address == "" {
		return "", fmt.Errorf("地址为空")
	}

	// DoH URL：https://... 格式
	if strings.HasPrefix(strings.ToLower(address), "https://") {
		// 去掉 @TLSServerName 后缀
		addr := address
		if idx := strings.LastIndex(addr, "@"); idx != -1 {
			addr = addr[:idx]
		}
		// 提取 host
		addr = strings.TrimPrefix(strings.ToLower(addr), "https://")
		if idx := strings.Index(addr, "/"); idx != -1 {
			addr = addr[:idx]
		}
		// 去掉端口
		if idx := strings.LastIndex(addr, ":"); idx != -1 {
			addr = addr[:idx]
		}
		if addr == "" {
			return "", fmt.Errorf("无法从 DoH URL 提取主机地址")
		}
		return addr, nil
	}

	// DoT IP@TLSServerName 格式：取 @ 前的 IP
	if strings.Contains(address, "@") {
		parts := strings.SplitN(address, "@", 2)
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
		return ip, nil
	}

	// 普通 IPv4 或域名，直接返回
	return address, nil
}

package backend

import (
	"fmt"
	"net"
	"strings"

	"github.com/betterlmy/dns-selector/selector"
)

// ResolveSystemDNSAddresses 将测速用的 DNSServer 转换为系统可接受的 IPv4 地址列表。
//
// 转换规则：
//   - UDP：直接返回 Address
//   - DoT IP 型：直接返回 Address
//   - DoT 域名型 + BootstrapIPs：返回 BootstrapIPs
//   - DoT 域名型无 BootstrapIPs：返回错误
//   - DoH URL host 是 IP：返回该 IP
//   - DoH URL host 是域名 + BootstrapIPs：返回 BootstrapIPs
//   - DoH URL host 是域名无 BootstrapIPs：返回错误
func ResolveSystemDNSAddresses(server selector.DNSServer) ([]string, error) {
	switch strings.ToLower(server.Protocol) {
	case "udp":
		if server.Address == "" {
			return nil, fmt.Errorf("服务器地址为空")
		}
		return []string{server.Address}, nil

	case "dot":
		return resolveDotAddress(server)

	case "doh":
		return resolveDoHAddress(server)

	default:
		return nil, fmt.Errorf("未知协议: %q", server.Protocol)
	}
}

func resolveDotAddress(server selector.DNSServer) ([]string, error) {
	addr := server.Address
	// IP@TLSServerName 格式：取 @ 前的部分
	if idx := strings.Index(addr, "@"); idx != -1 {
		addr = addr[:idx]
	}
	addr = strings.TrimSpace(addr)

	if net.ParseIP(addr) != nil {
		return []string{addr}, nil
	}

	// 域名型，优先使用 BootstrapIPs
	if len(server.BootstrapIPs) > 0 {
		return validateIPs(server.BootstrapIPs)
	}
	return nil, fmt.Errorf("DoT 服务器 %q 使用域名地址但未配置 BootstrapIP，无法写入系统 DNS", server.Address)
}

func resolveDoHAddress(server selector.DNSServer) ([]string, error) {
	addr := server.Address
	// 去掉 @TLSServerName 后缀
	if idx := strings.LastIndex(addr, "@"); idx != -1 {
		addr = addr[:idx]
	}

	host := extractURLHost(addr)
	if host == "" {
		return nil, fmt.Errorf("无法从 DoH URL %q 解析主机地址", server.Address)
	}

	if net.ParseIP(host) != nil {
		return []string{host}, nil
	}

	// 域名型，优先使用 BootstrapIPs
	if len(server.BootstrapIPs) > 0 {
		return validateIPs(server.BootstrapIPs)
	}
	return nil, fmt.Errorf("DoH 服务器 %q 使用域名地址但未配置 BootstrapIP，无法写入系统 DNS", server.Address)
}

// extractURLHost 从 https://host/path 中提取 host（去掉端口）。
func extractURLHost(rawURL string) string {
	s := strings.ToLower(rawURL)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}
	// 去掉端口
	if host, _, err := net.SplitHostPort(s); err == nil {
		return host
	}
	return s
}

// validateIPs 检查每个地址是否为合法 IP，返回合法 IP 列表。
func validateIPs(ips []string) ([]string, error) {
	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if net.ParseIP(ip) == nil {
			return nil, fmt.Errorf("BootstrapIP %q 不是合法 IP 地址", ip)
		}
		result = append(result, ip)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("BootstrapIPs 列表为空")
	}
	return result, nil
}

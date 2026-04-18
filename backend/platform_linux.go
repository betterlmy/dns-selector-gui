//go:build linux

package backend

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

// linuxPlatform 实现 PlatformProvider，优先使用 systemd-resolved。
type linuxPlatform struct{}

// NewPlatformProvider 创建 Linux 平台的 PlatformProvider 实例。
func NewPlatformProvider() PlatformProvider {
	return &linuxPlatform{}
}

// GetAdapters 使用 net.Interfaces() 获取所有活动网络适配器及其 DNS 和 IP 配置。
func (l *linuxPlatform) GetAdapters() ([]NetworkAdapterInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口失败: %w", err)
	}

	var results []NetworkAdapterInfo
	for _, iface := range ifaces {
		// Filter out loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get IP addresses
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		var ipAddresses []string
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				ipAddresses = append(ipAddresses, ip.String())
			}
		}

		// Only include interfaces with at least one IPv4 address
		if len(ipAddresses) == 0 {
			continue
		}

		// Get DNS servers for this interface
		dnsServers := l.getInterfaceDNS(iface.Name)

		results = append(results, NetworkAdapterInfo{
			Name:         iface.Name,
			InterfaceIdx: iface.Index,
			Status:       "Up",
			IPAddresses:  ipAddresses,
			CurrentDNS:   dnsServers,
		})
	}

	if results == nil {
		results = []NetworkAdapterInfo{}
	}
	return results, nil
}

// SetDNS 将指定 DNS 服务器设置到指定网络接口。
// 优先使用 resolvectl，回退到写入 /etc/resolv.conf。
func (l *linuxPlatform) SetDNS(adapterName string, primaryDNS string, secondaryDNS string) error {
	if adapterName == "" {
		return fmt.Errorf("适配器名称不能为空")
	}
	if primaryDNS == "" {
		return fmt.Errorf("首选 DNS 不能为空")
	}

	if hasResolvectl() {
		args := []string{"dns", adapterName, primaryDNS}
		if secondaryDNS != "" {
			args = append(args, secondaryDNS)
		}
		out, err := exec.Command("resolvectl", args...).CombinedOutput()
		if err != nil {
			if os.Getuid() != 0 {
				return fmt.Errorf("设置 DNS 失败：需要 root 权限，请使用 sudo 运行应用")
			}
			return fmt.Errorf("resolvectl 设置 DNS 失败: %s: %w", string(out), err)
		}
		return nil
	}

	// Fallback: write /etc/resolv.conf
	return writeResolvConf(primaryDNS, secondaryDNS)
}

// ResetToAuto 将指定网络接口的 DNS 恢复为自动获取（DHCP）。
func (l *linuxPlatform) ResetToAuto(adapterName string) error {
	if adapterName == "" {
		return fmt.Errorf("适配器名称不能为空")
	}

	if hasResolvectl() {
		out, err := exec.Command("resolvectl", "revert", adapterName).CombinedOutput()
		if err != nil {
			if os.Getuid() != 0 {
				return fmt.Errorf("重置 DNS 失败：需要 root 权限，请使用 sudo 运行应用")
			}
			return fmt.Errorf("resolvectl revert 失败: %s: %w", string(out), err)
		}
		return nil
	}

	// Fallback: restore /etc/resolv.conf from backup or write empty
	return restoreResolvConf()
}

// CheckAdmin 检查当前进程是否以 root 权限运行。
func (l *linuxPlatform) CheckAdmin() bool {
	return os.Getuid() == 0
}

// GetSystemTheme 通过 gsettings 检测 Linux 系统主题。
// 返回 "dark"（深色模式）或 "light"（浅色模式）。
func (l *linuxPlatform) GetSystemTheme() string {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").CombinedOutput()
	if err != nil {
		return "light"
	}
	result := strings.TrimSpace(strings.ToLower(string(out)))
	if strings.Contains(result, "prefer-dark") {
		return "dark"
	}
	return "light"
}

// --- 内部辅助函数 ---

// hasResolvectl 检测 systemd-resolved 是否可用。
func hasResolvectl() bool {
	_, err := exec.LookPath("resolvectl")
	return err == nil
}

// getInterfaceDNS 获取指定接口的 DNS 服务器列表。
// 优先使用 resolvectl，回退到解析 /etc/resolv.conf。
func (l *linuxPlatform) getInterfaceDNS(ifaceName string) []string {
	if hasResolvectl() {
		return parseDNSFromResolvectl(ifaceName)
	}
	return parseDNSFromResolvConf()
}

// parseDNSFromResolvectl 解析 resolvectl status <iface> 输出中的 DNS 服务器。
func parseDNSFromResolvectl(ifaceName string) []string {
	out, err := exec.Command("resolvectl", "status", ifaceName).CombinedOutput()
	if err != nil {
		return []string{}
	}

	var servers []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DNS Servers:") || strings.HasPrefix(line, "Current DNS Server:") {
			// Extract the value after the colon
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if value != "" && net.ParseIP(value) != nil {
					servers = append(servers, value)
				}
			}
		}
	}
	if servers == nil {
		servers = []string{}
	}
	return servers
}

// parseDNSFromResolvConf 解析 /etc/resolv.conf 中的 nameserver 行。
func parseDNSFromResolvConf() []string {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return []string{}
	}
	defer f.Close()

	var servers []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ip := fields[1]
				if net.ParseIP(ip) != nil {
					servers = append(servers, ip)
				}
			}
		}
	}
	if servers == nil {
		servers = []string{}
	}
	return servers
}

// writeResolvConf 备份并写入新的 /etc/resolv.conf。
func writeResolvConf(primaryIP string, secondaryIP string) error {
	// Backup existing resolv.conf
	if _, err := os.Stat("/etc/resolv.conf"); err == nil {
		data, err := os.ReadFile("/etc/resolv.conf")
		if err == nil {
			_ = os.WriteFile("/etc/resolv.conf.bak", data, 0644)
		}
	}

	var content strings.Builder
	content.WriteString("# Generated by DNS Selector GUI\n")
	content.WriteString(fmt.Sprintf("nameserver %s\n", primaryIP))
	if secondaryIP != "" {
		content.WriteString(fmt.Sprintf("nameserver %s\n", secondaryIP))
	}

	if err := os.WriteFile("/etc/resolv.conf", []byte(content.String()), 0644); err != nil {
		if os.Getuid() != 0 {
			return fmt.Errorf("写入 /etc/resolv.conf 失败：需要 root 权限，请使用 sudo 运行应用")
		}
		return fmt.Errorf("写入 /etc/resolv.conf 失败: %w", err)
	}
	return nil
}

// restoreResolvConf 从备份恢复 /etc/resolv.conf，或写入空配置。
func restoreResolvConf() error {
	bakPath := "/etc/resolv.conf.bak"
	if _, err := os.Stat(bakPath); err == nil {
		data, err := os.ReadFile(bakPath)
		if err != nil {
			return fmt.Errorf("读取 /etc/resolv.conf.bak 失败: %w", err)
		}
		if err := os.WriteFile("/etc/resolv.conf", data, 0644); err != nil {
			if os.Getuid() != 0 {
				return fmt.Errorf("恢复 /etc/resolv.conf 失败：需要 root 权限，请使用 sudo 运行应用")
			}
			return fmt.Errorf("恢复 /etc/resolv.conf 失败: %w", err)
		}
		return nil
	}

	// No backup exists, write empty resolv.conf
	content := "# Generated by DNS Selector GUI\n"
	if err := os.WriteFile("/etc/resolv.conf", []byte(content), 0644); err != nil {
		if os.Getuid() != 0 {
			return fmt.Errorf("写入 /etc/resolv.conf 失败：需要 root 权限，请使用 sudo 运行应用")
		}
		return fmt.Errorf("写入 /etc/resolv.conf 失败: %w", err)
	}
	return nil
}

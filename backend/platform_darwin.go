//go:build darwin

package backend

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

// darwinPlatform 实现 PlatformProvider，使用 macOS 原生命令。
type darwinPlatform struct{}

// NewPlatformProvider 创建 macOS 平台的 PlatformProvider 实例。
func NewPlatformProvider() PlatformProvider {
	return &darwinPlatform{}
}

// GetAdapters 使用 networksetup 和 ifconfig 获取所有活动网络适配器及其 DNS 和 IP 配置。
func (d *darwinPlatform) GetAdapters() ([]NetworkAdapterInfo, error) {
	// Parse hardware ports from networksetup
	out, err := exec.Command("networksetup", "-listallhardwareports").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行 networksetup -listallhardwareports 失败: %w", err)
	}

	type hwPort struct {
		serviceName string
		device      string
	}

	var ports []hwPort
	blocks := strings.Split(string(out), "\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		var serviceName, device string
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Hardware Port:") {
				serviceName = strings.TrimSpace(strings.TrimPrefix(line, "Hardware Port:"))
			} else if strings.HasPrefix(line, "Device:") {
				device = strings.TrimSpace(strings.TrimPrefix(line, "Device:"))
			}
		}
		if serviceName != "" && device != "" {
			ports = append(ports, hwPort{serviceName: serviceName, device: device})
		}
	}

	var results []NetworkAdapterInfo
	for _, port := range ports {
		// Get IP addresses via ifconfig
		ipAddresses := getDeviceIPs(port.device)
		if len(ipAddresses) == 0 {
			// Only include adapters that have an IP address (are active)
			continue
		}

		// Get interface index
		ifIndex := 0
		iface, err := net.InterfaceByName(port.device)
		if err == nil {
			ifIndex = iface.Index
		}

		// Get DNS servers
		dnsServers := getDNSServers(port.serviceName)

		results = append(results, NetworkAdapterInfo{
			Name:         port.serviceName,
			InterfaceIdx: ifIndex,
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

// SetDNS 使用 networksetup 将指定 DNS 服务器设置到指定网络服务。
func (d *darwinPlatform) SetDNS(adapterName string, primaryDNS string, secondaryDNS string) error {
	if adapterName == "" {
		return fmt.Errorf("适配器名称不能为空")
	}
	if primaryDNS == "" {
		return fmt.Errorf("首选 DNS 不能为空")
	}

	primaryIP, err := extractDNSIP(primaryDNS)
	if err != nil {
		return fmt.Errorf("无法从 %q 提取 DNS IP 地址: %w", primaryDNS, err)
	}

	args := []string{"-setdnsservers", adapterName, primaryIP}
	if secondaryDNS != "" {
		secondaryIP, err := extractDNSIP(secondaryDNS)
		if err == nil && secondaryIP != "" {
			args = append(args, secondaryIP)
		}
	}

	out, err := exec.Command("networksetup", args...).CombinedOutput()
	if err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "requires authorization") ||
			strings.Contains(err.Error(), "exit status 1") {
			return fmt.Errorf("设置 DNS 失败：需要管理员权限。请以 sudo 运行应用或在系统偏好设置中授权")
		}
		return fmt.Errorf("设置 DNS 失败: %s: %w", outStr, err)
	}
	return nil
}

// ResetToAuto 使用 networksetup 将指定网络服务的 DNS 恢复为自动获取（DHCP）。
func (d *darwinPlatform) ResetToAuto(adapterName string) error {
	if adapterName == "" {
		return fmt.Errorf("适配器名称不能为空")
	}

	out, err := exec.Command("networksetup", "-setdnsservers", adapterName, "Empty").CombinedOutput()
	if err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "requires authorization") ||
			strings.Contains(err.Error(), "exit status 1") {
			return fmt.Errorf("重置 DNS 失败：需要管理员权限。请以 sudo 运行应用或在系统偏好设置中授权")
		}
		return fmt.Errorf("重置 DNS 失败: %s: %w", outStr, err)
	}
	return nil
}

// CheckAdmin 检查当前进程是否以 root 权限运行。
func (d *darwinPlatform) CheckAdmin() bool {
	return os.Getuid() == 0
}

// GetSystemTheme 通过 defaults read 检测 macOS 系统主题。
// 返回 "dark"（深色模式）或 "light"（浅色模式）。
func (d *darwinPlatform) GetSystemTheme() string {
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").CombinedOutput()
	if err != nil {
		return "light"
	}
	if strings.TrimSpace(strings.ToLower(string(out))) == "dark" {
		return "dark"
	}
	return "light"
}

// --- 内部辅助函数 ---

// getDeviceIPs 通过 ifconfig 获取指定设备的 IPv4 地址列表。
func getDeviceIPs(device string) []string {
	out, err := exec.Command("ifconfig", device).CombinedOutput()
	if err != nil {
		return nil
	}

	var ips []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ip := fields[1]
				if ip != "" && ip != "0.0.0.0" {
					ips = append(ips, ip)
				}
			}
		}
	}
	return ips
}

// getDNSServers 通过 networksetup -getdnsservers 获取指定服务的 DNS 服务器列表。
func getDNSServers(serviceName string) []string {
	out, err := exec.Command("networksetup", "-getdnsservers", serviceName).CombinedOutput()
	if err != nil {
		return []string{}
	}

	outStr := strings.TrimSpace(string(out))
	// networksetup returns "There aren't any DNS Servers set on ..." when using DHCP
	if strings.Contains(outStr, "There aren't any DNS Servers") {
		return []string{}
	}

	var servers []string
	for _, line := range strings.Split(outStr, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && net.ParseIP(line) != nil {
			servers = append(servers, line)
		}
	}
	if servers == nil {
		servers = []string{}
	}
	return servers
}

package backend

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// DNSConfigService 通过 Windows API 和注册表读取、修改系统 DNS 配置。
// 完全不依赖 PowerShell / cmd，无子进程，不触发杀软告警。
type DNSConfigService struct{}

// NewDNSConfigService 创建 DNSConfigService 实例。
func NewDNSConfigService() *DNSConfigService {
	return &DNSConfigService{}
}

// GetAdapters 通过 iphlpapi.GetAdaptersAddresses 获取所有活动网络适配器及其 DNS 和 IP 配置。
func (d *DNSConfigService) GetAdapters() ([]NetworkAdapterInfo, error) {
	const (
		AF_INET             = 2
		GAA_FLAG_SKIP_ANYCAST   = 0x0002
		GAA_FLAG_SKIP_MULTICAST = 0x0004
		GAA_FLAG_INCLUDE_PREFIX = 0x0010
	)
	flags := uint32(GAA_FLAG_SKIP_ANYCAST | GAA_FLAG_SKIP_MULTICAST | GAA_FLAG_INCLUDE_PREFIX)

	var size uint32
	iphlpapi := syscall.NewLazyDLL("iphlpapi.dll")
	getAdaptersAddresses := iphlpapi.NewProc("GetAdaptersAddresses")

	ret, _, _ := getAdaptersAddresses.Call(
		uintptr(AF_INET), uintptr(flags), 0, 0, uintptr(unsafe.Pointer(&size)),
	)
	if ret != 111 && ret != 0 {
		return nil, fmt.Errorf("GetAdaptersAddresses 探测大小失败: %d", ret)
	}

	size += 4096
	buf := make([]byte, size)
	ret, _, _ = getAdaptersAddresses.Call(
		uintptr(AF_INET), uintptr(flags), 0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("GetAdaptersAddresses 失败: %d", ret)
	}

	type socketAddress struct {
		Sockaddr       *syscall.RawSockaddrAny
		SockaddrLength int32
	}
	// IP_ADAPTER_UNICAST_ADDRESS 简化版（只需要 Next 和 Address）
	type ipAdapterUnicastAddress struct {
		Length             uint32
		Flags              uint32
		Next               *ipAdapterUnicastAddress
		Address            socketAddress
		PrefixOrigin       uint32
		SuffixOrigin       uint32
		DadState           uint32
		ValidLifetime      uint32
		PreferredLifetime  uint32
		LeaseLifetime      uint32
		OnLinkPrefixLength uint8
	}
	type ipAdapterDNSServerAddress struct {
		Length   uint32
		Reserved uint32
		Next     *ipAdapterDNSServerAddress
		Address  socketAddress
	}
	type ipAdapterAddresses struct {
		Length                uint32
		IfIndex               uint32
		Next                  *ipAdapterAddresses
		AdapterName           *byte
		FirstUnicastAddress   *ipAdapterUnicastAddress // 改为正确的指针类型
		FirstAnycastAddress   uintptr
		FirstMulticastAddress uintptr
		FirstDNSServerAddress *ipAdapterDNSServerAddress
		DNSSuffix             *uint16
		Description           *uint16
		FriendlyName          *uint16
		PhysicalAddress       [8]byte
		PhysicalAddressLength uint32
		Flags                 uint32
		Mtu                   uint32
		IfType                uint32
		OperStatus            uint32
	}

	var results []NetworkAdapterInfo
	adapter := (*ipAdapterAddresses)(unsafe.Pointer(&buf[0]))
	for adapter != nil {
		if adapter.OperStatus == 1 {
			name := windows.UTF16PtrToString(adapter.FriendlyName)
			ifIndex := int(adapter.IfIndex)

			// 读取适配器自身的 IPv4 地址
			var ipAddresses []string
			unicast := adapter.FirstUnicastAddress
			for unicast != nil {
				if sa := unicast.Address.Sockaddr; sa != nil {
					ip := sockaddrToIP(sa)
					if ip != "" && ip != "0.0.0.0" {
						ipAddresses = append(ipAddresses, ip)
					}
				}
				unicast = unicast.Next
			}
			if ipAddresses == nil {
				ipAddresses = []string{}
			}

			// 读取 DNS 服务器地址
			var dnsServers []string
			dnsAddr := adapter.FirstDNSServerAddress
			for dnsAddr != nil {
				if sa := dnsAddr.Address.Sockaddr; sa != nil {
					ip := sockaddrToIP(sa)
					if ip != "" && !strings.HasPrefix(ip, "fec0") {
						dnsServers = append(dnsServers, ip)
					}
				}
				dnsAddr = dnsAddr.Next
			}
			if dnsServers == nil {
				dnsServers = []string{}
			}

			results = append(results, NetworkAdapterInfo{
				Name:         name,
				InterfaceIdx: ifIndex,
				Status:       "Up",
				IPAddresses:  ipAddresses,
				CurrentDNS:   dnsServers,
			})
		}
		adapter = adapter.Next
	}

	if results == nil {
		results = []NetworkAdapterInfo{}
	}
	return results, nil
}

// SetDNS 通过写注册表设置指定适配器的首选和备用 DNS 服务器。
// 支持 DoH URL（自动提取 IP）、DoT 域名（需要 IP）、普通 IP。
// Windows NameServer 格式：多个地址用逗号分隔，如 "8.8.8.8,8.8.4.4"。
func (d *DNSConfigService) SetDNS(adapterName string, primaryDNS string, secondaryDNS string) error {
	if adapterName == "" {
		return fmt.Errorf("适配器名称不能为空")
	}
	if primaryDNS == "" {
		return fmt.Errorf("首选 DNS 不能为空")
	}

	// 提取可用于系统 DNS 的 IP 地址（处理 DoH URL 和 DoT 域名）
	primaryIP, err := extractDNSIP(primaryDNS)
	if err != nil {
		return fmt.Errorf("无法从 %q 提取 DNS IP 地址: %w", primaryDNS, err)
	}

	// 构建 NameServer 值
	nameServer := primaryIP
	if secondaryDNS != "" {
		secondaryIP, err := extractDNSIP(secondaryDNS)
		if err == nil && secondaryIP != "" {
			nameServer = primaryIP + "," + secondaryIP
		}
	}

	guid, err := getAdapterGUID(adapterName)
	if err != nil {
		return fmt.Errorf("获取适配器 %q 的 GUID 失败: %w", adapterName, err)
	}

	keyPath := `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters\Interfaces\` + guid
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表键失败 (需要管理员权限): %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue("NameServer", nameServer); err != nil {
		return fmt.Errorf("写入 NameServer 失败: %w", err)
	}
	_ = k.SetStringValue("DhcpNameServer", "")

	notifyDNSChange(guid)
	flushDNSCache()
	return nil
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

// ResetToAuto 清空注册表 NameServer，恢复为 DHCP 自动获取 DNS。
func (d *DNSConfigService) ResetToAuto(adapterName string) error {
	if adapterName == "" {
		return fmt.Errorf("适配器名称不能为空")
	}

	guid, err := getAdapterGUID(adapterName)
	if err != nil {
		return fmt.Errorf("获取适配器 %q 的 GUID 失败: %w", adapterName, err)
	}

	keyPath := `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters\Interfaces\` + guid
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表键失败 (需要管理员权限): %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue("NameServer", ""); err != nil {
		return fmt.Errorf("清空 NameServer 失败: %w", err)
	}

	notifyDNSChange(guid)
	flushDNSCache()
	return nil
}

// CheckAdmin 通过 shell32.IsUserAnAdmin 检查管理员权限。
func (d *DNSConfigService) CheckAdmin() bool {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	ret, _, _ := shell32.NewProc("IsUserAnAdmin").Call()
	return ret != 0
}

// --- 内部辅助函数 ---

// notifyDNSChange 通过发送 WM_SETTINGCHANGE 广播消息通知系统 DNS 配置已变更，
// 同时尝试调用 NotifyIpInterfaceChange 触发网络栈重新加载。
func notifyDNSChange(guid string) {
	// 方法1：广播 WM_SETTINGCHANGE，通知所有窗口配置已变更
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")
	paramPtr, _ := syscall.UTF16PtrFromString("Environment")
	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
	)
	var result uintptr
	sendMessageTimeout.Call(
		HWND_BROADCAST,
		WM_SETTINGCHANGE,
		0,
		uintptr(unsafe.Pointer(paramPtr)),
		SMTO_ABORTIFHUNG,
		100,
		uintptr(unsafe.Pointer(&result)),
	)

	// 方法2：通过 iphlpapi 的 IpRenewAddress 触发接口重新获取配置
	// 注意：这会短暂中断连接，仅在必要时使用
	// 这里我们只做轻量级通知，不做 renew
}

// getAdapterGUID 通过适配器友好名称在注册表中查找其 GUID。
// 同时在 Tcpip\Parameters\Interfaces 下验证该 GUID 存在。
func getAdapterGUID(friendlyName string) (string, error) {
	const netKey = `SYSTEM\CurrentControlSet\Control\Network\{4D36E972-E325-11CE-BFC1-08002BE10318}`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, netKey, registry.READ)
	if err != nil {
		return "", fmt.Errorf("打开网络注册表键失败: %w", err)
	}
	defer k.Close()

	guids, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return "", fmt.Errorf("读取网络子键失败: %w", err)
	}

	for _, guid := range guids {
		connPath := netKey + `\` + guid + `\Connection`
		ck, err := registry.OpenKey(registry.LOCAL_MACHINE, connPath, registry.READ)
		if err != nil {
			continue
		}
		name, _, err := ck.GetStringValue("Name")
		ck.Close()
		if err != nil || !strings.EqualFold(name, friendlyName) {
			continue
		}

		// 验证该 GUID 在 Tcpip\Parameters\Interfaces 下存在
		tcpipPath := `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters\Interfaces\` + guid
		tk, err := registry.OpenKey(registry.LOCAL_MACHINE, tcpipPath, registry.READ)
		if err != nil {
			// 不在 Tcpip 接口列表中，跳过
			continue
		}
		tk.Close()
		return guid, nil
	}

	return "", fmt.Errorf("未找到适配器 %q 对应的可配置 GUID", friendlyName)
}

// flushDNSCache 刷新本地 DNS 解析缓存。
func flushDNSCache() {
	dnsapi := syscall.NewLazyDLL("dnsapi.dll")
	dnsapi.NewProc("DnsFlushResolverCache").Call()
}

// sockaddrToIP 将 RawSockaddrAny 转换为 IP 字符串。
func sockaddrToIP(sa *syscall.RawSockaddrAny) string {
	if sa == nil {
		return ""
	}
	switch sa.Addr.Family {
	case syscall.AF_INET:
		addr := (*syscall.RawSockaddrInet4)(unsafe.Pointer(sa))
		return net.IP(addr.Addr[:]).String()
	case syscall.AF_INET6:
		addr := (*syscall.RawSockaddrInet6)(unsafe.Pointer(sa))
		return net.IP(addr.Addr[:]).String()
	}
	return ""
}

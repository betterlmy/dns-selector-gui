//go:build windows

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

// windowsPlatform 实现 PlatformProvider，使用 Windows API。
type windowsPlatform struct{}

// NewPlatformProvider 创建当前平台的 PlatformProvider 实例。
func NewPlatformProvider() PlatformProvider {
	return &windowsPlatform{}
}

// GetAdapters 通过 iphlpapi.GetAdaptersAddresses 获取所有活动网络适配器及其 DNS 和 IP 配置。
func (w *windowsPlatform) GetAdapters() ([]NetworkAdapterInfo, error) {
	const (
		AF_INET                 = 2
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
		FirstUnicastAddress   *ipAdapterUnicastAddress
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
func (w *windowsPlatform) SetDNS(adapterName string, primaryDNS string, secondaryDNS string) error {
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

// ResetToAuto 清空注册表 NameServer，恢复为 DHCP 自动获取 DNS。
func (w *windowsPlatform) ResetToAuto(adapterName string) error {
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
func (w *windowsPlatform) CheckAdmin() bool {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	ret, _, _ := shell32.NewProc("IsUserAnAdmin").Call()
	return ret != 0
}

// GetSystemTheme 通过读取注册表检测 Windows 系统主题设置。
// 返回 "dark"（深色模式）或 "light"（浅色模式）。
func (w *windowsPlatform) GetSystemTheme() string {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		registry.QUERY_VALUE)
	if err != nil {
		return "light"
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return "light"
	}
	if val == 0 {
		return "dark"
	}
	return "light"
}

// --- 内部辅助函数 ---

// notifyDNSChange 通过发送 WM_SETTINGCHANGE 广播消息通知系统 DNS 配置已变更。
func notifyDNSChange(guid string) {
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
}

// getAdapterGUID 通过适配器友好名称在注册表中查找其 GUID。
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

		tcpipPath := `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters\Interfaces\` + guid
		tk, err := registry.OpenKey(registry.LOCAL_MACHINE, tcpipPath, registry.READ)
		if err != nil {
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

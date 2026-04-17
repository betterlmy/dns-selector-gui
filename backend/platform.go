package backend

// PlatformProvider 定义跨平台操作系统功能的统一接口。
// 各平台通过 build tags 提供独立实现（platform_windows.go、platform_darwin.go、platform_linux.go）。
type PlatformProvider interface {
	// GetAdapters 返回所有活动网络适配器及其 DNS 和 IP 配置。
	GetAdapters() ([]NetworkAdapterInfo, error)

	// SetDNS 将指定 DNS 服务器设置到指定网络适配器。
	// primaryDNS 为首选 DNS，secondaryDNS 为备用 DNS（可为空）。
	SetDNS(adapterName string, primaryDNS string, secondaryDNS string) error

	// ResetToAuto 将指定网络适配器的 DNS 恢复为自动获取（DHCP）。
	ResetToAuto(adapterName string) error

	// CheckAdmin 检查当前进程是否具有修改系统 DNS 配置的权限。
	CheckAdmin() bool

	// GetSystemTheme 检测系统主题，返回 "dark" 或 "light"。
	GetSystemTheme() string
}

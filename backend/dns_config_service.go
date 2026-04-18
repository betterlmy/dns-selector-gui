package backend

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


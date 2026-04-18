package backend

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// domainRegexp 匹配有效的域名格式（兼容 RFC 1123）。
var domainRegexp = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)

// ValidateServerAddress 根据协议类型验证 DNS 服务器地址。
// 验证通过返回 nil，否则返回包含格式说明的错误信息。
func ValidateServerAddress(protocol, address string) error {
	switch strings.ToLower(protocol) {
	case "udp":
		return validateUDPAddress(address)
	case "dot":
		return validateDoTAddress(address)
	case "doh":
		return validateDoHAddress(address)
	default:
		return fmt.Errorf("不支持的协议: %q，有效值为 udp、dot、doh", protocol)
	}
}

// validateUDPAddress 验证地址是否为有效的 IPv4 地址。
func validateUDPAddress(address string) error {
	ip := net.ParseIP(address)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("无效的 UDP 地址：必须为有效的 IPv4 地址（例: 8.8.8.8）")
	}
	return nil
}

// validateDoTAddress 验证地址是否为有效的域名或 "IP@TLSServerName" 格式。
func validateDoTAddress(address string) error {
	if address == "" {
		return fmt.Errorf("无效的 DoT 地址：必须为有效域名（例: dns.google）或 IP@TLSServerName 格式（例: 8.8.8.8@dns.google）")
	}

	// 检查 IP@TLSServerName 格式
	if strings.Contains(address, "@") {
		parts := strings.SplitN(address, "@", 2)
		ipPart := parts[0]
		tlsPart := parts[1]

		ip := net.ParseIP(ipPart)
		if ip == nil || ip.To4() == nil {
			return fmt.Errorf("无效的 DoT 地址：必须为有效域名（例: dns.google）或 IP@TLSServerName 格式（例: 8.8.8.8@dns.google）")
		}
		if tlsPart == "" {
			return fmt.Errorf("无效的 DoT 地址：必须为有效域名（例: dns.google）或 IP@TLSServerName 格式（例: 8.8.8.8@dns.google）")
		}
		return nil
	}

	// 否则必须为有效域名
	if !isValidDomain(address) {
		return fmt.Errorf("无效的 DoT 地址：必须为有效域名（例: dns.google）或 IP@TLSServerName 格式（例: 8.8.8.8@dns.google）")
	}
	return nil
}

// validateDoHAddress 验证地址是否为有效的 HTTPS URL 或 "https://IP/path@TLSServerName" 格式。
func validateDoHAddress(address string) error {
	if address == "" {
		return fmt.Errorf("无效的 DoH 地址：必须为有效的 HTTPS URL（例: https://dns.google/dns-query）或 https://IP/path@TLSServerName 格式")
	}

	// 检查 @TLSServerName 格式：按最后一个 "@" 分割
	if idx := strings.LastIndex(address, "@"); idx != -1 {
		urlPart := address[:idx]
		tlsPart := address[idx+1:]

		if tlsPart == "" {
			return fmt.Errorf("无效的 DoH 地址：必须为有效的 HTTPS URL（例: https://dns.google/dns-query）或 https://IP/path@TLSServerName 格式")
		}

		if !isValidHTTPSURL(urlPart) {
			return fmt.Errorf("无效的 DoH 地址：必须为有效的 HTTPS URL（例: https://dns.google/dns-query）或 https://IP/path@TLSServerName 格式")
		}
		return nil
	}

	// 否则必须为有效的 HTTPS URL
	if !isValidHTTPSURL(address) {
		return fmt.Errorf("无效的 DoH 地址：必须为有效的 HTTPS URL（例: https://dns.google/dns-query）或 https://IP/path@TLSServerName 格式")
	}
	return nil
}

// ValidateTestParams 验证测试参数。
// queries、warmup、concurrency 必须为正整数，timeout 必须为正数。
func ValidateTestParams(params TestParams) error {
	if params.Queries <= 0 {
		return fmt.Errorf("queries 必须为正整数，当前值: %d", params.Queries)
	}
	if params.Warmup <= 0 {
		return fmt.Errorf("warmup 必须为正整数，当前值: %d", params.Warmup)
	}
	if params.Concurrency <= 0 {
		return fmt.Errorf("concurrency 必须为正整数，当前值: %d", params.Concurrency)
	}
	if params.Timeout <= 0 {
		return fmt.Errorf("timeout 必须为正数，当前值: %f", params.Timeout)
	}
	return nil
}

// ValidateDomain 验证输入是否为有效的域名格式。
func ValidateDomain(domain string) error {
	if !isValidDomain(domain) {
		return fmt.Errorf("无效的域名格式: %q，必须为有效域名（例: example.com）", domain)
	}
	return nil
}

// isValidDomain 检查字符串是否为有效域名。
func isValidDomain(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	return domainRegexp.MatchString(s)
}

// isValidHTTPSURL 检查字符串是否为有效的 HTTPS URL（host 不为空）。
func isValidHTTPSURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return strings.ToLower(u.Scheme) == "https" && u.Host != ""
}

// ValidateBootstrapIP 验证 BootstrapIP 是否为合法的 IPv4 地址（允许为空）。
func ValidateBootstrapIP(ip string) error {
	if ip == "" {
		return nil
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil || parsed.To4() == nil {
		return fmt.Errorf("BootstrapIP %q 不是合法的 IPv4 地址", ip)
	}
	return nil
}

// ValidateTLSServerName 验证 TLSServerName 是否为合法域名（允许为空）。
func ValidateTLSServerName(name string) error {
	if name == "" {
		return nil
	}
	if !isValidDomain(name) {
		return fmt.Errorf("TLSServerName %q 不是合法的域名格式", name)
	}
	return nil
}

// ValidateServerEntry 对添加服务器请求做完整校验（地址 + 附加字段）。
func ValidateServerEntry(req AddServerRequest) error {
	if err := ValidateServerAddress(req.Protocol, req.Address); err != nil {
		return err
	}
	if err := ValidateBootstrapIP(req.BootstrapIP); err != nil {
		return err
	}
	if err := ValidateTLSServerName(req.TLSServerName); err != nil {
		return err
	}
	return nil
}

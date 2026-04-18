package backend

import (
	"testing"

	"github.com/betterlmy/dns-selector/selector"
)

func TestResolveSystemDNSAddresses(t *testing.T) {
	tests := []struct {
		name      string
		server    selector.DNSServer
		wantAddrs []string
		wantErr   bool
	}{
		{
			name:      "UDP 直接返回地址",
			server:    selector.DNSServer{Protocol: "udp", Address: "8.8.8.8"},
			wantAddrs: []string{"8.8.8.8"},
		},
		{
			name:      "DoT IP 型直接返回",
			server:    selector.DNSServer{Protocol: "dot", Address: "1.1.1.1"},
			wantAddrs: []string{"1.1.1.1"},
		},
		{
			name:      "DoT IP@TLSServerName 格式提取 IP",
			server:    selector.DNSServer{Protocol: "dot", Address: "8.8.8.8@dns.google"},
			wantAddrs: []string{"8.8.8.8"},
		},
		{
			name: "DoT 域名型 + BootstrapIPs 返回 bootstrap",
			server: selector.DNSServer{
				Protocol:     "dot",
				Address:      "dns.google",
				BootstrapIPs: []string{"8.8.8.8", "8.8.4.4"},
			},
			wantAddrs: []string{"8.8.8.8", "8.8.4.4"},
		},
		{
			name:    "DoT 域名型无 BootstrapIPs 返回错误",
			server:  selector.DNSServer{Protocol: "dot", Address: "dns.google"},
			wantErr: true,
		},
		{
			name:      "DoH URL host 是 IP 返回 IP",
			server:    selector.DNSServer{Protocol: "doh", Address: "https://1.1.1.1/dns-query"},
			wantAddrs: []string{"1.1.1.1"},
		},
		{
			name: "DoH URL host 是域名 + BootstrapIPs 返回 bootstrap",
			server: selector.DNSServer{
				Protocol:     "doh",
				Address:      "https://dns.google/dns-query",
				BootstrapIPs: []string{"8.8.8.8"},
			},
			wantAddrs: []string{"8.8.8.8"},
		},
		{
			name:    "DoH URL host 是域名无 BootstrapIPs 返回错误",
			server:  selector.DNSServer{Protocol: "doh", Address: "https://dns.google/dns-query"},
			wantErr: true,
		},
		{
			name:    "BootstrapIP 格式非法返回错误",
			server:  selector.DNSServer{Protocol: "dot", Address: "dns.google", BootstrapIPs: []string{"not-an-ip"}},
			wantErr: true,
		},
		{
			name:      "DoH URL 带端口 IP host",
			server:    selector.DNSServer{Protocol: "doh", Address: "https://1.1.1.1:443/dns-query"},
			wantAddrs: []string{"1.1.1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addrs, err := ResolveSystemDNSAddresses(tt.server)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误但未返回错误，got addrs=%v", addrs)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外错误: %v", err)
			}
			if len(addrs) != len(tt.wantAddrs) {
				t.Fatalf("地址数量不符: got %v, want %v", addrs, tt.wantAddrs)
			}
			for i, a := range addrs {
				if a != tt.wantAddrs[i] {
					t.Errorf("addrs[%d]: got %q, want %q", i, a, tt.wantAddrs[i])
				}
			}
		})
	}
}

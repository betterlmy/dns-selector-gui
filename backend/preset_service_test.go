package backend

import (
	"testing"

	"github.com/betterlmy/dns-selector/selector"
)

func TestSwitchPresetPreservesCustomData(t *testing.T) {
	p := NewPresetService()

	_ = p.AddCustomServer(selector.DNSServer{
		Name: "my-server", Address: "1.2.3.4", Protocol: "udp",
	})
	_ = p.AddCustomDomain("example.com")

	if err := p.SwitchPreset("global"); err != nil {
		t.Fatalf("SwitchPreset 失败: %v", err)
	}

	if len(p.GetCustomServers()) != 1 {
		t.Errorf("切换预设后 customServers 应保留 1 条，got %d", len(p.GetCustomServers()))
	}
	if p.GetCustomServers()[0].Address != "1.2.3.4" {
		t.Errorf("customServer 地址应为 1.2.3.4，got %q", p.GetCustomServers()[0].Address)
	}
	if len(p.GetCustomDomains()) != 1 || p.GetCustomDomains()[0] != "example.com" {
		t.Errorf("切换预设后 customDomains 应保留，got %v", p.GetCustomDomains())
	}

	_ = p.SwitchPreset("cn")
	if len(p.GetCustomServers()) != 1 || len(p.GetCustomDomains()) != 1 {
		t.Error("多次切换预设后自定义数据应始终保留")
	}
}

func TestSwitchPresetInvalidName(t *testing.T) {
	p := NewPresetService()
	if err := p.SwitchPreset("invalid"); err == nil {
		t.Error("无效预设名应返回错误")
	}
}

func TestSwitchPresetChangesCurrentPreset(t *testing.T) {
	p := NewPresetService()
	if p.GetCurrentPreset() != "cn" {
		t.Fatalf("初始预设应为 cn，got %q", p.GetCurrentPreset())
	}
	_ = p.SwitchPreset("global")
	if p.GetCurrentPreset() != "global" {
		t.Errorf("切换后预设应为 global，got %q", p.GetCurrentPreset())
	}
}

func TestRemoveCustomServerByCompositeKey(t *testing.T) {
	p := NewPresetService()

	// 使用不在任何预设中的自定义地址
	udp := selector.DNSServer{Name: "custom", Address: "10.0.0.1", Protocol: "udp"}
	dot := selector.DNSServer{Name: "custom", Address: "10.0.0.1", Protocol: "dot", TLSServerName: "custom.dns"}
	_ = p.AddCustomServer(udp)
	_ = p.AddCustomServer(dot)

	if len(p.GetCustomServers()) != 2 {
		t.Fatalf("应有 2 条记录，got %d", len(p.GetCustomServers()))
	}

	// 只删 UDP，DoT 应保留
	if err := p.RemoveCustomServer("udp", "10.0.0.1", ""); err != nil {
		t.Fatalf("删除 UDP 记录失败: %v", err)
	}
	if len(p.GetCustomServers()) != 1 {
		t.Errorf("删除后应剩 1 条，got %d", len(p.GetCustomServers()))
	}
	if p.GetCustomServers()[0].Protocol != "dot" {
		t.Errorf("剩余记录应为 DoT，got %q", p.GetCustomServers()[0].Protocol)
	}
	if p.GetCustomServers()[0].Address != "10.0.0.1" {
		t.Errorf("剩余 DoT 地址应为 10.0.0.1，got %q", p.GetCustomServers()[0].Address)
	}
}

func TestIsPresetServerByCompositeKey(t *testing.T) {
	p := NewPresetService()
	// CN 预设中 8.8.8.8 是 UDP
	if !p.IsPresetServer("udp", "8.8.8.8", "") {
		t.Error("8.8.8.8 UDP 应被识别为预设项")
	}
	// 同地址但不同协议不应被识别为预设
	if p.IsPresetServer("dot", "8.8.8.8", "") {
		t.Error("8.8.8.8 DoT 不应被识别为 CN 预设项")
	}
}

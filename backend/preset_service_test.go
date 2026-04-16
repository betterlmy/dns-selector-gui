package backend

import (
	"testing"

	"github.com/betterlmy/dns-selector/selector"
)

func TestNewPresetService(t *testing.T) {
	ps := NewPresetService()
	if ps.GetCurrentPreset() != "cn" {
		t.Errorf("GetCurrentPreset: got %q, want %q", ps.GetCurrentPreset(), "cn")
	}
	// Should have preset servers only, no custom ones
	servers := ps.GetMergedServers()
	cnPreset, _ := GetPreset("cn")
	if len(servers) != len(cnPreset.Servers) {
		t.Errorf("GetMergedServers: got %d, want %d", len(servers), len(cnPreset.Servers))
	}
	domains := ps.GetMergedDomains()
	if len(domains) != len(cnPreset.Domains) {
		t.Errorf("GetMergedDomains: got %d, want %d", len(domains), len(cnPreset.Domains))
	}
}

func TestSwitchPreset(t *testing.T) {
	ps := NewPresetService()

	// Add custom items first
	_ = ps.AddCustomServer(selector.DNSServer{Name: "Custom", Address: "9.9.9.9", Protocol: "udp"})
	_ = ps.AddCustomDomain("custom.example.com")

	// Switch to global — custom items should be cleared
	if err := ps.SwitchPreset("global"); err != nil {
		t.Fatalf("SwitchPreset(global): %v", err)
	}
	if ps.GetCurrentPreset() != "global" {
		t.Errorf("GetCurrentPreset: got %q, want %q", ps.GetCurrentPreset(), "global")
	}
	globalPreset, _ := GetPreset("global")
	if len(ps.GetMergedServers()) != len(globalPreset.Servers) {
		t.Errorf("custom servers not cleared after switch")
	}
	if len(ps.GetMergedDomains()) != len(globalPreset.Domains) {
		t.Errorf("custom domains not cleared after switch")
	}

	// Switch back to cn
	if err := ps.SwitchPreset("cn"); err != nil {
		t.Fatalf("SwitchPreset(cn): %v", err)
	}
	if ps.GetCurrentPreset() != "cn" {
		t.Errorf("GetCurrentPreset: got %q, want %q", ps.GetCurrentPreset(), "cn")
	}
}

func TestSwitchPresetInvalid(t *testing.T) {
	ps := NewPresetService()
	if err := ps.SwitchPreset("invalid"); err == nil {
		t.Error("SwitchPreset(invalid): expected error, got nil")
	}
	// Preset should remain unchanged
	if ps.GetCurrentPreset() != "cn" {
		t.Errorf("preset changed after invalid switch: got %q", ps.GetCurrentPreset())
	}
}

func TestGetMergedServers(t *testing.T) {
	ps := NewPresetService()
	cnPreset, _ := GetPreset("cn")

	custom := selector.DNSServer{Name: "MyDNS", Address: "9.9.9.9", Protocol: "udp"}
	_ = ps.AddCustomServer(custom)

	merged := ps.GetMergedServers()
	want := len(cnPreset.Servers) + 1
	if len(merged) != want {
		t.Fatalf("GetMergedServers: got %d, want %d", len(merged), want)
	}
	// Preset servers come first
	for i, s := range cnPreset.Servers {
		if merged[i].Address != s.Address {
			t.Errorf("server[%d]: got %q, want %q", i, merged[i].Address, s.Address)
		}
	}
	// Custom server is last
	last := merged[len(merged)-1]
	if last.Address != "9.9.9.9" {
		t.Errorf("last server: got %q, want %q", last.Address, "9.9.9.9")
	}
}

func TestGetMergedDomains(t *testing.T) {
	ps := NewPresetService()
	cnPreset, _ := GetPreset("cn")

	_ = ps.AddCustomDomain("custom.example.com")

	merged := ps.GetMergedDomains()
	want := len(cnPreset.Domains) + 1
	if len(merged) != want {
		t.Fatalf("GetMergedDomains: got %d, want %d", len(merged), want)
	}
	last := merged[len(merged)-1]
	if last != "custom.example.com" {
		t.Errorf("last domain: got %q, want %q", last, "custom.example.com")
	}
}

func TestIsPresetItem(t *testing.T) {
	ps := NewPresetService()
	cnPreset, _ := GetPreset("cn")

	// A preset server address should be recognized
	if !ps.IsPresetItem(cnPreset.Servers[0].Address) {
		t.Errorf("IsPresetItem(%q): got false, want true", cnPreset.Servers[0].Address)
	}
	// A preset domain should be recognized
	if !ps.IsPresetItem(cnPreset.Domains[0]) {
		t.Errorf("IsPresetItem(%q): got false, want true", cnPreset.Domains[0])
	}
	// A non-preset item should not be recognized
	if ps.IsPresetItem("nonexistent.example.com") {
		t.Error("IsPresetItem(nonexistent): got true, want false")
	}
}

func TestRemoveCustomServer(t *testing.T) {
	ps := NewPresetService()
	custom := selector.DNSServer{Name: "MyDNS", Address: "9.9.9.9", Protocol: "udp"}
	_ = ps.AddCustomServer(custom)

	before := len(ps.GetMergedServers())
	if err := ps.RemoveCustomServer("9.9.9.9"); err != nil {
		t.Fatalf("RemoveCustomServer: %v", err)
	}
	after := len(ps.GetMergedServers())
	if after != before-1 {
		t.Errorf("server count: got %d, want %d", after, before-1)
	}
}

func TestRemovePresetServerFails(t *testing.T) {
	ps := NewPresetService()
	cnPreset, _ := GetPreset("cn")
	addr := cnPreset.Servers[0].Address

	err := ps.RemoveCustomServer(addr)
	if err == nil {
		t.Fatal("RemoveCustomServer(preset): expected error, got nil")
	}
	if err.Error() != "无法删除预设服务器" {
		t.Errorf("error message: got %q, want %q", err.Error(), "无法删除预设服务器")
	}
}

func TestRemoveCustomDomain(t *testing.T) {
	ps := NewPresetService()
	_ = ps.AddCustomDomain("custom.example.com")

	before := len(ps.GetMergedDomains())
	if err := ps.RemoveCustomDomain("custom.example.com"); err != nil {
		t.Fatalf("RemoveCustomDomain: %v", err)
	}
	after := len(ps.GetMergedDomains())
	if after != before-1 {
		t.Errorf("domain count: got %d, want %d", after, before-1)
	}
}

func TestRemovePresetDomainFails(t *testing.T) {
	ps := NewPresetService()
	cnPreset, _ := GetPreset("cn")
	domain := cnPreset.Domains[0]

	err := ps.RemoveCustomDomain(domain)
	if err == nil {
		t.Fatal("RemoveCustomDomain(preset): expected error, got nil")
	}
	if err.Error() != "无法删除预设域名" {
		t.Errorf("error message: got %q, want %q", err.Error(), "无法删除预设域名")
	}
}

func TestRemoveNonexistentCustomServer(t *testing.T) {
	ps := NewPresetService()
	err := ps.RemoveCustomServer("nonexistent.addr")
	if err == nil {
		t.Error("RemoveCustomServer(nonexistent): expected error, got nil")
	}
}

func TestRemoveNonexistentCustomDomain(t *testing.T) {
	ps := NewPresetService()
	err := ps.RemoveCustomDomain("nonexistent.example.com")
	if err == nil {
		t.Error("RemoveCustomDomain(nonexistent): expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Task 2.6 – Preset completeness and switch-replaces-content tests
// ---------------------------------------------------------------------------

func TestCNPresetCompleteness(t *testing.T) {
	cnPreset, err := GetPreset("cn")
	if err != nil {
		t.Fatalf("GetPreset(cn): %v", err)
	}

	// Verify counts (source of truth: selector SDK)
	const wantServers = 32
	const wantDomains = 30
	if got := len(cnPreset.Servers); got != wantServers {
		t.Errorf("CN server count: got %d, want %d", got, wantServers)
	}
	if got := len(cnPreset.Domains); got != wantDomains {
		t.Errorf("CN domain count: got %d, want %d", got, wantDomains)
	}

	// Verify well-known servers exist
	wantAddrs := map[string]string{
		"223.5.5.5": "AliDNS",
		"8.8.8.8":   "Google",
		"1.1.1.1":   "Cloudflare",
	}
	addrSet := make(map[string]bool, len(cnPreset.Servers))
	for _, s := range cnPreset.Servers {
		addrSet[s.Address] = true
	}
	for addr, label := range wantAddrs {
		if !addrSet[addr] {
			t.Errorf("CN preset missing %s (%s)", label, addr)
		}
	}
}

func TestGlobalPresetCompleteness(t *testing.T) {
	globalPreset, err := GetPreset("global")
	if err != nil {
		t.Fatalf("GetPreset(global): %v", err)
	}

	// Verify counts (source of truth: selector SDK)
	const wantServers = 14
	const wantDomains = 24
	if got := len(globalPreset.Servers); got != wantServers {
		t.Errorf("Global server count: got %d, want %d", got, wantServers)
	}
	if got := len(globalPreset.Domains); got != wantDomains {
		t.Errorf("Global domain count: got %d, want %d", got, wantDomains)
	}

	// Verify well-known servers exist
	wantAddrs := map[string]string{
		"8.8.8.8": "Google",
		"1.1.1.1": "Cloudflare",
		"9.9.9.9": "Quad9",
	}
	addrSet := make(map[string]bool, len(globalPreset.Servers))
	for _, s := range globalPreset.Servers {
		addrSet[s.Address] = true
	}
	for addr, label := range wantAddrs {
		if !addrSet[addr] {
			t.Errorf("Global preset missing %s (%s)", label, addr)
		}
	}
}

func TestPresetSwitchReplacesContent(t *testing.T) {
	ps := NewPresetService()

	// Start on CN
	if ps.GetCurrentPreset() != "cn" {
		t.Fatalf("initial preset: got %q, want %q", ps.GetCurrentPreset(), "cn")
	}

	cnPreset, _ := GetPreset("cn")
	globalPreset, _ := GetPreset("global")

	// Verify initial content matches CN
	servers := ps.GetMergedServers()
	domains := ps.GetMergedDomains()
	if len(servers) != len(cnPreset.Servers) {
		t.Fatalf("initial servers: got %d, want %d", len(servers), len(cnPreset.Servers))
	}
	if len(domains) != len(cnPreset.Domains) {
		t.Fatalf("initial domains: got %d, want %d", len(domains), len(cnPreset.Domains))
	}

	// Switch to Global
	if err := ps.SwitchPreset("global"); err != nil {
		t.Fatalf("SwitchPreset(global): %v", err)
	}

	servers = ps.GetMergedServers()
	domains = ps.GetMergedDomains()
	if len(servers) != len(globalPreset.Servers) {
		t.Errorf("after switch to global, servers: got %d, want %d", len(servers), len(globalPreset.Servers))
	}
	if len(domains) != len(globalPreset.Domains) {
		t.Errorf("after switch to global, domains: got %d, want %d", len(domains), len(globalPreset.Domains))
	}
	// Spot-check: first server should match Global preset
	if servers[0].Address != globalPreset.Servers[0].Address {
		t.Errorf("first server after switch: got %q, want %q", servers[0].Address, globalPreset.Servers[0].Address)
	}
	// Spot-check: first domain should match Global preset
	if domains[0] != globalPreset.Domains[0] {
		t.Errorf("first domain after switch: got %q, want %q", domains[0], globalPreset.Domains[0])
	}

	// Switch back to CN
	if err := ps.SwitchPreset("cn"); err != nil {
		t.Fatalf("SwitchPreset(cn): %v", err)
	}

	servers = ps.GetMergedServers()
	domains = ps.GetMergedDomains()
	if len(servers) != len(cnPreset.Servers) {
		t.Errorf("after switch back to cn, servers: got %d, want %d", len(servers), len(cnPreset.Servers))
	}
	if len(domains) != len(cnPreset.Domains) {
		t.Errorf("after switch back to cn, domains: got %d, want %d", len(domains), len(cnPreset.Domains))
	}
	// Spot-check: first server should match CN preset
	if servers[0].Address != cnPreset.Servers[0].Address {
		t.Errorf("first server after switch back: got %q, want %q", servers[0].Address, cnPreset.Servers[0].Address)
	}
	// Spot-check: first domain should match CN preset
	if domains[0] != cnPreset.Domains[0] {
		t.Errorf("first domain after switch back: got %q, want %q", domains[0], cnPreset.Domains[0])
	}
}

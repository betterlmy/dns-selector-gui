package backend

import (
	"fmt"
	"testing"

	"github.com/betterlmy/dns-selector/selector"
	"pgregory.net/rapid"
)

// Feature: dns-selector-gui, Property 1: 预设项不可删除
// **Validates: Requirements 2.6**

func TestProperty1_PresetItemsCannotBeDeleted(t *testing.T) {
	// Feature: dns-selector-gui, Property 1: 预设项不可删除
	rapid.Check(t, func(t *rapid.T) {
		presetName := rapid.SampledFrom([]string{"cn", "global"}).Draw(t, "preset")
		ps := NewPresetService()
		if err := ps.SwitchPreset(presetName); err != nil {
			t.Fatalf("SwitchPreset(%q): %v", presetName, err)
		}
		preset, err := GetPreset(presetName)
		if err != nil {
			t.Fatalf("GetPreset(%q): %v", presetName, err)
		}

		// Test server deletion
		if len(preset.Servers) > 0 {
			idx := rapid.IntRange(0, len(preset.Servers)-1).Draw(t, "serverIdx")
			s := preset.Servers[idx]
			beforeLen := len(ps.GetMergedServers())
			err := ps.RemoveCustomServer(s.Protocol, s.Address, s.TLSServerName)
			if err == nil {
				t.Fatalf("expected error when deleting preset server %q", s.Address)
			}
			afterLen := len(ps.GetMergedServers())
			if afterLen != beforeLen {
				t.Fatalf("server list changed: before=%d, after=%d", beforeLen, afterLen)
			}
		}

		// Test domain deletion
		if len(preset.Domains) > 0 {
			idx := rapid.IntRange(0, len(preset.Domains)-1).Draw(t, "domainIdx")
			domain := preset.Domains[idx]
			beforeLen := len(ps.GetMergedDomains())
			err := ps.RemoveCustomDomain(domain)
			if err == nil {
				t.Fatalf("expected error when deleting preset domain %q", domain)
			}
			afterLen := len(ps.GetMergedDomains())
			if afterLen != beforeLen {
				t.Fatalf("domain list changed: before=%d, after=%d", beforeLen, afterLen)
			}
		}
	})
}

// Feature: dns-selector-gui, Property 3: 自定义项删除缩减列表
// **Validates: Requirements 5.7, 6.4**

func TestProperty3_CustomItemDeletionShrinksList(t *testing.T) {
	// Feature: dns-selector-gui, Property 3: 自定义项删除缩减列表
	rapid.Check(t, func(t *rapid.T) {
		ps := NewPresetService()

		// Generate 1-20 unique custom servers
		count := rapid.IntRange(1, 20).Draw(t, "count")
		for i := 0; i < count; i++ {
			addr := fmt.Sprintf("custom-%d.test", i)
			_ = ps.AddCustomServer(selector.DNSServer{Name: addr, Address: addr, Protocol: "udp"})
		}

		// Pick one to delete
		idx := rapid.IntRange(0, count-1).Draw(t, "deleteIdx")
		addr := fmt.Sprintf("custom-%d.test", idx)

		beforeLen := len(ps.GetMergedServers())
		err := ps.RemoveCustomServer("udp", addr, "")
		if err != nil {
			t.Fatalf("RemoveCustomServer(%q): %v", addr, err)
		}
		afterLen := len(ps.GetMergedServers())
		if afterLen != beforeLen-1 {
			t.Fatalf("server list: before=%d, after=%d, want %d", beforeLen, afterLen, beforeLen-1)
		}
		// Verify item is gone
		for _, s := range ps.GetMergedServers() {
			if s.Address == addr {
				t.Fatalf("deleted server %q still in list", addr)
			}
		}
	})

	// Same for domains
	rapid.Check(t, func(t *rapid.T) {
		ps := NewPresetService()

		count := rapid.IntRange(1, 20).Draw(t, "count")
		for i := 0; i < count; i++ {
			domain := fmt.Sprintf("custom-%d.example.test", i)
			_ = ps.AddCustomDomain(domain)
		}

		idx := rapid.IntRange(0, count-1).Draw(t, "deleteIdx")
		domain := fmt.Sprintf("custom-%d.example.test", idx)

		beforeLen := len(ps.GetMergedDomains())
		err := ps.RemoveCustomDomain(domain)
		if err != nil {
			t.Fatalf("RemoveCustomDomain(%q): %v", domain, err)
		}
		afterLen := len(ps.GetMergedDomains())
		if afterLen != beforeLen-1 {
			t.Fatalf("domain list: before=%d, after=%d, want %d", beforeLen, afterLen, beforeLen-1)
		}
		for _, d := range ps.GetMergedDomains() {
			if d == domain {
				t.Fatalf("deleted domain %q still in list", domain)
			}
		}
	})
}

// Feature: dns-selector-gui, Property 4: 添加有效域名扩展列表
// **Validates: Requirements 6.2**

func TestProperty4_AddValidDomainExtendsList(t *testing.T) {
	// Feature: dns-selector-gui, Property 4: 添加有效域名扩展列表
	rapid.Check(t, func(t *rapid.T) {
		ps := NewPresetService()

		// Generate a random domain that won't collide with presets
		label := rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "label")
		domain := label + ".example.test"

		beforeLen := len(ps.GetMergedDomains())
		err := ps.AddCustomDomain(domain)
		if err != nil {
			t.Fatalf("AddCustomDomain(%q): %v", domain, err)
		}
		afterLen := len(ps.GetMergedDomains())
		if afterLen != beforeLen+1 {
			t.Fatalf("domain list: before=%d, after=%d, want %d", beforeLen, afterLen, beforeLen+1)
		}

		found := false
		for _, d := range ps.GetMergedDomains() {
			if d == domain {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("added domain %q not found in merged list", domain)
		}
	})
}

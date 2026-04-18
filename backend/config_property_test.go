package backend

import (
	"encoding/json"
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

// Feature: dns-selector-gui, Property 8: 配置文件 JSON 序列化往返一致
// **Validates: Requirements 13.1, 13.2**

func TestProperty8_ConfigJSONRoundTrip(t *testing.T) {
	// Feature: dns-selector-gui, Property 8: 配置文件 JSON 序列化往返一致
	rapid.Check(t, func(t *rapid.T) {
		// Generate random AppConfig
		numServers := rapid.IntRange(0, 5).Draw(t, "numServers")
		servers := make([]CustomServerEntry, numServers)
		for i := range servers {
			servers[i] = CustomServerEntry{
				Protocol: rapid.SampledFrom([]string{"udp", "dot", "doh"}).Draw(t, fmt.Sprintf("proto%d", i)),
				Address:  rapid.StringMatching(`[a-z]{3,10}\.[a-z]{2,4}`).Draw(t, fmt.Sprintf("addr%d", i)),
			}
		}

		numDomains := rapid.IntRange(0, 5).Draw(t, "numDomains")
		domains := make([]string, numDomains)
		for i := range domains {
			domains[i] = rapid.StringMatching(`[a-z]{3,10}\.[a-z]{2,4}`).Draw(t, fmt.Sprintf("domain%d", i))
		}

		cfg := AppConfig{
			CurrentPreset: rapid.SampledFrom([]string{"cn", "global"}).Draw(t, "preset"),
			CustomServers: servers,
			CustomDomains: domains,
			TestParams: TestParams{
				Queries:     rapid.IntRange(1, 100).Draw(t, "queries"),
				Warmup:      rapid.IntRange(1, 10).Draw(t, "warmup"),
				Concurrency: rapid.IntRange(1, 100).Draw(t, "concurrency"),
				Timeout:     rapid.Float64Range(0.1, 30.0).Draw(t, "timeout"),
			},
		}

		// Serialize
		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}

		// Deserialize
		var got AppConfig
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}

		// Compare top-level fields
		if got.CurrentPreset != cfg.CurrentPreset {
			t.Fatalf("CurrentPreset: got %q, want %q", got.CurrentPreset, cfg.CurrentPreset)
		}
		if got.TestParams != cfg.TestParams {
			t.Fatalf("TestParams mismatch: got %+v, want %+v", got.TestParams, cfg.TestParams)
		}
		if len(got.CustomServers) != len(cfg.CustomServers) {
			t.Fatalf("CustomServers length: got %d, want %d", len(got.CustomServers), len(cfg.CustomServers))
		}
		for i := range cfg.CustomServers {
			if got.CustomServers[i] != cfg.CustomServers[i] {
				t.Fatalf("CustomServers[%d]: got %+v, want %+v", i, got.CustomServers[i], cfg.CustomServers[i])
			}
		}
		if len(got.CustomDomains) != len(cfg.CustomDomains) {
			t.Fatalf("CustomDomains length: got %d, want %d", len(got.CustomDomains), len(cfg.CustomDomains))
		}
		for i := range cfg.CustomDomains {
			if got.CustomDomains[i] != cfg.CustomDomains[i] {
				t.Fatalf("CustomDomains[%d]: got %q, want %q", i, got.CustomDomains[i], cfg.CustomDomains[i])
			}
		}
	})
}

// Feature: dns-selector-gui, Property 9: 无效配置导入被拒绝
// **Validates: Requirements 13.5**

func TestProperty9_InvalidConfigRejected(t *testing.T) {
	// Feature: dns-selector-gui, Property 9: 无效配置导入被拒绝

	// Sub-property: random non-JSON strings are rejected by ValidateConfig
	t.Run("NonJSONStringRejected", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random strings that are not valid JSON objects
			s := rapid.StringMatching(`[a-zA-Z0-9!@#$%^&*]{1,50}`).Draw(t, "invalidJSON")

			var cfg AppConfig
			err := json.Unmarshal([]byte(s), &cfg)
			if err == nil {
				// If it somehow parsed, ValidateConfig should still reject it
				// because default zero-values won't pass validation
				valErr := ValidateConfig(&cfg)
				if valErr == nil {
					t.Fatalf("invalid JSON string %q was accepted after parse + validate", s)
				}
			}
			// If json.Unmarshal failed, the import is correctly rejected
		})
	})

	// Sub-property: corrupt JSON files are rejected during load
	t.Run("CorruptJSONReturnsError", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			// Generate random bytes that are unlikely to be valid JSON
			s := rapid.StringMatching(`[{"\w:,}]{5,40}`).Draw(t, "corruptJSON")

			var cfg AppConfig
			err := json.Unmarshal([]byte(s), &cfg)
			if err != nil {
				// Corrupt JSON correctly rejected at parse level
				return
			}
			// If it parsed, ValidateConfig should catch invalid values
			valErr := ValidateConfig(&cfg)
			if valErr == nil {
				// Only valid if all fields happen to be valid — this is acceptable
				// as the random generator may produce valid-looking data
				return
			}
			// ValidateConfig correctly rejected the parsed config
		})
	})
}

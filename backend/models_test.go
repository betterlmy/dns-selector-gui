package backend

import (
	"encoding/json"
	"testing"
)

func TestDefaultTestParams(t *testing.T) {
	p := DefaultTestParams()
	if p.Queries != 10 {
		t.Errorf("Queries: got %d, want 10", p.Queries)
	}
	if p.Warmup != 1 {
		t.Errorf("Warmup: got %d, want 1", p.Warmup)
	}
	if p.Concurrency != 20 {
		t.Errorf("Concurrency: got %d, want 20", p.Concurrency)
	}
	if p.Timeout != 2.0 {
		t.Errorf("Timeout: got %f, want 2.0", p.Timeout)
	}
}

func TestDefaultAppConfig(t *testing.T) {
	cfg := DefaultAppConfig()
	if cfg.CurrentPreset != "cn" {
		t.Errorf("CurrentPreset: got %q, want %q", cfg.CurrentPreset, "cn")
	}
	if len(cfg.CustomServers) != 0 {
		t.Errorf("CustomServers: got %d items, want 0", len(cfg.CustomServers))
	}
	if len(cfg.CustomDomains) != 0 {
		t.Errorf("CustomDomains: got %d items, want 0", len(cfg.CustomDomains))
	}
	if cfg.TestParams != DefaultTestParams() {
		t.Errorf("TestParams: got %+v, want %+v", cfg.TestParams, DefaultTestParams())
	}
}

func TestAppConfigJSONRoundTrip(t *testing.T) {
	cfg := DefaultAppConfig()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got AppConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.CurrentPreset != cfg.CurrentPreset {
		t.Errorf("CurrentPreset mismatch after round-trip")
	}
	if got.TestParams != cfg.TestParams {
		t.Errorf("TestParams mismatch after round-trip")
	}
}

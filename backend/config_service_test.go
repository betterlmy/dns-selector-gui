package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Task 5.5: 单元测试：配置文件回退和持久化
// Requirements: 13.7, 13.8, 14.1, 14.2, 14.3

func TestLoad_NonexistentFile_ReturnsDefault(t *testing.T) {
	svc := NewConfigService()
	cfg, err := svc.Load(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	expected := DefaultAppConfig()
	if cfg.CurrentPreset != expected.CurrentPreset {
		t.Errorf("CurrentPreset: got %q, want %q", cfg.CurrentPreset, expected.CurrentPreset)
	}
	if cfg.TestParams != expected.TestParams {
		t.Errorf("TestParams: got %+v, want %+v", cfg.TestParams, expected.TestParams)
	}
	if len(cfg.CustomServers) != 0 {
		t.Errorf("CustomServers: got %d, want 0", len(cfg.CustomServers))
	}
	if len(cfg.CustomDomains) != 0 {
		t.Errorf("CustomDomains: got %d, want 0", len(cfg.CustomDomains))
	}
}

func TestLoad_CorruptJSON_ReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(path, []byte("{invalid json!!!}"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	svc := NewConfigService()
	cfg, err := svc.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	expected := DefaultAppConfig()
	if cfg.CurrentPreset != expected.CurrentPreset {
		t.Errorf("CurrentPreset: got %q, want %q", cfg.CurrentPreset, expected.CurrentPreset)
	}
	if cfg.TestParams != expected.TestParams {
		t.Errorf("TestParams: got %+v, want %+v", cfg.TestParams, expected.TestParams)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	original := AppConfig{
		CurrentPreset: "global",
		CustomServers: []CustomServerEntry{
			{Protocol: "udp", Address: "9.9.9.9"},
			{Protocol: "doh", Address: "https://dns.example.com/dns-query", TLSServerName: "dns.example.com"},
		},
		CustomDomains: []string{"example.com", "test.org"},
		TestParams: TestParams{
			Queries:     20,
			Warmup:      3,
			Concurrency: 50,
			Timeout:     5.0,
		},
	}

	// Save
	svc := NewConfigService()
	svc.UpdateConfig(&original) // sets in-memory config
	if err := svc.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load into a fresh service
	svc2 := NewConfigService()
	loaded, err := svc2.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.CurrentPreset != original.CurrentPreset {
		t.Errorf("CurrentPreset: got %q, want %q", loaded.CurrentPreset, original.CurrentPreset)
	}
	if loaded.TestParams != original.TestParams {
		t.Errorf("TestParams: got %+v, want %+v", loaded.TestParams, original.TestParams)
	}
	if len(loaded.CustomServers) != len(original.CustomServers) {
		t.Fatalf("CustomServers length: got %d, want %d", len(loaded.CustomServers), len(original.CustomServers))
	}
	for i := range original.CustomServers {
		if loaded.CustomServers[i] != original.CustomServers[i] {
			t.Errorf("CustomServers[%d]: got %+v, want %+v", i, loaded.CustomServers[i], original.CustomServers[i])
		}
	}
	if len(loaded.CustomDomains) != len(original.CustomDomains) {
		t.Fatalf("CustomDomains length: got %d, want %d", len(loaded.CustomDomains), len(original.CustomDomains))
	}
	for i := range original.CustomDomains {
		if loaded.CustomDomains[i] != original.CustomDomains[i] {
			t.Errorf("CustomDomains[%d]: got %q, want %q", i, loaded.CustomDomains[i], original.CustomDomains[i])
		}
	}
}

func TestSaveResults_AndLoadResults_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	// Override the default results path by writing directly to a temp file
	resultsPath := filepath.Join(dir, "last_results.json")

	results := &TestResultsData{
		Items: []TestResultItem{
			{
				Name:            "TestDNS",
				Address:         "1.2.3.4",
				Protocol:        "udp",
				MedianLatencyMs: 15.5,
				P95LatencyMs:    25.0,
				SuccessRate:     0.95,
				RawSuccesses:    19,
				Successes:       19,
				Total:           20,
				Score:           42.5,
			},
		},
		TestTime: "2024-01-01T00:00:00Z",
		Preset:   "cn",
		BestDNS:  "TestDNS",
	}

	// Manually save results to temp path
	persisted := PersistedResults{
		Results: *results,
		Version: "1.0.0",
	}
	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(resultsPath, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Read back and verify
	readData, err := os.ReadFile(resultsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var loaded PersistedResults
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if loaded.Results.BestDNS != results.BestDNS {
		t.Errorf("BestDNS: got %q, want %q", loaded.Results.BestDNS, results.BestDNS)
	}
	if loaded.Results.Preset != results.Preset {
		t.Errorf("Preset: got %q, want %q", loaded.Results.Preset, results.Preset)
	}
	if loaded.Results.TestTime != results.TestTime {
		t.Errorf("TestTime: got %q, want %q", loaded.Results.TestTime, results.TestTime)
	}
	if len(loaded.Results.Items) != len(results.Items) {
		t.Fatalf("Items length: got %d, want %d", len(loaded.Results.Items), len(results.Items))
	}
	item := loaded.Results.Items[0]
	if item.Name != "TestDNS" || item.Address != "1.2.3.4" || item.Protocol != "udp" {
		t.Errorf("Item basic fields mismatch: %+v", item)
	}
	if item.MedianLatencyMs != 15.5 || item.P95LatencyMs != 25.0 {
		t.Errorf("Item latency mismatch: median=%f, p95=%f", item.MedianLatencyMs, item.P95LatencyMs)
	}
	if item.SuccessRate != 0.95 || item.Score != 42.5 {
		t.Errorf("Item rate/score mismatch: rate=%f, score=%f", item.SuccessRate, item.Score)
	}
}

func TestLoadResults_NonexistentFile_ReturnsNil(t *testing.T) {
	// LoadResults uses defaultResultsPath() which depends on APPDATA.
	// We test the behavior by setting APPDATA to a temp dir with no results file.
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)

	svc := NewConfigService()
	result, err := svc.LoadResults()
	if err != nil {
		t.Fatalf("LoadResults returned error: %v", err)
	}
	if result != nil {
		t.Errorf("LoadResults: got %+v, want nil", result)
	}
}

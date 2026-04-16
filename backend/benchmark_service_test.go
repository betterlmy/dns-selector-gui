package backend

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/betterlmy/dns-selector/selector"
)

// --- Task 6.1: BenchmarkService core logic tests ---

func TestNewBenchmarkService(t *testing.T) {
	bs := NewBenchmarkService()
	if bs == nil {
		t.Fatal("NewBenchmarkService returned nil")
	}
	if bs.IsRunning() {
		t.Error("new BenchmarkService should not be running")
	}
}

func TestBuildSelector_EmptyServers(t *testing.T) {
	bs := NewBenchmarkService()
	err := bs.BuildSelector(nil, []string{"example.com"}, DefaultTestParams())
	if err == nil {
		t.Error("expected error for empty servers")
	}
}

func TestBuildSelector_EmptyDomains(t *testing.T) {
	bs := NewBenchmarkService()
	servers := []selector.DNSServer{{Name: "test", Address: "8.8.8.8", Protocol: "udp"}}
	err := bs.BuildSelector(servers, nil, DefaultTestParams())
	if err == nil {
		t.Error("expected error for empty domains")
	}
}

func TestBuildSelector_Success(t *testing.T) {
	bs := NewBenchmarkService()
	servers := []selector.DNSServer{{Name: "Google", Address: "8.8.8.8", Protocol: "udp"}}
	domains := []string{"example.com"}
	err := bs.BuildSelector(servers, domains, DefaultTestParams())
	if err != nil {
		t.Fatalf("BuildSelector failed: %v", err)
	}
	if bs.selector == nil {
		t.Error("selector should be set after BuildSelector")
	}
}

func TestRunBenchmark_NoSelector(t *testing.T) {
	bs := NewBenchmarkService()
	_, err := bs.RunBenchmark(context.Background(), func() {})
	if err == nil {
		t.Error("expected error when selector not built")
	}
}

func TestRunBenchmark_ContextCancel(t *testing.T) {
	bs := NewBenchmarkService()
	servers := []selector.DNSServer{{Name: "Google", Address: "8.8.8.8", Protocol: "udp"}}
	domains := []string{"example.com"}
	params := TestParams{Queries: 5, Warmup: 0, Concurrency: 1, Timeout: 1.0}
	if err := bs.BuildSelector(servers, domains, params); err != nil {
		t.Fatalf("BuildSelector failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := bs.RunBenchmark(ctx, func() {})
	// With an already-cancelled context, the SDK should return an error or empty results
	// Either way, IsRunning should be false after return
	_ = err
	if bs.IsRunning() {
		t.Error("IsRunning should be false after RunBenchmark returns")
	}
}

func TestIsRunning_DefaultFalse(t *testing.T) {
	bs := NewBenchmarkService()
	if bs.IsRunning() {
		t.Error("IsRunning should be false by default")
	}
}

// --- Task 6.2: Score calculation and result sorting tests ---

func TestCalculateScore_AllTimeout(t *testing.T) {
	// median == 0 means all timed out
	score := CalculateScore(0, 0, 0, 0)
	if score != 0 {
		t.Errorf("expected 0 for all-timeout, got %f", score)
	}
}

func TestCalculateScore_ZeroSuccessRate(t *testing.T) {
	score := CalculateScore(0.05, 0.1, 0, 10)
	if score != 0 {
		t.Errorf("expected 0 for zero success rate, got %f", score)
	}
}

func TestCalculateScore_WithJitterPenalty(t *testing.T) {
	// validSamples >= 5: full formula
	median := 0.05 // 50ms
	p95 := 0.1     // 100ms
	successRate := 1.0
	validSamples := 10

	expected := (1.0 / median) * (successRate * successRate) * (median / p95)
	got := CalculateScore(median, p95, successRate, validSamples)

	if math.Abs(got-expected) > 1e-9 {
		t.Errorf("expected %f, got %f", expected, got)
	}
}

func TestCalculateScore_WithoutJitterPenalty(t *testing.T) {
	// validSamples < 5: omit jitter penalty
	median := 0.05
	p95 := 0.1
	successRate := 0.8
	validSamples := 3

	expected := (1.0 / median) * (successRate * successRate)
	got := CalculateScore(median, p95, successRate, validSamples)

	if math.Abs(got-expected) > 1e-9 {
		t.Errorf("expected %f, got %f", expected, got)
	}
}

func TestCalculateScore_PerfectServer(t *testing.T) {
	// Perfect: low latency, 100% success, no jitter
	median := 0.01
	p95 := 0.01
	successRate := 1.0
	validSamples := 100

	expected := (1.0 / median) * 1.0 * (median / p95) // = 100
	got := CalculateScore(median, p95, successRate, validSamples)

	if math.Abs(got-expected) > 1e-9 {
		t.Errorf("expected %f, got %f", expected, got)
	}
}

func TestSortResultsByScore_Empty(t *testing.T) {
	sorted, bestDNS := SortResultsByScore(nil)
	if len(sorted) != 0 {
		t.Errorf("expected empty slice, got %d items", len(sorted))
	}
	if bestDNS != "" {
		t.Errorf("expected empty bestDNS, got %q", bestDNS)
	}
}

func TestSortResultsByScore_SingleItem(t *testing.T) {
	items := []TestResultItem{{Name: "Server1", Score: 42.0}}
	sorted, bestDNS := SortResultsByScore(items)
	if len(sorted) != 1 {
		t.Fatalf("expected 1 item, got %d", len(sorted))
	}
	if bestDNS != "Server1" {
		t.Errorf("expected bestDNS=Server1, got %q", bestDNS)
	}
}

func TestSortResultsByScore_MultipleItems(t *testing.T) {
	items := []TestResultItem{
		{Name: "Low", Score: 10},
		{Name: "High", Score: 100},
		{Name: "Mid", Score: 50},
		{Name: "Zero", Score: 0},
	}
	sorted, bestDNS := SortResultsByScore(items)

	if bestDNS != "High" {
		t.Errorf("expected bestDNS=High, got %q", bestDNS)
	}

	for i := 1; i < len(sorted); i++ {
		if sorted[i].Score > sorted[i-1].Score {
			t.Errorf("not sorted descending at index %d: %f > %f", i, sorted[i].Score, sorted[i-1].Score)
		}
	}
}

func TestSortResultsByScore_DoesNotMutateOriginal(t *testing.T) {
	items := []TestResultItem{
		{Name: "B", Score: 10},
		{Name: "A", Score: 50},
	}
	SortResultsByScore(items)
	// Original should be unchanged
	if items[0].Name != "B" || items[1].Name != "A" {
		t.Error("SortResultsByScore mutated the original slice")
	}
}

func TestProcessResults_Basic(t *testing.T) {
	rawResults := []selector.BenchmarkResult{
		{
			Name:             "ServerA",
			Address:          "1.1.1.1",
			Protocol:         "udp",
			MedianTime:       50 * time.Millisecond,
			P95Time:          100 * time.Millisecond,
			SuccessRate:      1.0,
			Successes:        10,
			Total:            10,
			RawSuccesses:     10,
			AnswerMismatches: 0,
		},
		{
			Name:             "ServerB",
			Address:          "8.8.8.8",
			Protocol:         "udp",
			MedianTime:       200 * time.Millisecond,
			P95Time:          300 * time.Millisecond,
			SuccessRate:      0.8,
			Successes:        8,
			Total:            10,
			RawSuccesses:     9,
			AnswerMismatches: 1,
		},
	}

	result := ProcessResults(rawResults, "cn")

	if result == nil {
		t.Fatal("ProcessResults returned nil")
	}
	if result.Preset != "cn" {
		t.Errorf("expected preset=cn, got %q", result.Preset)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	// ServerA should be best (higher score)
	if result.BestDNS != "ServerA" {
		t.Errorf("expected BestDNS=ServerA, got %q", result.BestDNS)
	}
	// Items should be sorted descending by score
	if result.Items[0].Score < result.Items[1].Score {
		t.Error("items not sorted by score descending")
	}
	// Verify TestTime is set
	if result.TestTime == "" {
		t.Error("TestTime should be set")
	}
}

func TestProcessResults_AllTimeout(t *testing.T) {
	rawResults := []selector.BenchmarkResult{
		{
			Name:        "TimeoutServer",
			Address:     "10.0.0.1",
			Protocol:    "udp",
			MedianTime:  0,
			P95Time:     0,
			SuccessRate: 0,
			Successes:   0,
			Total:       10,
		},
	}

	result := ProcessResults(rawResults, "global")
	if result.Items[0].Score != 0 {
		t.Errorf("expected score=0 for timeout server, got %f", result.Items[0].Score)
	}
	if !result.Items[0].IsTimeout {
		t.Error("expected IsTimeout=true for server with 0 successes and 0 success rate")
	}
}

func TestProcessResults_Empty(t *testing.T) {
	result := ProcessResults(nil, "cn")
	if result == nil {
		t.Fatal("ProcessResults returned nil for empty input")
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
	if result.BestDNS != "" {
		t.Errorf("expected empty BestDNS, got %q", result.BestDNS)
	}
}

func TestProcessResults_LatencyConversion(t *testing.T) {
	rawResults := []selector.BenchmarkResult{
		{
			Name:        "Test",
			Address:     "1.2.3.4",
			Protocol:    "udp",
			MedianTime:  50 * time.Millisecond,
			P95Time:     100 * time.Millisecond,
			SuccessRate: 1.0,
			Successes:   10,
			Total:       10,
		},
	}

	result := ProcessResults(rawResults, "cn")
	item := result.Items[0]

	if math.Abs(item.MedianLatencyMs-50.0) > 0.01 {
		t.Errorf("expected MedianLatencyMs=50.0, got %f", item.MedianLatencyMs)
	}
	if math.Abs(item.P95LatencyMs-100.0) > 0.01 {
		t.Errorf("expected P95LatencyMs=100.0, got %f", item.P95LatencyMs)
	}
}

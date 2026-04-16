package backend

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/betterlmy/dns-selector/selector"
)

// BenchmarkService 封装 selector SDK，管理 DNS 测试的生命周期。
type BenchmarkService struct {
	selector *selector.Selector
	running  bool
	mu       sync.Mutex
}

// NewBenchmarkService 创建 BenchmarkService 实例。
func NewBenchmarkService() *BenchmarkService {
	return &BenchmarkService{}
}

// BuildSelector 根据服务器列表、域名列表和测试参数构建 selector.Selector 实例。
func (b *BenchmarkService) BuildSelector(servers []selector.DNSServer, domains []string, params TestParams) error {
	if len(servers) == 0 {
		return fmt.Errorf("服务器列表为空")
	}
	if len(domains) == 0 {
		return fmt.Errorf("域名列表为空")
	}

	timeout := time.Duration(params.Timeout * float64(time.Second))

	b.mu.Lock()
	defer b.mu.Unlock()
	b.selector = selector.NewSelector(servers, domains, params.Queries, params.Warmup, params.Concurrency, timeout)
	return nil
}

// RunBenchmark 执行 DNS 测试。阻塞直到完成或 context 被取消。
// progressCb 在每次单个查询完成时被调用。
func (b *BenchmarkService) RunBenchmark(ctx context.Context, progressCb func()) ([]selector.BenchmarkResult, error) {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil, fmt.Errorf("测试正在运行中")
	}
	if b.selector == nil {
		b.mu.Unlock()
		return nil, fmt.Errorf("selector 未构建，请先调用 BuildSelector")
	}
	b.running = true
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		b.running = false
		b.mu.Unlock()
	}()

	results, err := b.selector.Benchmark(ctx, progressCb)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// IsRunning 返回当前是否有测试正在执行。
func (b *BenchmarkService) IsRunning() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.running
}

// CalculateScore 计算 DNS 服务器的综合评分。
// 公式：(1/中位延迟秒) × (成功率²) × (中位延迟/P95延迟)
// 当有效样本数 < 5 时，省略抖动惩罚因子（中位延迟/P95延迟）。
// 当所有查询均超时（中位延迟 == 0 或成功率 == 0）时，Score 为 0。
func CalculateScore(medianSeconds, p95Seconds, successRate float64, validSamples int) float64 {
	if medianSeconds <= 0 || successRate <= 0 {
		return 0
	}

	score := (1.0 / medianSeconds) * (successRate * successRate)

	// 仅在有足够样本时应用抖动惩罚
	if validSamples >= 5 && p95Seconds > 0 {
		score *= medianSeconds / p95Seconds
	}

	return score
}

// SortResultsByScore 按 Score 降序排列结果，并返回排序后的切片和最高分服务器名称。
func SortResultsByScore(items []TestResultItem) (sorted []TestResultItem, bestDNS string) {
	sorted = make([]TestResultItem, len(items))
	copy(sorted, items)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	if len(sorted) > 0 {
		bestDNS = sorted[0].Name
	}
	return sorted, bestDNS
}

// ProcessResults 将原始测试结果转换为按 Score 降序排列的 TestResultsData。
// BestDNS 设置为最高分服务器。
func ProcessResults(rawResults []selector.BenchmarkResult, preset string) *TestResultsData {
	items := make([]TestResultItem, 0, len(rawResults))

	for _, r := range rawResults {
		medianMs := float64(r.MedianTime.Microseconds()) / 1000.0
		p95Ms := float64(r.P95Time.Microseconds()) / 1000.0
		medianSeconds := r.MedianTime.Seconds()
		p95Seconds := r.P95Time.Seconds()

		// 判断是否全部超时：成功率为 0 且成功数为 0
		isTimeout := r.SuccessRate == 0 && r.Successes == 0

		// 有效样本数 = 成功查询数
		validSamples := r.Successes

		score := CalculateScore(medianSeconds, p95Seconds, r.SuccessRate, validSamples)

		item := TestResultItem{
			Name:             r.Name,
			Address:          r.Address,
			Protocol:         r.Protocol,
			MedianLatencyMs:  medianMs,
			P95LatencyMs:     p95Ms,
			SuccessRate:      r.SuccessRate,
			RawSuccesses:     r.RawSuccesses,
			Successes:        r.Successes,
			Total:            r.Total,
			AnswerMismatches: r.AnswerMismatches,
			Score:            score,
			IsTimeout:        isTimeout,
		}
		items = append(items, item)
	}

	sortedItems, bestDNS := SortResultsByScore(items)

	return &TestResultsData{
		Items:    sortedItems,
		TestTime: time.Now().Format(time.RFC3339),
		Preset:   preset,
		BestDNS:  bestDNS,
	}
}

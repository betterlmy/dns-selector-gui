package backend

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/betterlmy/dns-selector/selector"
)

// 任务 15.3: 集成测试 —— 完整 Benchmark 流程
// 验证需求: 8.1, 8.4, 8.6, 14.1

func TestIntegration_FullBenchmarkFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式下跳过集成测试")
	}

	// 步骤 1: 创建 BenchmarkService
	bs := NewBenchmarkService()

	// 步骤 2: 使用最小配置构建 selector（加快测试速度）
	servers := []selector.DNSServer{
		{Name: "Google", Address: "8.8.8.8", Protocol: "udp"},
	}
	domains := []string{"example.com"}
	params := TestParams{Queries: 2, Warmup: 1, Concurrency: 1, Timeout: 3.0}

	err := bs.BuildSelector(servers, domains, params)
	if err != nil {
		t.Fatalf("BuildSelector 失败: %v", err)
	}

	// 步骤 3: 执行测试并验证进度回调
	var progressCount int64
	progressCb := func() {
		atomic.AddInt64(&progressCount, 1)
	}

	results, err := bs.RunBenchmark(context.Background(), progressCb)
	if err != nil {
		t.Fatalf("RunBenchmark 失败: %v", err)
	}

	// 步骤 4: 验证返回了测试结果
	if len(results) == 0 {
		t.Fatal("期望至少 1 个测试结果")
	}

	// 步骤 5: 验证进度回调被调用
	if atomic.LoadInt64(&progressCount) == 0 {
		t.Error("进度回调从未被调用")
	}

	// 步骤 6: 处理原始结果
	processed := ProcessResults(results, "cn")
	if processed == nil {
		t.Fatal("ProcessResults 返回 nil")
	}

	// 步骤 7: 验证处理后的结果结构
	if len(processed.Items) != 1 {
		t.Errorf("期望 1 个结果项，实际 %d 个", len(processed.Items))
	}
	if processed.BestDNS == "" {
		t.Error("BestDNS 应该被设置")
	}
	if processed.TestTime == "" {
		t.Error("TestTime 应该被设置")
	}

	// 步骤 8: 测试持久化 —— 保存到临时目录，再加载回来验证
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)

	cs := NewConfigService()
	err = cs.SaveResults(processed)
	if err != nil {
		t.Fatalf("SaveResults 失败: %v", err)
	}

	loaded, err := cs.LoadResults()
	if err != nil {
		t.Fatalf("LoadResults 失败: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadResults 返回 nil")
	}
	if loaded.BestDNS != processed.BestDNS {
		t.Errorf("BestDNS 不匹配: 实际 %q, 期望 %q", loaded.BestDNS, processed.BestDNS)
	}
	if len(loaded.Items) != len(processed.Items) {
		t.Errorf("Items 数量不匹配: 实际 %d, 期望 %d", len(loaded.Items), len(processed.Items))
	}
}

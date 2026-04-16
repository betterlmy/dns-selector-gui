package backend

import (
	"fmt"
	"math"
	"testing"

	"pgregory.net/rapid"
)

// Feature: dns-selector-gui, Property 6: Score 计算公式正确性
// **Validates: Requirements 8.7, 8.8, 8.9**

func TestProperty6_ScoreCalculationCorrectness(t *testing.T) {
	// Sub-property: with >= 5 valid samples, full formula applies
	t.Run("FullFormula", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			median := rapid.Float64Range(0.001, 10.0).Draw(t, "median")
			p95 := rapid.Float64Range(median, 20.0).Draw(t, "p95") // p95 >= median
			successRate := rapid.Float64Range(0.01, 1.0).Draw(t, "successRate")
			validSamples := rapid.IntRange(5, 1000).Draw(t, "validSamples")

			got := CalculateScore(median, p95, successRate, validSamples)
			expected := (1.0 / median) * (successRate * successRate) * (median / p95)

			if math.Abs(got-expected) > 1e-9 {
				t.Fatalf("got %f, want %f", got, expected)
			}
		})
	})

	// Sub-property: with < 5 valid samples, jitter penalty omitted
	t.Run("NoJitterPenalty", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			median := rapid.Float64Range(0.001, 10.0).Draw(t, "median")
			p95 := rapid.Float64Range(median, 20.0).Draw(t, "p95")
			successRate := rapid.Float64Range(0.01, 1.0).Draw(t, "successRate")
			validSamples := rapid.IntRange(0, 4).Draw(t, "validSamples")

			got := CalculateScore(median, p95, successRate, validSamples)
			expected := (1.0 / median) * (successRate * successRate)

			if math.Abs(got-expected) > 1e-9 {
				t.Fatalf("got %f, want %f", got, expected)
			}
		})
	})

	// Sub-property: all-timeout returns 0
	t.Run("AllTimeoutReturnsZero", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			p95 := rapid.Float64Range(0, 10.0).Draw(t, "p95")
			validSamples := rapid.IntRange(0, 100).Draw(t, "validSamples")

			got := CalculateScore(0, p95, 0, validSamples)
			if got != 0 {
				t.Fatalf("expected 0 for all-timeout, got %f", got)
			}
		})
	})
}

// Feature: dns-selector-gui, Property 7: 测试结果按 Score 降序排列且推荐最高分
// **Validates: Requirements 10.2, 10.4**

func TestProperty7_ResultsSortedDescendingAndBestDNS(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		count := rapid.IntRange(1, 50).Draw(t, "count")
		items := make([]TestResultItem, count)
		for i := range items {
			items[i] = TestResultItem{
				Name:  fmt.Sprintf("Server%d", i),
				Score: rapid.Float64Range(0, 1000).Draw(t, fmt.Sprintf("score%d", i)),
			}
		}

		sorted, bestDNS := SortResultsByScore(items)

		// Verify descending order
		for i := 1; i < len(sorted); i++ {
			if sorted[i].Score > sorted[i-1].Score {
				t.Fatalf("not descending at index %d: %f > %f", i, sorted[i].Score, sorted[i-1].Score)
			}
		}

		// Verify bestDNS is the highest scorer
		if bestDNS != sorted[0].Name {
			t.Fatalf("bestDNS=%q, but sorted[0].Name=%q", bestDNS, sorted[0].Name)
		}
	})
}

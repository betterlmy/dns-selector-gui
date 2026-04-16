package backend

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

// Feature: dns-selector-gui, Property 2: 服务器地址格式验证
// **Validates: Requirements 5.3, 5.4, 5.5, 5.6**

func TestProperty2_ServerAddressFormatValidation(t *testing.T) {
	// Feature: dns-selector-gui, Property 2: 服务器地址格式验证

	// Sub-property: valid IPv4 addresses are always accepted for UDP
	t.Run("UDP_ValidIPv4", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			a := rapid.IntRange(0, 255).Draw(t, "a")
			b := rapid.IntRange(0, 255).Draw(t, "b")
			c := rapid.IntRange(0, 255).Draw(t, "c")
			d := rapid.IntRange(0, 255).Draw(t, "d")
			addr := fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
			err := ValidateServerAddress("udp", addr)
			if err != nil {
				t.Fatalf("valid IPv4 %q rejected: %v", addr, err)
			}
		})
	})

	// Sub-property: valid domains are always accepted for DoT
	t.Run("DoT_ValidDomain", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			label := rapid.StringMatching(`[a-z][a-z0-9]{1,10}`).Draw(t, "label")
			tld := rapid.SampledFrom([]string{"com", "net", "org", "io"}).Draw(t, "tld")
			domain := label + "." + tld
			err := ValidateServerAddress("dot", domain)
			if err != nil {
				t.Fatalf("valid domain %q rejected: %v", domain, err)
			}
		})
	})

	// Sub-property: valid HTTPS URLs are always accepted for DoH
	t.Run("DoH_ValidHTTPS", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			label := rapid.StringMatching(`[a-z][a-z0-9]{1,10}`).Draw(t, "label")
			tld := rapid.SampledFrom([]string{"com", "net", "org"}).Draw(t, "tld")
			url := "https://" + label + "." + tld + "/dns-query"
			err := ValidateServerAddress("doh", url)
			if err != nil {
				t.Fatalf("valid HTTPS URL %q rejected: %v", url, err)
			}
		})
	})

	// Sub-property: non-IPv4 strings are rejected for UDP
	t.Run("UDP_RejectsNonIPv4", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := rapid.StringMatching(`[a-z]{3,20}`).Draw(t, "nonIP")
			err := ValidateServerAddress("udp", s)
			if err == nil {
				t.Fatalf("non-IPv4 string %q accepted for UDP", s)
			}
		})
	})
}

// Feature: dns-selector-gui, Property 5: 测试参数验证
// **Validates: Requirements 7.3**

func TestProperty5_TestParamsValidation(t *testing.T) {
	// Feature: dns-selector-gui, Property 5: 测试参数验证

	// Sub-property: all-positive params are accepted
	t.Run("AllPositiveAccepted", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			params := TestParams{
				Queries:     rapid.IntRange(1, 10000).Draw(t, "queries"),
				Warmup:      rapid.IntRange(1, 100).Draw(t, "warmup"),
				Concurrency: rapid.IntRange(1, 1000).Draw(t, "concurrency"),
				Timeout:     rapid.Float64Range(0.001, 60.0).Draw(t, "timeout"),
			}
			err := ValidateTestParams(params)
			if err != nil {
				t.Fatalf("valid params %+v rejected: %v", params, err)
			}
		})
	})

	// Sub-property: zero or negative queries are rejected
	t.Run("NonPositiveQueriesRejected", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			params := DefaultTestParams()
			params.Queries = rapid.IntRange(-1000, 0).Draw(t, "queries")
			err := ValidateTestParams(params)
			if err == nil {
				t.Fatalf("non-positive queries %d accepted", params.Queries)
			}
		})
	})

	// Sub-property: zero or negative timeout is rejected
	t.Run("NonPositiveTimeoutRejected", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			params := DefaultTestParams()
			params.Timeout = rapid.Float64Range(-100.0, 0.0).Draw(t, "timeout")
			err := ValidateTestParams(params)
			if err == nil {
				t.Fatalf("non-positive timeout %f accepted", params.Timeout)
			}
		})
	})
}

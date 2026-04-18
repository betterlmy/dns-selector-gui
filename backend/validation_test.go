package backend

import (
	"testing"
)

func TestValidateServerAddress_UDP_Valid(t *testing.T) {
	validAddresses := []string{
		"8.8.8.8",
		"1.1.1.1",
		"223.5.5.5",
		"192.168.1.1",
		"0.0.0.0",
		"255.255.255.255",
	}
	for _, addr := range validAddresses {
		if err := ValidateServerAddress("udp", addr); err != nil {
			t.Errorf("ValidateServerAddress(udp, %q) returned error: %v", addr, err)
		}
	}
}

func TestValidateServerAddress_UDP_Invalid(t *testing.T) {
	invalidAddresses := []string{
		"",
		"not-an-ip",
		"256.1.1.1",
		"1.2.3",
		"::1",
		"2001:db8::1",
		"dns.google",
		"https://8.8.8.8",
	}
	for _, addr := range invalidAddresses {
		if err := ValidateServerAddress("udp", addr); err == nil {
			t.Errorf("ValidateServerAddress(udp, %q) expected error, got nil", addr)
		}
	}
}

func TestValidateServerAddress_DoT_ValidDomain(t *testing.T) {
	validDomains := []string{
		"dns.google",
		"dot.pub",
		"one.one.one.one",
		"dns.alidns.com",
		"example.com",
		"a.b.c.d",
	}
	for _, addr := range validDomains {
		if err := ValidateServerAddress("dot", addr); err != nil {
			t.Errorf("ValidateServerAddress(dot, %q) returned error: %v", addr, err)
		}
	}
}

func TestValidateServerAddress_DoT_ValidIPAtTLS(t *testing.T) {
	validAddresses := []string{
		"8.8.8.8@dns.google",
		"1.1.1.1@one.one.one.one",
		"223.5.5.5@dns.alidns.com",
	}
	for _, addr := range validAddresses {
		if err := ValidateServerAddress("dot", addr); err != nil {
			t.Errorf("ValidateServerAddress(dot, %q) returned error: %v", addr, err)
		}
	}
}

func TestValidateServerAddress_DoT_Invalid(t *testing.T) {
	invalidAddresses := []string{
		"",
		"not valid domain!",
		"8.8.8.8@",
		"notanip@dns.google",
		"-invalid.com",
	}
	for _, addr := range invalidAddresses {
		if err := ValidateServerAddress("dot", addr); err == nil {
			t.Errorf("ValidateServerAddress(dot, %q) expected error, got nil", addr)
		}
	}
}

func TestValidateServerAddress_DoH_ValidURL(t *testing.T) {
	validURLs := []string{
		"https://dns.google/dns-query",
		"https://cloudflare-dns.com/dns-query",
		"https://doh.pub/dns-query",
		"https://1.1.1.1/dns-query",
	}
	for _, addr := range validURLs {
		if err := ValidateServerAddress("doh", addr); err != nil {
			t.Errorf("ValidateServerAddress(doh, %q) returned error: %v", addr, err)
		}
	}
}

func TestValidateServerAddress_DoH_ValidURLAtTLS(t *testing.T) {
	validAddresses := []string{
		"https://1.1.1.1/dns-query@cloudflare-dns.com",
		"https://8.8.8.8/dns-query@dns.google",
	}
	for _, addr := range validAddresses {
		if err := ValidateServerAddress("doh", addr); err != nil {
			t.Errorf("ValidateServerAddress(doh, %q) returned error: %v", addr, err)
		}
	}
}

func TestValidateServerAddress_DoH_Invalid(t *testing.T) {
	invalidAddresses := []string{
		"",
		"http://dns.google/dns-query",
		"dns.google/dns-query",
		"https://@tls",
		"ftp://dns.google/dns-query",
		"https://1.1.1.1/dns-query@",
	}
	for _, addr := range invalidAddresses {
		if err := ValidateServerAddress("doh", addr); err == nil {
			t.Errorf("ValidateServerAddress(doh, %q) expected error, got nil", addr)
		}
	}
}

func TestValidateServerAddress_UnsupportedProtocol(t *testing.T) {
	if err := ValidateServerAddress("tcp", "8.8.8.8"); err == nil {
		t.Error("expected error for unsupported protocol, got nil")
	}
}

func TestValidateServerAddress_CaseInsensitiveProtocol(t *testing.T) {
	cases := []struct {
		protocol string
		address  string
	}{
		{"UDP", "8.8.8.8"},
		{"Udp", "1.1.1.1"},
		{"DOT", "dns.google"},
		{"DoT", "dns.google"},
		{"DOH", "https://dns.google/dns-query"},
		{"DoH", "https://dns.google/dns-query"},
	}
	for _, tc := range cases {
		if err := ValidateServerAddress(tc.protocol, tc.address); err != nil {
			t.Errorf("ValidateServerAddress(%q, %q) returned error: %v", tc.protocol, tc.address, err)
		}
	}
}

func TestValidateTestParams_Valid(t *testing.T) {
	if err := ValidateTestParams(DefaultTestParams()); err != nil {
		t.Errorf("ValidateTestParams(DefaultTestParams()) returned error: %v", err)
	}
}

func TestValidateTestParams_InvalidQueries(t *testing.T) {
	p := DefaultTestParams()
	p.Queries = 0
	if err := ValidateTestParams(p); err == nil {
		t.Error("expected error for queries=0, got nil")
	}
}

func TestValidateTestParams_InvalidWarmup(t *testing.T) {
	p := DefaultTestParams()
	p.Warmup = -1
	if err := ValidateTestParams(p); err == nil {
		t.Error("expected error for warmup=-1, got nil")
	}
}

func TestValidateTestParams_InvalidConcurrency(t *testing.T) {
	p := DefaultTestParams()
	p.Concurrency = 0
	if err := ValidateTestParams(p); err == nil {
		t.Error("expected error for concurrency=0, got nil")
	}
}

func TestValidateTestParams_InvalidTimeoutZero(t *testing.T) {
	p := DefaultTestParams()
	p.Timeout = 0
	if err := ValidateTestParams(p); err == nil {
		t.Error("expected error for timeout=0, got nil")
	}
}

func TestValidateTestParams_InvalidTimeoutNegative(t *testing.T) {
	p := DefaultTestParams()
	p.Timeout = -1.0
	if err := ValidateTestParams(p); err == nil {
		t.Error("expected error for timeout=-1.0, got nil")
	}
}

func TestValidateDomain_Valid(t *testing.T) {
	validDomains := []string{
		"example.com",
		"sub.domain.com",
		"a.b.c",
	}
	for _, d := range validDomains {
		if err := ValidateDomain(d); err != nil {
			t.Errorf("ValidateDomain(%q) returned error: %v", d, err)
		}
	}
}

func TestValidateDomain_Invalid(t *testing.T) {
	invalidDomains := []string{
		"",
		"-invalid.com",
		"invalid-.com",
		"has space.com",
		"has@symbol.com",
	}
	for _, d := range invalidDomains {
		if err := ValidateDomain(d); err == nil {
			t.Errorf("ValidateDomain(%q) expected error, got nil", d)
		}
	}
}

func TestValidateBootstrapIP_RejectsIPv6(t *testing.T) {
	err := ValidateBootstrapIP("2001:4860:4860::8888")
	if err == nil {
		t.Fatal("expected IPv6 bootstrap IP to be rejected")
	}
}

func TestValidateServerEntry_RejectsInvalidTLSServerName(t *testing.T) {
	err := ValidateServerEntry(AddServerRequest{
		Protocol:      "dot",
		Address:       "dns.google",
		TLSServerName: "bad host name",
	})
	if err == nil {
		t.Fatal("expected invalid TLS server name to be rejected")
	}
}

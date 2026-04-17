//go:build linux

package backend

import "testing"

// Compile-time check: linuxPlatform must satisfy PlatformProvider.
var _ PlatformProvider = &linuxPlatform{}

func TestLinuxPlatform_ImplementsInterface(t *testing.T) {
	// This test verifies that linuxPlatform implements PlatformProvider.
	// The compile-time check above (var _ PlatformProvider = &linuxPlatform{})
	// ensures the interface is satisfied. If any method is missing or has a
	// wrong signature, this file will fail to compile.
	p := NewPlatformProvider()
	if p == nil {
		t.Fatal("NewPlatformProvider() returned nil")
	}
}

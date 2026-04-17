//go:build windows

package backend

import "testing"

// Compile-time check: windowsPlatform must satisfy PlatformProvider.
var _ PlatformProvider = &windowsPlatform{}

func TestWindowsPlatform_ImplementsInterface(t *testing.T) {
	// This test verifies that windowsPlatform implements PlatformProvider.
	// The compile-time check above (var _ PlatformProvider = &windowsPlatform{})
	// ensures the interface is satisfied. If any method is missing or has a
	// wrong signature, this file will fail to compile.
	p := NewPlatformProvider()
	if p == nil {
		t.Fatal("NewPlatformProvider() returned nil")
	}
}

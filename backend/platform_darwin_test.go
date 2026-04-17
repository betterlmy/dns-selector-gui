//go:build darwin

package backend

import "testing"

// Compile-time check: darwinPlatform must satisfy PlatformProvider.
var _ PlatformProvider = &darwinPlatform{}

func TestDarwinPlatform_ImplementsInterface(t *testing.T) {
	// This test verifies that darwinPlatform implements PlatformProvider.
	// The compile-time check above (var _ PlatformProvider = &darwinPlatform{})
	// ensures the interface is satisfied. If any method is missing or has a
	// wrong signature, this file will fail to compile.
	p := NewPlatformProvider()
	if p == nil {
		t.Fatal("NewPlatformProvider() returned nil")
	}
}

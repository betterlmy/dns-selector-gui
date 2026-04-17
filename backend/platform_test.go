package backend

import (
	"errors"
	"testing"
)

// mockPlatform implements PlatformProvider with configurable return values.
type mockPlatform struct {
	adapters    []NetworkAdapterInfo
	adaptersErr error
	setDNSErr   error
	resetErr    error
	isAdmin     bool
	theme       string
}

// Compile-time check: mockPlatform must satisfy PlatformProvider.
var _ PlatformProvider = &mockPlatform{}

func (m *mockPlatform) GetAdapters() ([]NetworkAdapterInfo, error) {
	return m.adapters, m.adaptersErr
}

func (m *mockPlatform) SetDNS(adapterName, primaryDNS, secondaryDNS string) error {
	return m.setDNSErr
}

func (m *mockPlatform) ResetToAuto(adapterName string) error {
	return m.resetErr
}

func (m *mockPlatform) CheckAdmin() bool {
	return m.isAdmin
}

func (m *mockPlatform) GetSystemTheme() string {
	return m.theme
}

func TestMockPlatform_ImplementsInterface(t *testing.T) {
	mock := &mockPlatform{
		adapters: []NetworkAdapterInfo{
			{Name: "eth0", InterfaceIdx: 1, Status: "up", IPAddresses: []string{"192.168.1.10"}, CurrentDNS: []string{"8.8.8.8"}},
		},
		adaptersErr: nil,
		setDNSErr:   nil,
		resetErr:    nil,
		isAdmin:     true,
		theme:       "dark",
	}

	// GetAdapters
	adapters, err := mock.GetAdapters()
	if err != nil {
		t.Fatalf("GetAdapters returned unexpected error: %v", err)
	}
	if len(adapters) != 1 {
		t.Fatalf("GetAdapters: got %d adapters, want 1", len(adapters))
	}
	if adapters[0].Name != "eth0" {
		t.Errorf("GetAdapters: adapter name = %q, want %q", adapters[0].Name, "eth0")
	}

	// SetDNS
	if err := mock.SetDNS("eth0", "8.8.8.8", "8.8.4.4"); err != nil {
		t.Errorf("SetDNS returned unexpected error: %v", err)
	}

	// ResetToAuto
	if err := mock.ResetToAuto("eth0"); err != nil {
		t.Errorf("ResetToAuto returned unexpected error: %v", err)
	}

	// CheckAdmin
	if !mock.CheckAdmin() {
		t.Error("CheckAdmin: got false, want true")
	}

	// GetSystemTheme
	if theme := mock.GetSystemTheme(); theme != "dark" {
		t.Errorf("GetSystemTheme: got %q, want %q", theme, "dark")
	}
}

func TestMockPlatform_ErrorPropagation(t *testing.T) {
	errAdapters := errors.New("adapter failure")
	errSetDNS := errors.New("set dns failure")
	errReset := errors.New("reset failure")

	mock := &mockPlatform{
		adapters:    nil,
		adaptersErr: errAdapters,
		setDNSErr:   errSetDNS,
		resetErr:    errReset,
		isAdmin:     false,
		theme:       "light",
	}

	// GetAdapters error
	_, err := mock.GetAdapters()
	if !errors.Is(err, errAdapters) {
		t.Errorf("GetAdapters error: got %v, want %v", err, errAdapters)
	}

	// SetDNS error
	if err := mock.SetDNS("eth0", "1.1.1.1", ""); !errors.Is(err, errSetDNS) {
		t.Errorf("SetDNS error: got %v, want %v", err, errSetDNS)
	}

	// ResetToAuto error
	if err := mock.ResetToAuto("eth0"); !errors.Is(err, errReset) {
		t.Errorf("ResetToAuto error: got %v, want %v", err, errReset)
	}

	// CheckAdmin returns false
	if mock.CheckAdmin() {
		t.Error("CheckAdmin: got true, want false")
	}

	// GetSystemTheme returns light
	if theme := mock.GetSystemTheme(); theme != "light" {
		t.Errorf("GetSystemTheme: got %q, want %q", theme, "light")
	}
}

// --- DNSConfigService delegation tests ---

func TestDNSConfigService_GetAdapters_Delegates(t *testing.T) {
	expected := []NetworkAdapterInfo{
		{Name: "Wi-Fi", InterfaceIdx: 2, Status: "up", IPAddresses: []string{"10.0.0.5"}, CurrentDNS: []string{"1.1.1.1"}},
		{Name: "Ethernet", InterfaceIdx: 3, Status: "up", IPAddresses: []string{"192.168.1.100"}, CurrentDNS: []string{"8.8.8.8", "8.8.4.4"}},
	}
	mock := &mockPlatform{adapters: expected}
	svc := NewDNSConfigService(mock)

	adapters, err := svc.GetAdapters()
	if err != nil {
		t.Fatalf("GetAdapters returned unexpected error: %v", err)
	}
	if len(adapters) != len(expected) {
		t.Fatalf("GetAdapters: got %d adapters, want %d", len(adapters), len(expected))
	}
	for i, a := range adapters {
		if a.Name != expected[i].Name {
			t.Errorf("adapter[%d].Name = %q, want %q", i, a.Name, expected[i].Name)
		}
	}
}

func TestDNSConfigService_SetDNS_Delegates(t *testing.T) {
	mock := &mockPlatform{}
	svc := NewDNSConfigService(mock)

	if err := svc.SetDNS("Wi-Fi", "1.1.1.1", "1.0.0.1"); err != nil {
		t.Errorf("SetDNS returned unexpected error: %v", err)
	}
}

func TestDNSConfigService_ResetToAuto_Delegates(t *testing.T) {
	mock := &mockPlatform{}
	svc := NewDNSConfigService(mock)

	if err := svc.ResetToAuto("Wi-Fi"); err != nil {
		t.Errorf("ResetToAuto returned unexpected error: %v", err)
	}
}

func TestDNSConfigService_CheckAdmin_Delegates(t *testing.T) {
	mock := &mockPlatform{isAdmin: true}
	svc := NewDNSConfigService(mock)

	if !svc.CheckAdmin() {
		t.Error("CheckAdmin: got false, want true")
	}

	mock.isAdmin = false
	if svc.CheckAdmin() {
		t.Error("CheckAdmin: got true, want false")
	}
}

// --- DNSConfigService error propagation tests ---

func TestDNSConfigService_GetAdapters_PropagatesError(t *testing.T) {
	expectedErr := errors.New("network unavailable")
	mock := &mockPlatform{adaptersErr: expectedErr}
	svc := NewDNSConfigService(mock)

	_, err := svc.GetAdapters()
	if !errors.Is(err, expectedErr) {
		t.Errorf("GetAdapters error: got %v, want %v", err, expectedErr)
	}
}

func TestDNSConfigService_SetDNS_PropagatesError(t *testing.T) {
	expectedErr := errors.New("permission denied")
	mock := &mockPlatform{setDNSErr: expectedErr}
	svc := NewDNSConfigService(mock)

	err := svc.SetDNS("eth0", "8.8.8.8", "")
	if !errors.Is(err, expectedErr) {
		t.Errorf("SetDNS error: got %v, want %v", err, expectedErr)
	}
}

func TestDNSConfigService_ResetToAuto_PropagatesError(t *testing.T) {
	expectedErr := errors.New("reset failed")
	mock := &mockPlatform{resetErr: expectedErr}
	svc := NewDNSConfigService(mock)

	err := svc.ResetToAuto("eth0")
	if !errors.Is(err, expectedErr) {
		t.Errorf("ResetToAuto error: got %v, want %v", err, expectedErr)
	}
}

// --- AppService delegation tests ---

func TestAppService_GetSystemTheme_Delegates(t *testing.T) {
	mock := &mockPlatform{theme: "dark"}
	app := &AppService{platform: mock}

	if theme := app.GetSystemTheme(); theme != "dark" {
		t.Errorf("GetSystemTheme: got %q, want %q", theme, "dark")
	}

	mock.theme = "light"
	if theme := app.GetSystemTheme(); theme != "light" {
		t.Errorf("GetSystemTheme: got %q, want %q", theme, "light")
	}
}

func TestAppService_IsAdmin_Delegates(t *testing.T) {
	mock := &mockPlatform{isAdmin: true}
	app := &AppService{platform: mock}

	if !app.IsAdmin() {
		t.Error("IsAdmin: got false, want true")
	}

	mock.isAdmin = false
	if app.IsAdmin() {
		t.Error("IsAdmin: got true, want false")
	}
}

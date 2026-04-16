package main

import (
	"fmt"
	"unsafe"
)

type dnsInterfaceSettings struct {
	Version             uint32
	_                   uint32
	Flags               uint64
	Domain              *uint16
	NameServer          *uint16
	SearchList          *uint16
	RegistrationEnabled uint32
	RegisterAdapterName uint32
	EnableLLMNR         uint32
	QueryAdapterName    uint32
	ProfileNameServer   *uint16
}

func main() {
	var s dnsInterfaceSettings
	fmt.Printf("Size:                       %d\n", unsafe.Sizeof(s))
	fmt.Printf("Version offset:             %d\n", unsafe.Offsetof(s.Version))
	fmt.Printf("Flags offset:               %d\n", unsafe.Offsetof(s.Flags))
	fmt.Printf("Domain offset:              %d\n", unsafe.Offsetof(s.Domain))
	fmt.Printf("NameServer offset:          %d\n", unsafe.Offsetof(s.NameServer))
	fmt.Printf("SearchList offset:          %d\n", unsafe.Offsetof(s.SearchList))
	fmt.Printf("RegistrationEnabled offset: %d\n", unsafe.Offsetof(s.RegistrationEnabled))
	fmt.Printf("RegisterAdapterName offset: %d\n", unsafe.Offsetof(s.RegisterAdapterName))
	fmt.Printf("EnableLLMNR offset:         %d\n", unsafe.Offsetof(s.EnableLLMNR))
	fmt.Printf("QueryAdapterName offset:    %d\n", unsafe.Offsetof(s.QueryAdapterName))
	fmt.Printf("ProfileNameServer offset:   %d\n", unsafe.Offsetof(s.ProfileNameServer))
	// 期望 C 64-bit 布局:
	// Version(0,4) pad(4,4) Flags(8,8) Domain(16,8) NameServer(24,8)
	// SearchList(32,8) RegistrationEnabled(40,4) RegisterAdapterName(44,4)
	// EnableLLMNR(48,4) QueryAdapterName(52,4) ProfileNameServer(56,8)
	// Total: 64
}

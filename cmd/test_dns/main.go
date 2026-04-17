//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	dnsInterfaceSettingsVersion1 = 1
	dnsSettingNameserver         = 0x0002
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
	// 1. 列出所有适配器 GUID
	const netKey = `SYSTEM\CurrentControlSet\Control\Network\{4D36E972-E325-11CE-BFC1-08002BE10318}`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, netKey, registry.READ)
	if err != nil {
		fmt.Println("打开注册表失败:", err)
		return
	}
	defer k.Close()

	guids, _ := k.ReadSubKeyNames(-1)
	fmt.Println("=== 网络适配器列表 ===")
	for _, guid := range guids {
		connPath := netKey + `\` + guid + `\Connection`
		ck, err := registry.OpenKey(registry.LOCAL_MACHINE, connPath, registry.READ)
		if err != nil {
			continue
		}
		name, _, _ := ck.GetStringValue("Name")
		ck.Close()
		if name != "" {
			fmt.Printf("  GUID: %s  Name: %s\n", guid, name)
		}
	}

	// 2. 找第一个活动适配器测试
	var testGUID, testName string
	for _, guid := range guids {
		connPath := netKey + `\` + guid + `\Connection`
		ck, err := registry.OpenKey(registry.LOCAL_MACHINE, connPath, registry.READ)
		if err != nil {
			continue
		}
		name, _, _ := ck.GetStringValue("Name")
		ck.Close()
		if name != "" {
			testGUID = guid
			testName = name
			break
		}
	}

	if testGUID == "" {
		fmt.Println("未找到适配器")
		return
	}
	fmt.Printf("\n测试适配器: %s (%s)\n", testName, testGUID)

	// 3. 解析 GUID
	s := testGUID
	if !strings.HasPrefix(s, "{") {
		s = "{" + s + "}"
	}
	ole32 := windows.NewLazySystemDLL("ole32.dll")
	clsidFromString := ole32.NewProc("CLSIDFromString")
	guidStr, _ := syscall.UTF16PtrFromString(s)
	var g windows.GUID
	hr, _, _ := clsidFromString.Call(
		uintptr(unsafe.Pointer(guidStr)),
		uintptr(unsafe.Pointer(&g)),
	)
	fmt.Printf("CLSIDFromString hr=0x%08X GUID=%v\n", hr, g)

	// 4. 调用 SetInterfaceDnsSettings
	dnsPtr, _ := syscall.UTF16PtrFromString("8.8.8.8")
	settings := dnsInterfaceSettings{
		Version:    dnsInterfaceSettingsVersion1,
		Flags:      dnsSettingNameserver,
		NameServer: dnsPtr,
	}
	fmt.Printf("settings size=%d Version=%d Flags=0x%X NameServer=%p\n",
		unsafe.Sizeof(settings), settings.Version, settings.Flags, settings.NameServer)

	iphlpapi := windows.NewLazySystemDLL("iphlpapi.dll")
	proc := iphlpapi.NewProc("SetInterfaceDnsSettings")
	ret, _, lastErr := proc.Call(
		uintptr(unsafe.Pointer(&g)),
		uintptr(unsafe.Pointer(&settings)),
	)
	fmt.Printf("SetInterfaceDnsSettings ret=%d lastErr=%v\n", ret, lastErr)
}

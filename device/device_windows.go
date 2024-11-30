package device

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/hellokeke123/anApp/model"
	"github.com/hellokeke123/anApp/winipcfg"
	"github.com/hellokeke123/anApp/winsys"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/tun"
	"io"
	"log"
	"math"
	"net"
	"net/netip"
	"unsafe"
)

// Device is ...
type Device struct {
	// NativeTun is ...
	*tun.NativeTun
	//
	io.Writer
	// Name is ...
	Name string

	fwpmSession uintptr
}

// 做配置
func (device *Device) Config() (err error) {
	return nil
}

// determineGUID is ...
// generate GUID from tun name
func determineGUID(name string) *windows.GUID {
	b := make([]byte, unsafe.Sizeof(windows.GUID{}))
	if _, err := io.ReadFull(hkdf.New(md5.New, []byte(name), nil, nil), b); err != nil {
		return nil
	}
	return (*windows.GUID)(unsafe.Pointer(&b[0]))
}

// CreateTUN is ...
func CreateTUN(name string, mtu int) (dev *Device, err error) {
	dev = &Device{}
	device, err := tun.CreateTUNWithRequestedGUID(name, determineGUID(name), mtu)
	if err != nil {
		return
	}
	dev.NativeTun = device.(*tun.NativeTun)
	//dev.MTU = dev.NativeTun.MTU
	dev.Writer = dev

	if dev.Name, err = dev.NativeTun.Name(); err != nil {
		return
	}
	/*	if dev.MTU, err = dev.NativeTun.MTU; err != nil {
		return
	}*/
	return
}

func (device *Device) InitGateway() {
	// 获取LUID用于配置网络
	link := winipcfg.LUID(device.NativeTun.LUID())

	ip, err := netip.ParsePrefix(model.TunIp)
	if err != nil {
		panic(err)
	}
	err = link.SetIPAddresses([]netip.Prefix{ip})
	//gateway, gerr := netip.ParseAddr("0.0.0.0")
	//if gerr != nil {
	//	fmt.Println("gerr", gerr)
	//}
	////prefix, _ := netip.ParsePrefix("10.0.0.0/24")
	//prefix, _ := netip.ParsePrefix("0.0.0.0/0")
	//link.SetRoutes([]*winipcfg.RouteData{
	//	{prefix, gateway, 1},
	//})
	err = device.addRouteEntry4([]string{"0.0.0.0/0"})
	if err != nil {
		panic(err)
	}

	if model.ContextConfigImp.ContextClient.EnableEnforceDns {
		device.SetDns4(model.DNS1, model.DNS2)
	}
	// 限制流量
	// 后面需要考虑释放问题
	if model.ContextConfigImp.ContextClient.EnableEnforceDns {
		device.configFwp()
	}
	go device.monitorEnforceDns()
}

func (d *Device) monitorEnforceDns() {
	model.ContextConfigImp.ContextClient.ReadEnableEnforceDnsChan = make(chan bool, 1)
	model.ContextConfigImp.ContextClient.WriteEnableEnforceDnsChan = make(chan bool, 1)
	for {
		enableEnforceDns := <-model.ContextConfigImp.ContextClient.ReadEnableEnforceDnsChan
		if enableEnforceDns != model.ContextConfigImp.ContextClient.EnableEnforceDns {
			model.ContextConfigImp.ContextClient.EnableEnforceDns = enableEnforceDns
			if enableEnforceDns {
				d.configFwp()
			} else {
				if d.fwpmSession != 0 {
					winsys.FwpmEngineClose0(d.fwpmSession)
					d.fwpmSession = 0
				}
			}
		}

		model.ContextConfigImp.ContextClient.WriteEnableEnforceDnsChan <- true
	}
}

func (d *Device) SetDns4(dns1 string, dns2 string) {
	luid := winipcfg.LUID(d.NativeTun.LUID())
	err := luid.SetDNS(windows.AF_INET, []netip.Addr{netip.MustParseAddr(dns1), netip.MustParseAddr(dns2)}, []string{})
	if err != nil {
		log.Println("window 设置dns错误", err)
	}

	// luid.FlushDNS(windows.AF_INET)
	// luid.DisableDNSRegistration()

	/*	inetIf, err := luid.IPInterface(winipcfg.AddressFamily(windows.AF_INET))
		if err != nil {
			panic(err)
		}
		inetIf.ForwardingEnabled = true
		inetIf.RouterDiscoveryBehavior = winipcfg.RouterDiscoveryDisabled
		inetIf.DadTransmits = 0
		inetIf.ManagedAddressConfigurationSupported = false
		inetIf.OtherStatefulConfigurationSupported = false
		inetIf.NLMTU = 1500
		inetIf.UseAutomaticMetric = false
		inetIf.Metric = 0
		err = inetIf.Set()
		if err != nil {
			panic(err)
		}*/
}

// 筛选平台 流量过滤
// https://learn.microsoft.com/zh-cn/windows/win32/api/_fwp/
// 限制其它网卡使用dns
func (d *Device) configFwp() {
	var engine uintptr
	session := &winsys.FWPM_SESSION0{Flags: winsys.FWPM_SESSION_FLAG_DYNAMIC}
	err := winsys.FwpmEngineOpen0(nil, winsys.RPC_C_AUTHN_DEFAULT, nil, session, unsafe.Pointer(&engine))
	if err != nil {
		panic(err)
	}
	d.fwpmSession = engine

	subLayerKey, err := windows.GenerateGUID()
	if err != nil {
		panic(err)
	}

	subLayer := winsys.FWPM_SUBLAYER0{}
	subLayer.SubLayerKey = subLayerKey
	subLayer.DisplayData = winsys.CreateDisplayData(d.Name, "auto-route rules")
	subLayer.Weight = math.MaxUint16
	err = winsys.FwpmSubLayerAdd0(engine, &subLayer, 0)
	if err != nil {
		panic(err)
	}

	processAppID, err := winsys.GetCurrentProcessAppID()
	if err != nil {
		panic(err)
	}
	defer winsys.FwpmFreeMemory0(unsafe.Pointer(&processAppID))

	var filterId uint64
	permitCondition := make([]winsys.FWPM_FILTER_CONDITION0, 1)
	permitCondition[0].FieldKey = winsys.FWPM_CONDITION_ALE_APP_ID
	permitCondition[0].MatchType = winsys.FWP_MATCH_EQUAL
	permitCondition[0].ConditionValue.Type = winsys.FWP_BYTE_BLOB_TYPE
	permitCondition[0].ConditionValue.Value = uintptr(unsafe.Pointer(processAppID))

	permitFilter4 := winsys.FWPM_FILTER0{}
	permitFilter4.FilterCondition = &permitCondition[0]
	permitFilter4.NumFilterConditions = 1
	permitFilter4.DisplayData = winsys.CreateDisplayData(d.Name, "protect ipv4")
	permitFilter4.SubLayerKey = subLayerKey
	permitFilter4.LayerKey = winsys.FWPM_LAYER_ALE_AUTH_CONNECT_V4
	permitFilter4.Action.Type = winsys.FWP_ACTION_PERMIT
	permitFilter4.Weight.Type = winsys.FWP_UINT8
	permitFilter4.Weight.Value = uintptr(13)
	permitFilter4.Flags = winsys.FWPM_FILTER_FLAG_CLEAR_ACTION_RIGHT
	err = winsys.FwpmFilterAdd0(engine, &permitFilter4, 0, &filterId)
	if err != nil {
		panic(err)
	}

	netInterface, err := net.InterfaceByName(d.Name)
	if err != nil {
		panic(err)
	}

	tunCondition := make([]winsys.FWPM_FILTER_CONDITION0, 1)
	tunCondition[0].FieldKey = winsys.FWPM_CONDITION_LOCAL_INTERFACE_INDEX
	tunCondition[0].MatchType = winsys.FWP_MATCH_EQUAL
	tunCondition[0].ConditionValue.Type = winsys.FWP_UINT32
	tunCondition[0].ConditionValue.Value = uintptr(uint32(netInterface.Index))

	tunFilter4 := winsys.FWPM_FILTER0{}
	tunFilter4.FilterCondition = &tunCondition[0]
	tunFilter4.NumFilterConditions = 1
	tunFilter4.DisplayData = winsys.CreateDisplayData(d.Name, "allow ipv4")
	tunFilter4.SubLayerKey = subLayerKey
	tunFilter4.LayerKey = winsys.FWPM_LAYER_ALE_AUTH_CONNECT_V4
	tunFilter4.Action.Type = winsys.FWP_ACTION_PERMIT
	tunFilter4.Weight.Type = winsys.FWP_UINT8
	tunFilter4.Weight.Value = uintptr(11)
	err = winsys.FwpmFilterAdd0(engine, &tunFilter4, 0, &filterId)
	if err != nil {
		panic(err)
	}

	blockDNSCondition := make([]winsys.FWPM_FILTER_CONDITION0, 1)
	// https://learn.microsoft.com/zh-cn/windows/win32/fwp/filtering-condition-identifiers-
	blockDNSCondition[0].FieldKey = winsys.FWPM_CONDITION_IP_REMOTE_PORT
	blockDNSCondition[0].MatchType = winsys.FWP_MATCH_EQUAL
	blockDNSCondition[0].ConditionValue.Type = winsys.FWP_UINT16
	blockDNSCondition[0].ConditionValue.Value = uintptr(uint16(53))
	/*		blockDNSCondition[1].FieldKey = winsys.FWPM_CONDITION_IP_PROTOCOL
			blockDNSCondition[1].MatchType = winsys.FWP_MATCH_EQUAL
			blockDNSCondition[1].ConditionValue.Type = winsys.FWP_UINT8
			blockDNSCondition[1].ConditionValue.Value = uintptr(uint8(winsys.IPPROTO_UDP))*/

	blockDNSFilter4 := winsys.FWPM_FILTER0{}
	blockDNSFilter4.FilterCondition = &blockDNSCondition[0]
	blockDNSFilter4.NumFilterConditions = 1
	blockDNSFilter4.DisplayData = winsys.CreateDisplayData(d.Name, "block ipv4 dns")
	blockDNSFilter4.SubLayerKey = subLayerKey
	blockDNSFilter4.LayerKey = winsys.FWPM_LAYER_ALE_AUTH_CONNECT_V4
	blockDNSFilter4.Action.Type = winsys.FWP_ACTION_BLOCK
	blockDNSFilter4.Weight.Type = winsys.FWP_UINT8
	blockDNSFilter4.Weight.Value = uintptr(10)
	err = winsys.FwpmFilterAdd0(engine, &blockDNSFilter4, 0, &filterId)
	if err != nil {
		panic(err)
	}
}

// SetInterfaceAddress is ...
// 192.168.1.11/24
// fe80:08ef:ae86:68ef::11/64
func (d *Device) SetInterfaceAddress(address string) error {
	if _, _, gateway, err := getInterfaceConfig4(address); err == nil {
		return d.setInterfaceAddress4("", address, gateway)
	}
	if _, _, gateway, err := getInterfaceConfig6(address); err == nil {
		return d.setInterfaceAddress6("", address, gateway)
	}
	return errors.New("tun device address error")
}

// setInterfaceAddress4 is ...
// https://github.com/WireGuard/wireguard-windows/blob/ef8d4f03bbb6e407bc4470b2134a9ab374155633/tunnel/addressconfig.go#L60-L168
func (d *Device) setInterfaceAddress4(addr, mask, gateway string) error {
	luid := winipcfg.LUID(d.NativeTun.LUID())

	addresses := append([]netip.Prefix{}, netip.MustParsePrefix(mask))

	err := luid.SetIPAddressesForFamily(windows.AF_INET, addresses)
	if errors.Is(err, windows.ERROR_OBJECT_ALREADY_EXISTS) {
		//cleanupAddressesOnDisconnectedInterfaces(windows.AF_INET, addresses)
		err = luid.SetIPAddressesForFamily(windows.AF_INET, addresses)
	}
	if err != nil {
		return err
	}

	err = luid.SetDNS(windows.AF_INET, []netip.Addr{netip.MustParseAddr(gateway)}, []string{})
	return err
}

// setInterfaceAddress6 is ...
func (d *Device) setInterfaceAddress6(addr, mask, gateway string) error {
	luid := winipcfg.LUID(d.NativeTun.LUID())

	addresses := append([]netip.Prefix{}, netip.MustParsePrefix(mask))

	err := luid.SetIPAddressesForFamily(windows.AF_INET6, addresses)
	if errors.Is(err, windows.ERROR_OBJECT_ALREADY_EXISTS) {
		//cleanupAddressesOnDisconnectedInterfaces(windows.AF_INET6, addresses)
		err = luid.SetIPAddressesForFamily(windows.AF_INET6, addresses)
	}
	if err != nil {
		return err
	}

	err = luid.SetDNS(windows.AF_INET6, []netip.Addr{netip.MustParseAddr(gateway)}, []string{})
	return err
}

// addRouteEntry is ...
func (d *Device) addRouteEntry4(cidr []string) error {
	luid := winipcfg.LUID(d.NativeTun.LUID())

	routes := make(map[winipcfg.RouteData]bool, len(cidr))
	for _, item := range cidr {
		ipNet, err := netip.ParsePrefix(item)
		if err != nil {
			return fmt.Errorf("ParsePrefix error: %w", err)
		}
		routes[winipcfg.RouteData{
			Destination: ipNet,
			NextHop:     netip.IPv4Unspecified(),
			Metric:      0,
		}] = true
	}

	for r := range routes {
		if err := luid.AddRoute(r.Destination, r.NextHop, r.Metric); err != nil {
			return fmt.Errorf("AddRoute error: %w", err)
		}
	}

	return nil
}

// addRouteEntry6 is ...
func (d *Device) addRouteEntry6(cidr []string) error {
	luid := winipcfg.LUID(d.NativeTun.LUID())

	routes := make(map[winipcfg.RouteData]bool, len(cidr))
	for _, item := range cidr {
		ipNet, err := netip.ParsePrefix(item)
		if err != nil {
			return fmt.Errorf("ParsePrefix error: %w", err)
		}
		routes[winipcfg.RouteData{
			Destination: ipNet,
			NextHop:     netip.IPv6Unspecified(),
			Metric:      0,
		}] = true
	}

	for r := range routes {
		if err := luid.AddRoute(r.Destination, r.NextHop, r.Metric); err != nil {
			return fmt.Errorf("AddRoute error: %w", err)
		}
	}

	return nil
}

func (d *Device) Close() (err error) {
	err = d.NativeTun.Close()
	if d.fwpmSession != 0 {
		winsys.FwpmEngineClose0(d.fwpmSession)
	}
	return
}

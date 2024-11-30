package device

import (
	"errors"
	"github.com/hellokeke123/anApp/model"
	"net"
	"unsafe"
)

// NewDevice is ...
func NewDevice(name string) (*Device, error) {
	return CreateTUN(name, model.MTU)
}

// DeviceType is ...
func (d *Device) DeviceType() string {
	return "wireguard"
}

// Write is ...
func (d *Device) Write(b []byte) (int, error) {
	bfs := make([][]byte, 1)
	bfs[0] = b
	return d.NativeTun.Write(bfs, 0)
}

func (d *Device) Read(p []byte) (n int, err error) {
	bfs := make([][]byte, 1)
	bfs[0] = p
	sizes := make([]int, 1)
	in, err := d.NativeTun.Read(bfs, sizes, 0)
	if in < 1 {
		return 0, err
	} else {
		return sizes[0], err
	}

}

// AddRouteEntry is ...
// 198.18.0.0/16
// 8.8.8.8/32
func (d *Device) AddRouteEntry(cidr []string) error {
	cidr4 := make([]string, 0, len(cidr))
	cidr6 := make([]string, 0, len(cidr))
	for _, item := range cidr {
		ip, _, err := net.ParseCIDR(item)
		if err != nil {
			return err
		}
		if ip.To4() != nil {
			cidr4 = append(cidr4, item)
			continue
		}
		if ip.To16() != nil {
			cidr6 = append(cidr6, item)
			continue
		}
	}
	if len(cidr4) > 0 {
		if err := d.addRouteEntry4(cidr4); err != nil {
			return err
		}
	}
	if len(cidr6) > 0 {
		if err := d.addRouteEntry6(cidr6); err != nil {
			return err
		}
	}
	return nil
}

// getInterfaceConfig4 is ...
func getInterfaceConfig4(cidr string) (addr, mask, gateway string, err error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		err = errors.New("not ipv4 address")
		return
	}

	addr = ipv4.String()
	mask = net.IP(ipNet.Mask).String()
	ipv4 = ipNet.IP.To4()
	ipv4[net.IPv4len-1]++
	gateway = ipv4.String()

	return
}

// getInterfaceConfig6 is ...
func getInterfaceConfig6(cidr string) (addr, mask, gateway string, err error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return
	}

	ipv6 := ip.To16()
	if ipv6 == nil {
		err = errors.New("not ipv6 address")
		return
	}

	addr = ipv6.String()
	mask = net.IP(ipNet.Mask).String()
	ipv6 = ipNet.IP.To16()
	ipv6[net.IPv6len-1]++
	gateway = ipv6.String()

	return
}

// parse4 is ...
func parse4(addr string) [4]byte {
	ip := net.ParseIP(addr).To4()
	return *(*[4]byte)(unsafe.Pointer(&ip[0]))
}

// parse6 is ...
func parse6(addr string) [16]byte {
	ip := net.ParseIP(addr).To16()
	return *(*[16]byte)(unsafe.Pointer(&ip[0]))
}

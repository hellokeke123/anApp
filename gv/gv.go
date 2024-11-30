package gv

import (
	"fmt"
	"github.com/hellokeke123/anApp/device"
	"github.com/hellokeke123/anApp/endpoint"
	ltcp "github.com/hellokeke123/anApp/handler/remoteSocketHandler/tcp"
	ludp "github.com/hellokeke123/anApp/handler/remoteSocketHandler/udp"
	"golang.org/x/time/rate"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

func Start(device *device.Device, mtu int) (err error) {
	var stackImp = stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			ipv6.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
			icmp.NewProtocol6,
		},
		HandleLocal: false,
	})
	defer func(stackImp *stack.Stack) {
		if err != nil {
			stackImp.Close()
		}
	}(stackImp)

	// set NICID to 1
	const NICID = tcpip.NICID(1)

	// WithDefaultTTL sets the default TTL used by stack.
	{
		opt := tcpip.DefaultTTLOption(64)
		if tcperr := stackImp.SetNetworkProtocolOption(ipv4.ProtocolNumber, &opt); tcperr != nil {
			err = fmt.Errorf("set ipv4 default TTL: %s", tcperr)
			return
		}
		if tcperr := stackImp.SetNetworkProtocolOption(ipv6.ProtocolNumber, &opt); tcperr != nil {
			err = fmt.Errorf("set ipv6 default TTL: %s", tcperr)
			return
		}
	}

	// set forwarding
	if tcperr := stackImp.SetForwardingDefaultAndAllNICs(ipv4.ProtocolNumber, true); tcperr != nil {
		err = fmt.Errorf("set ipv4 forwarding error: %s", tcperr)
		return
	}
	if tcperr := stackImp.SetForwardingDefaultAndAllNICs(ipv6.ProtocolNumber, true); tcperr != nil {
		err = fmt.Errorf("set ipv6 forwarding error: %s", tcperr)
		return
	}

	// WithICMPBurst sets the number of ICMP messages that can be sent
	// in a single burst.
	stackImp.SetICMPBurst(50)

	// WithICMPLimit sets the maximum number of ICMP messages permitted
	// by rate limiter.
	stackImp.SetICMPLimit(rate.Limit(1000))

	//stackImp.SetForwardingDefaultAndAllNICs(header.IPv4ProtocolNumber, false)

	// WithTCPBufferSizeRange sets the receive and send buffer size range for TCP.
	{
		rcvOpt := tcpip.TCPReceiveBufferSizeRangeOption{Min: 4 << 10, Default: 212 << 10, Max: 4 << 20}
		if tcperr := stackImp.SetTransportProtocolOption(tcp.ProtocolNumber, &rcvOpt); tcperr != nil {
			err = fmt.Errorf("set TCP receive buffer size range: %s", tcperr)
			return
		}
		sndOpt := tcpip.TCPSendBufferSizeRangeOption{Min: 4 << 10, Default: 212 << 10, Max: 4 << 20}
		if tcperr := stackImp.SetTransportProtocolOption(tcp.ProtocolNumber, &sndOpt); tcperr != nil {
			err = fmt.Errorf("set TCP send buffer size range: %s", tcperr)
			return
		}
	}

	// WithTCPCongestionControl sets the current congestion control algorithm.
	{
		opt := tcpip.CongestionControlOption("reno")
		if tcperr := stackImp.SetTransportProtocolOption(tcp.ProtocolNumber, &opt); tcperr != nil {
			err = fmt.Errorf("set TCP congestion control algorithm: %s", tcperr)
			return
		}
	}

	// WithTCPModerateReceiveBuffer sets receive buffer moderation for TCP.
	{
		opt := tcpip.TCPDelayEnabled(false)
		if tcperr := stackImp.SetTransportProtocolOption(tcp.ProtocolNumber, &opt); tcperr != nil {
			err = fmt.Errorf("set TCP delay: %s", err)
			return
		}
	}
	stackImp.HandleLocal()
	// WithTCPModerateReceiveBuffer sets receive buffer moderation for TCP.
	{
		opt := tcpip.TCPModerateReceiveBufferOption(true)
		if tcperr := stackImp.SetTransportProtocolOption(tcp.ProtocolNumber, &opt); tcperr != nil {
			err = fmt.Errorf("set TCP moderate receive buffer: %s", tcperr)
			return
		}
	}

	// WithTCPSACKEnabled sets the SACK option for TCP.
	{
		opt := tcpip.TCPSACKEnabled(true)
		if tcperr := stackImp.SetTransportProtocolOption(tcp.ProtocolNumber, &opt); tcperr != nil {
			err = fmt.Errorf("set TCP SACK: %s", tcperr)
			return
		}
	}

	//mustSubnet := func(s string) tcpip.Subnet {
	//	_, ipNet, err := net.ParseCIDR(s)
	//	if err != nil {
	//		log.Panic(fmt.Errorf("unable to ParseCIDR(%s): %w", s, err))
	//	}
	//	subnet, err := tcpip.NewSubnet(tcpip.AddrFrom4([4]byte(ipNet.IP)), tcpip.MaskFrom("0.0.0.0"))
	//	if err != nil {
	//		log.Panic(fmt.Errorf("unable to NewSubnet(%s): %w", ipNet, err))
	//	}
	//	return subnet
	//}

	// Add default route table for IPv4 and IPv6
	// This will handle all incoming ICMP packets.
	subnet, err := tcpip.NewSubnet(tcpip.AddrFrom4([4]byte{0, 0, 0, 0}), tcpip.MaskFromBytes([]byte{0, 0, 0, 0}))
	stackImp.SetRouteTable([]tcpip.Route{
		{
			// Destination: header.IPv4EmptySubnet,
			Destination: subnet,
			NIC:         NICID,
		},
		//{
		//	// Destination: header.IPv6EmptySubnet,
		//	Destination: mustSubnet("::/0"),
		//	NIC:         NICID,
		//},
	})
	// Important: We must initiate transport protocol handlers
	// before creating NIC, otherwise NIC would dispatch packets
	// to stack and cause race condition.
	// 处理tcp连接
	handler := ltcp.TcpHandler{
		StackImp: stackImp,
	}
	stackImp.SetTransportProtocolHandler(tcp.ProtocolNumber, tcp.NewForwarder(stackImp, 16<<10, 1<<15, handler.HandleStream).HandlePacket)

	stackImp.SetTransportProtocolHandler(udp.ProtocolNumber, ludp.NewUdpStack(stackImp, true).HandlePacket)

	// WithCreatingNIC creates NIC for stack.
	if tcperr := stackImp.CreateNIC(NICID, endpoint.NewEndpoint(device, mtu)); tcperr != nil {
		err = fmt.Errorf("fail to create NIC in stack: %s", tcperr)
		return
	}
	// 欺骗性获取端点，所有ip都能获取端点
	stackImp.SetSpoofing(NICID, true)
	//
	//networkEndpoint, _ := stackImp.GetNetworkEndpoint(NICID, header.IPv4ProtocolNumber)
	//networkEndpoint
	// WithPromiscuousMode sets promiscuous mode in the given NIC.
	// In past we did s.AddAddressRange to assign 0.0.0.0/0 onto
	// the interface. We need that to be able to terminate all the
	// incoming connections - to any ip. AddressRange API has been
	// removed and the suggested workaround is to use Promiscuous
	// mode. https://github.com/google/gvisor/issues/3876
	//
	// Ref: https://github.com/majek/slirpnetstack/blob/master/stack.go
	if tcperr := stackImp.SetPromiscuousMode(NICID, true); tcperr != nil {
		err = fmt.Errorf("set promiscuous mode: %s", tcperr)
		return
	}
	return
}

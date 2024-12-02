package udp

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/hellokeke123/anApp/model"
	"github.com/hellokeke123/anApp/protocol/dns"
	"github.com/hellokeke123/anApp/protocol/socket"
	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"io"
	"log"
	"net"
	"net/netip"
	"strconv"
)

type UdpStack struct {
	*stack.Stack
	noChecksum    bool
	useDefaultTTL bool
	ttl           uint8
	tos           uint8
	owner         tcpip.PacketOwner
}

func NewUdpStack(stack *stack.Stack, noChecksum bool) *UdpStack {
	// 处理udp连接
	return &UdpStack{
		Stack:         stack,
		noChecksum:    true,
		useDefaultTTL: true,
		ttl:           0,
		tos:           0,
		owner:         nil,
	}
}

// HandlePacket is to handle UDP connections
func (udpStack *UdpStack) HandlePacket(id stack.TransportEndpointID, pkbf *stack.PacketBuffer) bool {
	// Ref: gVisor pkg/tcpip/transport/udp/endpoint.go HandlePacket
	/*	if id.RemotePort != 8080 {
		return false
	}*/

	go udpStack.HandleRemotePacket(id, pkbf.Clone())

	return true
}

func (udpStack *UdpStack) HandleRemotePacket(id stack.TransportEndpointID, pkbf *stack.PacketBuffer) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("HandlePacket from panic: %v", r)
		}
	}()

	sourPayloadData := make([]byte, 2024)
	sourceBuffer := pkbf.ToBuffer()
	// 荷载数据 ， 不包含头
	atn, perr := sourceBuffer.ReadAt(sourPayloadData, int64(pkbf.HeaderSize()))
	if perr != nil && perr != io.EOF {
		log.Println("udp读取源数据错误", perr)
		return
	}

	route, tcperr := udpStack.FindRoute(pkbf.NICID, pkbf.Network().DestinationAddress(), pkbf.Network().SourceAddress(), pkbf.NetworkProtocolNumber, false)
	if tcperr != nil {
		log.Println("udp转发错误", tcperr)
	}

	log.Println("UDP"+"远程地址 ", pkbf.Network().DestinationAddress(), " 源地址：", pkbf.Network().SourceAddress())
	// 返回远程数据
	remoteErr := udpStack.readRemoteDate(func(destData []byte) error {
		sdError := udpStack.send(route, id, buffer.MakeWithData(destData))
		if sdError != nil {
			log.Println("udp 发送错误:", sdError)
			return errors.New(sdError.String())
		}
		return nil
	}, sourPayloadData[:atn], pkbf.Network().DestinationAddress().String()+":"+strconv.Itoa(int(id.LocalPort)), strconv.Itoa(int(id.RemotePort)))

	if remoteErr != nil {
		log.Println("udp远程错误", remoteErr)
	}
}

// send sends the given packet.
func (udpStack *UdpStack) send(route *stack.Route, id stack.TransportEndpointID, buffer buffer.Buffer) tcpip.Error {
	const ProtocolNumber = header.UDPProtocolNumber

	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		ReserveHeaderBytes: header.UDPMinimumSize + int(route.MaxHeaderLength()),
		Payload:            buffer,
	})
	pkt.Owner = udpStack.owner
	defer pkt.DecRef()

	// Initialize the UDP header.
	udp := header.UDP(pkt.TransportHeader().Push(header.UDPMinimumSize))
	pkt.TransportProtocolNumber = ProtocolNumber

	//fmt.Println("发送", "远程地址端口 ", id.RemotePort, " 源地址端口：", id.LocalPort)

	length := uint16(pkt.Size())
	udp.Encode(&header.UDPFields{
		SrcPort: id.LocalPort,
		DstPort: id.RemotePort,
		Length:  length,
	})

	// Set the checksum field unless TX checksum offload is enabled.
	// On IPv4, UDP checksum is optional, and a zero value indicates the
	// transmitter skipped the checksum generation (RFC768).
	// On IPv6, UDP checksum is not optional (RFC2460 Section 8.1).
	if route.RequiresTXTransportChecksum() &&
		(!udpStack.noChecksum || route.NetProto() == header.IPv6ProtocolNumber) {
		xsum := route.PseudoHeaderChecksum(ProtocolNumber, length)
		udp.SetChecksum(^udp.CalculateChecksum(xsum))
	}

	if udpStack.useDefaultTTL {
		udpStack.ttl = route.DefaultTTL()
	}

	/*	if id.RemotePort == 8080 {
		pktData := make([]byte, 2024)
		pkt.ToView().ReadAt(pktData, 0)
		parseUdpDataPacket(pktData)
	}*/

	// 写入包
	if err := route.WritePacket(stack.NetworkHeaderParams{
		Protocol: ProtocolNumber,
		TTL:      udpStack.ttl,
		TOS:      udpStack.tos,
	}, pkt); err != nil {
		route.Stats().UDP.PacketSendErrors.Increment()
		return err
	}

	// Track count of packets sent.
	route.Stats().UDP.PacketsSent.Increment()
	return nil
}

func parseUdpDataPacket(data []byte) {

	packet := gopacket.NewPacket(data, layers.LayerTypeUDP, gopacket.Default)

	//// IP层
	//ipLayer := packet.Layer(layers.LayerTypeIPv4)
	//if ipLayer != nil {
	//	ip, _ := ipLayer.(*layers.IPv4)
	//	fmt.Println("IP Src IP:", ip.SrcIP)
	//	fmt.Println("IP Dst IP:", ip.DstIP)
	//}

	// UDP层
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		fmt.Println("UDP Src Port:", udp.SrcPort)
		fmt.Println("UDP Dst Port:", udp.DstPort)
		fmt.Println("UDP Length:", udp.Length)
		fmt.Println("UDP Checksum:", udp.Checksum)
		fmt.Printf("UDP Payload: %s\n", udp.Payload)
	}
}

/*
*

	创建远程连接并发送请求
*/
func (udpStack *UdpStack) readRemoteDate(sendLocal func(destData []byte) error, data []byte, remoteAddressStr, port string) error {
	addrPort, _ := netip.ParseAddrPort(remoteAddressStr)

	// 强制修改 dns
	if model.ContextConfigImp.ContextClient.IsEnableEnforceDns(addrPort.Port()) {

		buff, err := dns.SendDoh(model.ContextConfigImp.ContextClient.GetEnableEnforceDHO(), bytes.NewBuffer(data))

		if buff == nil || err != nil {
			log.Println("dns-doh失败:", err)
			return err
		} else {
			log.Println("dns-doh成功")
			err := sendLocal(buff.Bytes())
			return err
		}
	}
	remoteAddr, _ := net.ResolveUDPAddr(model.UDP, remoteAddressStr)
	route := model.FindContainRoute(remoteAddr.IP, model.Routes)
	dialer := socket.GetDialer(route)

	if true {

		log.Println(model.UDP, "direct", route.Ip.String(), "==>", remoteAddr.String())
		identify := route.Ip.String() + port + "-" + remoteAddressStr
		//portNumber, _ := strconv.Atoi(port)
		go attachConnect(identify, data,
			&connectionState{
				dst:       remoteAddr,
				dial:      dialer,
				sendLocal: sendLocal,
			})
		//if err != nil {
		//	log.Println(model.UDP, "direct", route.Ip.String(), "==>", remoteAddr.String(), "返回错误", err)
		//}
	}
	return nil
}

///*
//*
//
//	创建远程连接并发送请求
//
//func (udpStack *UdpStack) readRemoteDate(sendLocal func(destData []byte) error, data []byte, remoteAddressStr string) error {
//	addrPort, _ := netip.ParseAddrPort(remoteAddressStr)
//
//	// 强制修改 dns
//	if model.ContextConfigImp.ContextClient.IsEnableEnforceDns(addrPort.Port()) {
//
//		buff, err := dns.SendDoh(model.ContextConfigImp.ContextClient.GetEnableEnforceDHO(), bytes.NewBuffer(data))
//
//		if buff == nil || err != nil {
//			log.Println("dns-doh失败:", err)
//			return err
//		} else {
//			log.Println("dns-doh成功")
//			err := sendLocal(buff.Bytes())
//			return err
//		}
//	}
//	remoteAddr, _ := net.ResolveUDPAddr(model.UDP, remoteAddressStr)
//	route := model.FindContainRoute(remoteAddr.IP, model.Routes)
//	dialer := socket.GetDialer(route)
//
//	if true {
//
//		log.Println(model.UDP, "direct", route.Ip.String(), "==>", remoteAddr.String())
//		// 拨号连接
//		c, err := dialer.Dial(model.UDP, remoteAddr.String())
//		if err != nil {
//			return err
//		}
//		conn := c.(*net.UDPConn)
//		defer conn.Close()
//
//		// 发送数据到服务器
//		_, err = conn.Write(data)
//		if err != nil {
//			return err
//		}
//		// 接收服务器的响应
//		for {
//			conn.SetDeadline(time.Now().Add(2 * time.Second))
//			byts := make([]byte, 4096)
//			n, _, err := conn.ReadFromUDP(byts)
//
//			if n > 0 {
//				slerr := sendLocal(byts[:n])
//				if slerr != nil {
//					return slerr
//				}
//				if n < 4096 {
//					log.Println(model.UDP, ":正常退出", err)
//					break
//				}
//
//				if err != nil && err == io.EOF {
//					conn.Close()
//					log.Println(model.UDP, ":正常退出", err)
//					break
//				}
//			} else if err != nil && err != io.EOF {
//				conn.Close()
//				log.Println(model.UDP, ":发送到本地报错", err)
//				return err
//			} else if err != nil && err == io.EOF {
//				conn.Close()
//				log.Println(model.UDP, ":正常退出", err)
//				break
//			}
//		}
//		return nil
//	}
//	return nil
//}
//*/

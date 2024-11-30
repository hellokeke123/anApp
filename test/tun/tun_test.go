package netlink

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/hellokeke123/anApp/device"
	"github.com/hellokeke123/anApp/model"
	"github.com/hellokeke123/anApp/winipcfg"
	"log"
	"net"
	"net/netip"
	"testing"
)

/*
*

		tun 实现的是驱动处理，也就是物理层到网络层之间数据的处理

	    可以理解为读和写都是物理层获得的数据 ，网络层还是被系统托管的，相当于是第三方主机
*/
func TestTunWrite(t *testing.T) {
	newDevice, err := device.NewDevice("anApp1")
	if err != nil {
		t.Fatal(err)
	}
	defer newDevice.Close()
	newDevice.InitGateway()
	// 获取LUID用于配置网络
	link := winipcfg.LUID(newDevice.NativeTun.LUID())

	ip, err := netip.ParsePrefix(model.TunIp)
	if err != nil {
		panic(err)
	}
	err = link.SetIPAddresses([]netip.Prefix{ip})
	if err != nil {
		panic(err)
	}

	// 创建 IP 层
	lip := layers.IPv4{
		SrcIP:    net.IP{1, 2, 3, 4},
		DstIP:    net.IP{10, 0, 0, 2},
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}

	// 创建 TCP SYN 包
	tcp := layers.TCP{
		SrcPort:    layers.TCPPort(12345),
		DstPort:    layers.TCPPort(30341),
		Seq:        0,
		Ack:        0,
		DataOffset: 5,
		SYN:        true,
		Window:     14600,
	}

	tcp.SetNetworkLayerForChecksum(&lip)

	// 创建 Packet
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	err = gopacket.SerializeLayers(buffer, options,
		&lip,
		&tcp,
	)
	if err != nil {
		log.Fatal(err)
	}

	// 发送 Packet
	packetData := buffer.Bytes()
	newDevice.Write(packetData)
}

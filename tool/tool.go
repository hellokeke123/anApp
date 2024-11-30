package tool

import (
	"github.com/hellokeke123/anApp/model"
	"golang.org/x/net/ipv4"
	"net"
	"strconv"
	"strings"
)

func ParseIpData(data []byte) (packet *model.IpPacket, err error) {
	packet = &model.IpPacket{}

	header := ipv4.Header{}

	err = header.Parse(data)

	if err == nil {
		packet.Header = &header
	} else {
		panic(err)
	}

	// 判断是否为 TCP 协议
	switch header.Protocol {
	case 1:
		packet.Protocol = "icmp"
	case 6:
		packet.Protocol = "tcp"
	case 17:
		packet.Protocol = "udp"
	default:
		packet.Protocol = "unknown"

	}
	return
}

// 计算ip校验和
func CalculateChecksum(data []byte) uint16 {
	length := len(data)
	sum := uint32(0)

	// 将数据按照每 16 位进行拆分，然后相加
	for i := 0; i < length-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	// 处理奇数个字节的情况
	if length%2 != 0 {
		sum += uint32(data[length-1]) << 8
	}

	// 将溢出部分加到结果中
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	// 取反得到校验和
	checksum := uint16(^sum)

	return checksum
}

// tcp校验和
func CalculateTcpChecksum(data []byte) uint16 {
	sum := uint32(0)

	// 遍历每两个字节
	for i := 0; i < len(data)-1; i += 2 {
		// 将两个字节合并为一个 16 位数
		word := uint32(data[i])<<8 + uint32(data[i+1])
		// 加到校验和上
		sum += uint32(word)
	}

	// 如果数据长度为奇数，则最后一个字节单独处理
	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1])
	}

	// 将进位的高 16 位加到低 16 位上
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	// 取反得到校验和
	checksum := ^uint16(sum)

	return checksum
}

func IPToUInt32(ipnr net.IP) uint32 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum uint32

	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)

	return sum
}

package model

import "golang.org/x/net/ipv4"

type IpPacket struct {
	Header   *ipv4.Header // 头信息
	Data     []byte       //
	Protocol string
}

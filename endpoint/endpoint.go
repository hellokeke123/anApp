package endpoint

import (
	"errors"
	"github.com/hellokeke123/anApp/device"
	"github.com/hellokeke123/anApp/model"
	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/sync"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"io"
	"log"
	_ "unsafe"
)

// Endpoint is ...
type Endpoint struct {
	// Endpoint is ...
	*channel.Endpoint
	// Device is ...
	*device.Device
	// Writer is ...
	Writer io.Writer

	mtu  int
	mu   sync.Mutex
	buff []byte
}

func NewEndpoint(dev *device.Device, mtu int) stack.LinkEndpoint {
	var wt io.Writer = dev
	if wt == nil {
		log.Panic(errors.New("not a valid device for windows"))
	}
	ep := &Endpoint{
		Endpoint: channel.New(512, uint32(mtu), ""),
		Device:   dev,
		Writer:   wt,
		buff:     make([]byte, mtu),
	}
	ep.Endpoint.AddNotify(ep)
	return ep
}

// Attach is to attach device to stack
// 从网卡读取数据
func (e *Endpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.Endpoint.Attach(dispatcher)

	// WinDivert has no Reader
	var r io.Reader = e.Device
	if r == nil {
		log.Panic(errors.New("not a valid device for windows"))
		return
	}
	// WinTun
	go func(r io.Reader, size int, ep *channel.Endpoint) {
		for {
			buf := make([]byte, size)
			nr, err := r.Read(buf)
			if err != nil {
				break
			}
			buf = buf[:nr]
			//fmt.Println(ipPacket.Protocol, ipPacket.Header.Src, ipPacket.Header.Dst)
			pktBuffer := stack.NewPacketBuffer(stack.PacketBufferOptions{
				ReserveHeaderBytes: 0,
				Payload:            buffer.MakeWithData(buf),
				IsForwardedPacket:  true,
			})
			switch header.IPVersion(buf) {
			case header.IPv4Version:
				ep.InjectInbound(header.IPv4ProtocolNumber, pktBuffer)
			case header.IPv6Version:
				ep.InjectInbound(header.IPv6ProtocolNumber, pktBuffer)
			}
			pktBuffer.DecRef()
		}
	}(r, model.MTU+4, e.Endpoint)
}

// WriteNotify is to write packets back to system
// 回写数据
func (e *Endpoint) WriteNotify() {
	pkt := e.Endpoint.Read()

	e.mu.Lock()
	buf := append(e.buff[:0], pkt.NetworkHeader().View().AsSlice()...)
	buf = append(buf, pkt.TransportHeader().View().AsSlice()...)
	vv := pkt.Data()

	n := 2048
	data := make([]byte, n)
	toBuffer := vv.ToBuffer()
	for {
		at, err := toBuffer.ReadAt(data, 0)
		if err == io.EOF {
			buf = append(buf, data[:at]...)
			break
		} else {
			buf = append(buf, data[:at]...)
		}
	}
	_, err := e.Writer.Write(buf)
	if err != nil {
		log.Printf("WriteNotify err: %v", err)
	}
	e.mu.Unlock()
}

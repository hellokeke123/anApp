package tcp

import (
	"fmt"
	"github.com/hellokeke123/anApp/model"
	"github.com/hellokeke123/anApp/protocol/socket"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
	"io"
	"log"
	"net"
	"time"
)

type TcpHandler struct {
	StackImp *stack.Stack
}

// HandleStream is to handle incoming TCP connections
func (th *TcpHandler) HandleStream(r *tcp.ForwarderRequest) {
	id := r.ID()
	wq := waiter.Queue{}
	ep, tcperr := r.CreateEndpoint(&wq)
	if tcperr != nil {
		fmt.Println("tcp<--->create endpoint error: %v", tcperr)
		// prevent potential half-open TCP connection leak.
		r.Complete(true)
		return
	}
	r.Complete(false)

	// set keepalive
	if err := func(ep tcpip.Endpoint) error {
		ep.SocketOptions().SetKeepAlive(true)
		idleOpt := tcpip.KeepaliveIdleOption(60 * time.Second)
		if tcperr := ep.SetSockOpt(&idleOpt); tcperr != nil {
			return fmt.Errorf("set keepalive idle: %s", tcperr)
		}
		intervalOpt := tcpip.KeepaliveIntervalOption(30 * time.Second)
		if tcperr := ep.SetSockOpt(&intervalOpt); tcperr != nil {
			return fmt.Errorf("set keepalive interval: %s", tcperr)
		}
		return nil
	}(ep); err != nil {
		//s.Error("tcp %v:%v <---> %v:%v create endpoint error: %v",
		//	net.IP(id.RemoteAddress),
		//	int(id.RemotePort),
		//	net.IP(id.LocalAddress),
		//	int(id.LocalPort),
		//	err,
		//)
	}

	go th.TcpHandle(gonet.NewTCPConn(&wq, ep), &net.TCPAddr{IP: net.IP(id.LocalAddress.AsSlice()), Port: int(id.LocalPort)}, &net.TCPAddr{IP: net.IP(id.RemoteAddress.AsSlice()), Port: int(id.RemotePort)})
}

// 处理tcp连接
func (th *TcpHandler) TcpHandle(tcp *gonet.TCPConn, localAddress *net.TCPAddr, remoteAddress *net.TCPAddr) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("TcpHandle from panic: %v", r)
		}
	}()
	//log.Println("tcp", localAddress.IP, remoteAddress.IP)
	//n := 2048
	//buf := make([]byte, n)
	//tcp.Read(buf)
	//log.Println("获得数据", string(buf))

	remoteAddr, err := net.ResolveTCPAddr("tcp", localAddress.AddrPort().String())
	//log.Println(model.TCP, "远程ip:", localAddress.IP.String())

	route := model.FindContainRoute(remoteAddr.IP, model.Routes)
	// 创建连接
	dialer := socket.GetDialer(route)
	//localAddr, err := net.ResolveTCPAddr("tcp", route.Ip.String()+":0")
	/*	localAddr, err := net.ResolveTCPAddr("tcp", route.Ip.String()+":0")
		log.Println(model.TCP, "", localAddr.String(), "==>", remoteAddr.String())*/

	if err != nil {
		log.Println("远程失败:", err)
	}

	if err != nil {
		log.Println("远程失败:", err)
	}

	//if true {
	if true {
		// 拨号连接
		c, err := dialer.Dial(model.TCP, remoteAddr.String())
		if err != nil {
			log.Println(model.TCP, " direct ", route.Ip.String(), "==>", remoteAddr.String(), "连接失败:", err)
			return
		}
		conn := c.(*net.TCPConn)
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Duration(10) * time.Second)
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		defer conn.Close()
		// 将两个连接之间的数据复制

		go func() {
			_, err = io.Copy(conn, tcp)
			if err != nil {
				log.Println(model.TCP, " direct ", "从本地连接到远程连接复制数据出错:", err)
			}
		}()

		_, err = io.Copy(tcp, conn)
		if err != nil {
			log.Println(model.TCP, " direct ", "从远程连接到本地连接复制数据出错:", err)
		}
		conn.Close()
		tcp.Close()
	}

	//log.Println("获得数据", string(buf))
}

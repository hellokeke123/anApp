package udp

import (
	"github.com/hellokeke123/anApp/model"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

/**
主要原理是udp保持连接, 根据udp的标识（源ip端口 目的ip端口保证连接一致）
*/

const bufferStateSize = 1024

type connectionState struct {
	identify           string
	src                *net.UDPAddr
	dst                *net.UDPAddr
	dial               net.Dialer
	conn               *net.UDPConn
	lastActive         time.Time
	sendLocal          func(destData []byte) error // data callback
	disconnectCallback func(conn *connectionState)
	writeLock          sync.Mutex
}
type ForwardState struct {
	// srcIp:srcPort
	connections map[string]*connectionState
	deadline    time.Duration
	closed      bool
}

// 锁
var lock sync.Mutex

var forwardState = ForwardState{
	connections: make(map[string]*connectionState),
	deadline:    time.Minute * 5,
}

// 附加连接
func attachConnect(identify string, buff []byte, connState *connectionState) (err error) {
	defer func() {
		if r := recover(); r != nil {
			//log.Println("udp => attachConnect", r)
		}

	}()
	// 检查是否已有连接
	state, ok := forwardState.connections[identify]
	if !ok {
		lock.Lock()
		defer lock.Unlock()

		state, ok = forwardState.connections[identify]
		if !ok {
			connState.dial.LocalAddr = connState.src
			c, err := connState.dial.Dial(model.UDP, connState.dst.String())
			if err != nil {
				//log.Println("udp dial error:", err)
				return err
			} else {
				connState.conn = c.(*net.UDPConn)
			}
			go connState.run()
			forwardState.connections[identify] = connState

			connState.writeLock.Lock()
			defer connState.writeLock.Unlock()
			connState.conn.Write(buff)
			state = connState
		}

	} else {
		state.writeLock.Lock()
		defer state.writeLock.Unlock()
		state.conn.Write(buff)
	}

	return err
}

// 处理远端读写
func (connState *connectionState) run() {
	defer func() {
		if r := recover(); r != nil {
			//log.Println("udp => run ==> err", r)
		}
		connState.conn.Close()
		delete(forwardState.connections, connState.identify)

	}()

	_, err := io.Copy(connState, connState.conn)
	if err != nil {
		log.Println("upd.io.Copy", err)
	}
}

// 这里的写到tun网卡,返回数据给客户端
func (connState *connectionState) Write(p []byte) (n int, err error) {
	err = connState.sendLocal(p)
	if err != nil {
		return 0, err
	}
	return len(p), err
}

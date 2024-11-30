package Gorilla

import (
	"fmt"
	"github.com/hashicorp/yamux"
	"net"
	"testing"
)

func TestServerYamux(t *testing.T) {
	// Accept a TCP connection
	listener, _ := net.Listen("tcp", "192.168.5.19:8487")
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			session, err := yamux.Server(conn, nil)
			if err != nil {
				return
			}

			for {
				// Accept a stream
				stream, err := session.Accept()
				if err != nil {
					return
				}

				go func() {
					// Listen for a message
					buf := make([]byte, 1024)
					for {
						n, rerr := stream.Read(buf)
						if rerr != nil {
							break
						}
						fmt.Println("data：", string(buf[:n]))
					}
				}()
			}

		}()
	}

}

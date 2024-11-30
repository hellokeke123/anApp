package Gorilla

import (
	"github.com/hashicorp/yamux"
	"log"
	"net"
	"testing"
	"time"
)

func TestLocalYamux(t *testing.T) {
	// Get a TCP connection
	conn, err := net.Dial("tcp", "192.168.5.19:8487")
	if err != nil {
		panic(err)
	}

	// Setup client side of yamux
	session, err := yamux.Client(conn, nil)
	if err != nil {
		panic(err)
	}

	// Open a new stream
	stream, err := session.Open()
	if err != nil {
		panic(err)
	}

	// Stream implements net.Conn
	stream.Write([]byte("ping"))

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			message := []byte("Hello, WebSocket server!")

			stream.Write(message)
			log.Println("Message sent to server")
		}
	}
}

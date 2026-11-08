package main

import (
	"bufio"
	"io"
	"log"
	"net"
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("Client connected:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				log.Printf("Read error: %v", err)
			}
			return
		}

		_, err = conn.Write([]byte(message))
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()

	log.Println("Echo server listening on :8090")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}
		go handleConn(conn)
	}

}

/*
echo server
- listener, accept loop, goroutine per client, echo bytes back
learn:
- tcp is a byte stream not messages. connections are long lived and read returning io.eof means the peer closed
test
-two or three nc sessions at once
*/

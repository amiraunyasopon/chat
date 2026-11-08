package main

import "fmt" 

func main() {
	fmt.Println("Hello, World!")
}

/*
echo server
- listener, accept loop, goroutine per client, echo bytes back
learn:
- tcp is a byte stream not messages. connections are long lived and read returning io.eof means the peer closed
test
-two or three nc sessions at once
*/

/* return ip:port */

package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "'host:port'")
		os.Exit(1)
	}
	service := os.Args[1]

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	fmt.Printf("Start a new listener on %s\n", tcpAddr.String())

	for {
		conn, err := listener.Accept()

		if err != nil {
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(time.Minute))
			fmt.Printf("(%s):(%s): Accepted\n", time.Now(), conn.RemoteAddr().String())

			response := `HTTP/1.1 200 OK
Content-Length: %d
Server: onion
Content-Type: text/plain; charset=utf-8

%s`
			response = fmt.Sprintf(response, len(conn.RemoteAddr().String()), conn.RemoteAddr().String())
			conn.Write([]byte(response))
		}(conn)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

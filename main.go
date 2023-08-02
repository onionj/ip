/*
  YourIP Server
  Simple TCP server to return the client's IP address.
*/

package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

var version string = "" // set it just in Makefile

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
	fmt.Printf("Start a new listener on %s\nversion:%s\n", tcpAddr.String(), version)

	const RESPONSE string = `HTTP/1.1 200 OK
Content-Length: %d
Server: Onion
Content-Type: text/plain; charset=utf-8

%s`
	for {
		conn, err := listener.Accept()

		if err != nil {
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()
			fmt.Printf("(%s) %s Accepted\n", time.Now().String()[:23], conn.RemoteAddr().String())

			conn.SetDeadline(time.Now().Add(time.Minute))

			bf := make([]byte, 1024)
			n, err := conn.Read(bf)
			if err == nil {
				fmt.Printf("(%s) Read %d byte from %s\n", time.Now().String()[:23], n, conn.RemoteAddr().String())
			}

			remoteHost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				remoteHost = conn.RemoteAddr().String()
			}

			response := fmt.Sprintf(RESPONSE, len(remoteHost), remoteHost)

			n, err = conn.Write([]byte(response))
			if err == nil {
				fmt.Printf("(%s) Write %d byte to %s\n", time.Now().String()[:23], n, conn.RemoteAddr().String())
			}
		}(conn)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

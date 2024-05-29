/*
  YourIP Server
  Simple TCP server to return the client's IP address.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
	"yourip/geolocation"
)

var version = "" // set it just in Makefile

const RESPONSE string = `HTTP/1.1 200 OK
Content-Length: %d
Server: Onion
Content-Type: %s; charset=utf-8

%s`

const REFRESHPEROIDH = 200

const (
	AnimationModeBanner = "1\n"
	AnimationModeFlight = "2\n"
)

type JsonResponse struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
}

var (
	bindAddress      string
	workers          int
	useIPGeolocation bool
	ipGeolocation    *geolocation.IPGeolocation
)

func main() {
	flag.StringVar(&bindAddress, "bind", ":8080", "The address to bind the TCP server to. (shorthand -b)")
	flag.StringVar(&bindAddress, "b", ":8080", "(shorthand for --bind)")

	flag.IntVar(&workers, "worker", 3, "number of workers")
	flag.IntVar(&workers, "w", 3, "number of workers")

	flag.BoolVar(&useIPGeolocation, "geolocation", false, "ip geolocation service (shorthand -g)")
	flag.BoolVar(&useIPGeolocation, "g", false, "(shorthand for --geolocation)")
	flag.Parse()

	if useIPGeolocation {
		ipGeolocation = geolocation.New(time.Duration(REFRESHPEROIDH))
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", bindAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	defer func(listener *net.TCPListener) {
		_ = listener.Close()
	}(listener)
	fmt.Printf("Start a new listener on %s\nversion:%s\n", tcpAddr.String(), version)

	for ; workers != 0; workers-- {
		fmt.Println("Running Worker", workers)
		go Worker(listener)
	}

	ch := make(chan struct{})
	<-ch
}

func Worker(listener *net.TCPListener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept err:", err)
			continue
		}

		fmt.Printf("(%s) %s Accepted\n", time.Now().String()[:23], conn.RemoteAddr().String())

		err = conn.SetDeadline(time.Now().Add(time.Minute))
		if err != nil {
			fmt.Println("SetDeadline err:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("(%s) %s Close err: %s\n", time.Now().String()[:23], conn.RemoteAddr().String(), err.Error())
		} else {
			fmt.Printf("(%s) %s Closed\n", time.Now().String()[:23], conn.RemoteAddr().String())
		}
	}()

	requestBuffer := make([]byte, 2048)
	bytesRead, err := conn.Read(requestBuffer)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Connection closed by client")
		} else {
			fmt.Println("Error reading request:", err)
			return
		}
	}

	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	remoteAddrStr := remoteAddr.IP.String()

	country := ""
	if useIPGeolocation {
		country, _ = ipGeolocation.Query(remoteAddr.IP)
	}

	// for netcat client:
	if bytesRead <= 1 {
		fmt.Printf("(%s) return tcp text response to %s (%s)\n", time.Now().String()[:23], remoteAddrStr, country)
		_, _ = conn.Write([]byte(fmt.Sprintf("%s %s", remoteAddrStr, country)))
		return
	}

	request := string(requestBuffer[:bytesRead])

	if strings.Contains(request, " /json") {
		fmt.Printf("(%s) return json response to %s (%s)\n", time.Now().String()[:23], remoteAddrStr, country)

		jsonResponseByte, err := json.Marshal(JsonResponse{IP: remoteAddrStr, Country: country})
		if err != nil {
			fmt.Println("create json response err:", err)
			return
		}
		_, _ = conn.Write([]byte(fmt.Sprintf(RESPONSE, len(jsonResponseByte), "application/json", jsonResponseByte)))
		return

	} else {
		ipWithCountry := fmt.Sprintf("%s %s", remoteAddrStr, country)

		if request == AnimationModeBanner || request == AnimationModeFlight {
			StreamAnimation(conn, ipWithCountry, request)
			return
		}

		fmt.Printf("(%s) return http text/plain response to %s (%s)\n", time.Now().String()[:23], remoteAddrStr, country)
		_, _ = conn.Write([]byte(fmt.Sprintf(RESPONSE, len(ipWithCountry), "text/plain", ipWithCountry)))
	}
}

// StreamAnimation : Stream Client IP Like an Animation (for example client open a tcp connection with netcat and send `1`)
func StreamAnimation(conn net.Conn, response string, animationMode string) {
	switch animationMode {
	case AnimationModeBanner:
		fmt.Printf("(%s) Stream animation to %s (Banner)\n", time.Now().String()[:23], response)

		faces := []string{"(^_^)", "[o_o]", "(^.^)", "(\".\")", "($.$)"}
		randomIndex := rand.Intn(len(faces))
		face := faces[randomIndex]

		response += "  " + face + "                  "

		for {
			if _, err := conn.Write(append(append([]byte(" ["), []byte(response)...), []byte("]\x0D")...)); err != nil {
				break
			}
			time.Sleep(time.Second / 5)

			// shift string
			response = response[len(response)-1:] + response[:len(response)-1]
		}

	case AnimationModeFlight:
		fmt.Printf("(%s) Stream animation to %s (Flight)\n", time.Now().String()[:23], response)

		response = "    " + response + "    "
		flight := ` %s\                                  
 %s|      |~\______/~~\__  |          
 %s|______ \_____======= )-+          
 %s|                 |/    |          
 %s/                 ()               `
		flight = fmt.Sprintf(
			flight,
			strings.Repeat("-", len(response)),
			strings.Repeat(" ", len(response)),
			response,
			strings.Repeat(" ", len(response)),
			strings.Repeat("-", len(response)))

		arrayFlight := strings.Split(flight, "\n")

		for {

			if _, err := conn.Write(append([]byte(strings.Join(arrayFlight, "\n")), []byte("\x0D\u001b[4A")...)); err != nil {
				break
			}

			time.Sleep(time.Second / 5)

			// shift string
			for index, lineFlight := range arrayFlight {
				arrayFlight[index] = lineFlight[len(lineFlight)-1:] + lineFlight[:len(lineFlight)-1]
			}

		}
	}
}

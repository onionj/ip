/*
  YourIP Server
  Simple TCP server to return the client's IP address.
*/

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var version string = "" // set it just in Makefile

const RESPONSE string = `HTTP/1.1 200 OK
Content-Length: %d
Server: Onion
Content-Type: %s; charset=utf-8

%s`

type ResponseMode int8

const (
	ResponseModeText ResponseMode = iota
	ResponseModeHTTPText
	ResponseModeHTTPJson
	ResponseModeTextAnimation
)

const (
	AnimationModeBanner = "1\n"
	AnimationModeFlight = "2\n"
)

type JsonResponse struct {
	IP string `json:"ip"`
}

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

	for {
		if conn, err := listener.Accept(); err == nil {
			go Handler(conn)
		}
	}
}

// Accept the connection, Prepare Response and send it to client
func Handler(conn net.Conn) {
	defer func() {
		conn.Close()
		fmt.Printf("(%s) %s Closed\n", time.Now().String()[:23], conn.RemoteAddr().String())
	}()

	fmt.Printf("(%s) %s Accepted\n", time.Now().String()[:23], conn.RemoteAddr().String())

	conn.SetDeadline(time.Now().Add(time.Minute))

	// Read request from tcp connection
	RequestBuffer := make([]byte, 1024)
	numberOfBytes, err := conn.Read(RequestBuffer)
	if err == nil {
		fmt.Printf("(%s) Read %d byte from %s\n", time.Now().String()[:23], numberOfBytes, conn.RemoteAddr().String())
	}

	response := ""
	animationMode := ""

	// Chose response type
	if numberOfBytes <= 1 {
		response, _ = CreateResponse(conn.RemoteAddr().String(), ResponseModeText)

	} else {
		request := string(RequestBuffer[:numberOfBytes])

		if strings.Contains(request, "GET /json") {
			response, err = CreateResponse(conn.RemoteAddr().String(), ResponseModeHTTPJson)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else if request == AnimationModeBanner || request == AnimationModeFlight {
			response, _ = CreateResponse(conn.RemoteAddr().String(), ResponseModeText)
			animationMode = request

		} else {
			response, _ = CreateResponse(conn.RemoteAddr().String(), ResponseModeHTTPText)
		}
	}

	// Write response to TCP connection
	if animationMode != "" {
		StreamAnimation(conn, response, animationMode)
	} else {
		numberOfBytes, err = conn.Write([]byte(response))
		if err == nil {
			fmt.Printf("(%s) Write %d byte to %s\n", time.Now().String()[:23], numberOfBytes, conn.RemoteAddr().String())
		}
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// Return Client IP in these formats:
//
//	simple text: for pure TCP like NetCat
//	http text:   for browsers and curl
//	http JSON:   for API Call
func CreateResponse(remoteAddr string, mode ResponseMode) (string, error) {
	remoteHost, _, _ := net.SplitHostPort(remoteAddr)

	switch mode {

	case ResponseModeHTTPText:
		return fmt.Sprintf(RESPONSE, len(remoteHost), "text/plain", remoteHost), nil

	case ResponseModeHTTPJson:
		jsonResponse, err := json.Marshal(JsonResponse{IP: remoteHost})
		if err != nil {
			return "", fmt.Errorf("failed to marshal message to JSON: %v", err)
		}
		return fmt.Sprintf(RESPONSE, len(jsonResponse), "application/json", jsonResponse), nil

	default:
		return remoteHost, nil
	}
}

// Stream Client IP Like an Animation!
//
// > for example client open a tcp connection with netcat and send `1`
func StreamAnimation(conn net.Conn, response string, animationMode string) {
	switch animationMode {
	case AnimationModeBanner:
		fmt.Printf("(%s) Stream animation to %s (Banner)\n", time.Now().String()[:23], conn.RemoteAddr().String())

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
		fmt.Printf("(%s) Stream animation to %s (Flight)\n", time.Now().String()[:23], conn.RemoteAddr().String())

		response = "    " + response + "    "
		flight := `            %s\                                  
            %s|      |~~\_____/~~\__  |          
            %s|______ \______====== )-+          
            %s|              ~~~|/~~  |          
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
			for indx, lineFlight := range arrayFlight {
				arrayFlight[indx] = lineFlight[len(lineFlight)-1:] + lineFlight[:len(lineFlight)-1]
			}

		}
	}
}


# TCP Server Example

This is a simple HTTP server written in Go that listens for incoming connections and responds the client's IP address.

## Usage
### Method One:

`./ip 'host:port'`
Replace `'host:port'` with the desired host and port to listen on.

For example: `./ip :80`

### Method Two:
`go run main.go 'host:port'`

Replace `'host:port'` with the desired host and port to listen on.


## How it works

The server uses the `net` package in Go to resolve the TCP address, start a listener, and accept incoming connections. It sets a deadline of one minute for each connection and responds with an HTTP 200 OK message containing the client's IP address.

## License

This project is licensed under the [MIT License](https://github.com/onionj/ip/blob/master/LICENSE).

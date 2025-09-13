package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/mohdrashid9678/r-http/internal/request"
	"github.com/mohdrashid9678/r-http/internal/response"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalln("error setting up listener:", err)
	}
	defer listener.Close()
	log.Println("Server is listening on port 42069...")

	// The main loop accepts connections and spawns goroutines.
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection processes a single incoming network connection.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("Connection accepted from %s.", remoteAddr)

	// Parse the request
	req, err := request.Parse(conn)
	if err != nil {
		log.Printf("error parsing request from %s: %v", remoteAddr, err)

		// If parsing fails, create and send a 400 Bad Request response.
		var parseErr *request.ParseError
		if errors.As(err, &parseErr) {
			errorBody := []byte(parseErr.Error())
			resp := response.New(parseErr.StatusCode, errorBody)

			if writeErr := resp.Write(conn); writeErr != nil {
				log.Printf("error sending error response to %s: %v", remoteAddr, writeErr)
			}
		}
		return // Stop processing this connection.
	}

	// Send the response
	body := []byte(fmt.Sprintf("Hello! You sent a %s request for the path: %s", req.Method, req.Target))
	resp := response.New(200, body)

	// Added a chad header
	resp.Headers["Server"] = "RHTTP-Server/1.0"

	if err := resp.Write(conn); err != nil {
		log.Printf("error writing response to %s: %v", remoteAddr, err)
	}

	log.Printf("Connection from %s closed.", remoteAddr)
}

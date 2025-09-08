package main

import (
	"fmt"
	"log"
	"net"

	"github.com/mohdrashid9678/r-http/internal/request"
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
// Its responsibility is now much clearer: it uses the request package
// to parse the request and then prints the result.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Connection accepted from %s.\n", conn.RemoteAddr())

	// Use the Parse function from our new package. This is the key interaction.
	req, err := request.Parse(conn)
	if err != nil {
		log.Printf("Error parsing request from %s: %v", conn.RemoteAddr(), err)
		// In a real server, we would write a "400 Bad Request" response here.
		return
	}

	// Print the structured request data that was parsed by the request package.
	fmt.Println("--- Parsed Request ---")
	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("Target: %s\n", req.Target)
	fmt.Printf("Version: %s\n", req.Version)
	fmt.Println("Headers:")
	for key, value := range req.Headers {
		fmt.Printf("  %s: %s\n", key, value)
	}
	fmt.Println("----------------------")

	fmt.Printf("Connection from %s closed.\n", conn.RemoteAddr())
}

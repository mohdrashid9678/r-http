# rhttp

**A Minimal HTTP/1.1 Implementation from Scratch in Go**

`rhttp` is a learning-oriented, practical implementation of the HTTP/1.1 protocol in Go. Instead of relying on Go's `net/http` package, this project is built directly on top of raw TCP connections, implementing the HTTP/1.1 protocol (RFC 9112) step by step.

With `rhttp`, you can:

- Understand how HTTP/1.1 works under the hood.
- Explore parsing of requests, building of responses, and routing.
- Easily build a custom HTTP server on top of it.

This project aims to provide both a working server and an educational foundation for anyone curious about how HTTP servers are built at the protocol level.

---

## Table of Contents

- [Why rhttp](#why-rhttp)
- [Core Principles](#core-principles)
- [Features](#features)
- [Install](#install)
- [Quick Start](#quick-start)

  - [Parse a Request](#parse-a-request)
  - [Build a Response](#build-a-response)
  - [Assemble a Minimal Server](#assemble-a-minimal-server)

- [API Overview](#api-overview)
- [Design Notes](#design-notes)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Why rhttp

Unlike frameworks or toolkits, `rhttp` is about **learning by doing**. It’s an implementation of HTTP/1.1 over TCP, giving you full control of the connection lifecycle and the ability to see what really happens when a browser talks to a server.

This makes it great for:

- Developers who want to build an HTTP server from scratch.
- Students learning networking, protocols, and systems programming in Go.
- Experimenting with protocol extensions, research, or custom servers.

---

## Core Principles

1. **Build from Scratch** — Every part of HTTP/1.1 (parsing, responses, routing) is implemented manually on top of Go's `net` package.
2. **Educational Value** — The codebase emphasizes clarity, correctness, and mapping directly to RFC 9112.
3. **Practical Server Building** — While educational, `rhttp` is fully capable of powering simple HTTP servers.

---

## Features

- **Raw TCP-based request parsing** — Parses the request-line, headers, and body stream directly from the socket.
- **Response builder** — Write correct HTTP/1.1 responses to any `io.Writer`.
- **Radix-tree router** — Supports parameterized routes and efficient handler lookups.
- **Helper functions** — JSON and text responses with minimal boilerplate.
- **Custom error types** — For handling bad requests, not found, and other HTTP errors.

---

## Install

```bash
go get github.com/mohdrashid9678/rhttp
```

```bash
go get github.com/mohdrashid9678/rhttp/request
```

```bash
go get github.com/mohdrashid9678/rhttp/response
```

```bash
go get github.com/mohdrashid9678/rhttp/httperrors
```

---

## Quick Start

### Start a Minimal Server

```go
package main

import (
	"log"

	"github.com/mohdrashid9678/rhttp"
	"github.com/mohdrashid9678/rhttp/request"
	"github.com/mohdrashid9678/rhttp/response"
)

func main() {
	// 1. Create a new server instance
	server := rhttp.New(":8080")

	// 2. Register a handler for the home page
	server.AddRoute("GET", "/", handleHomePage)

	// 3. Start the server.
	log.Println("Starting server on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleHomePage is the handler for the "/" route.
func handleHomePage(req *request.Request) (*response.Response, error) {
	log.Printf("Serving request for target: %s", req.Target)
	return response.Text(200, "Welcome to the home page!")
}
```

### Parse a Request

```go
package main

import (
    "fmt"
    "log"
    "strings"

    "github.com/mohdrashid9678/rhttp/request"
)

func main() {
    raw := "GET /users/123 HTTP/1.1\r\n" +
        "Host: api.example.com\r\n" +
        "Accept: application/json\r\n\r\n"

    r := strings.NewReader(raw)

    req, err := request.Parse(r)
    if err != nil {
        log.Fatalf("parse error: %v", err)
    }
    defer req.Body.Close()

    fmt.Println("Method:", req.Method)
    fmt.Println("Target:", req.Target)
    fmt.Println("Host:", req.Headers.Get("Host"))
}
```

### Build a Response

```go
package main

import (
    "bytes"
    "log"

    "github.com/mohdrashid9678/rhttp/response"
)

func main() {
    data := map[string]string{"status": "ok", "user_id": "123"}
    resp := response.JSON(200, data)
    resp.Headers.Set("X-Custom-Header", "my-value")

    var buf bytes.Buffer
    if err := resp.Write(&buf); err != nil {
        log.Fatalf("write failed: %v", err)
    }

    log.Printf("raw response:\n%s", buf.String())
}
```

---

## API Overview

- **request**: Parse raw HTTP requests → `Parse(io.Reader) (*Request, error)`
- **response**: Build and write HTTP responses → `Write(io.Writer) error`
- **router**: Register routes and map them to handlers → parameterized routing supported
- **httperrors**: Predefined HTTP error helpers

---

## Roadmap

- Add middleware support (logging, authentication).
- Context support (`context.Context`) for cancellations and deadlines.
- Chunked transfer encoding.
- Persistent connections / keep-alive.
- Experimental HTTP/2 implementation.

---

## Contributing

Contributions are welcome — especially improvements to protocol correctness, examples, and documentation.

---

## License

MIT License — see [LICENSE](LICENSE).

package request

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
)

// Request now contains a streaming body and path parameters.
type Request struct {
	Method     string
	Target     string
	Version    string
	Headers    map[string]string
	Body       io.ReadCloser
	PathParams map[string]string
	ctx        context.Context
}

// bodyReader implements io.ReadCloser for the request body.
type bodyReader struct {
	io.Reader
	closer io.Closer
}

func (br *bodyReader) Close() error {
	// Closing the body reader should not close the underlying connection
	// until the response is written. We can make this a no-op for now.
	return nil
}

// Parse now constructs a streaming body reader instead of reading it all.
func Parse(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	req := &Request{
		Headers:    make(map[string]string),
		PathParams: make(map[string]string),
		ctx:        context.Background(),
	}

	if err := parseRequestLine(reader, req); err != nil {
		return nil, err
	}
	if err := parseHeaders(reader, req); err != nil {
		return nil, err
	}

	contentLengthStr := req.Headers["Content-Length"]
	if contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64); err == nil && contentLength > 0 {
		req.Body = &bodyReader{
			Reader: io.LimitReader(reader, contentLength),
			closer: conn,
		}
	} else {
		// Body is empty or Content-Length is invalid/missing.
		req.Body = &bodyReader{
			Reader: strings.NewReader(""),
			closer: conn,
		}
	}

	return req, nil
}

func parseRequestLine(r *bufio.Reader, req *Request) error {
	line, _, err := r.ReadLine()
	if err != nil {
		return err
	}
	parts := strings.Split(string(line), " ")
	if len(parts) != 3 {
		return errors.New("malformed request line")
	}
	req.Method, req.Target, req.Version = parts[0], parts[1], parts[2]
	return nil
}

func parseHeaders(r *bufio.Reader, req *Request) error {
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			return err
		}
		if len(line) == 0 {
			break
		}
		parts := strings.SplitN(string(line), ":", 2)
		if len(parts) != 2 {
			continue // Malformed header
		}
		req.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return nil
}

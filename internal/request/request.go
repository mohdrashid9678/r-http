package request

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// ParseError represents a client-facing error that occurs during request parsing.
// It includes a suggested HTTP status code that the server can return.
type ParseError struct {
	// StatusCode is the suggested HTTP status code to return to the client.
	StatusCode int
	// Msg is the developer-facing error message.
	Msg string
}

// Error implements the standard error interface.
func (e *ParseError) Error() string {
	return e.Msg
}

// Request holds all the parsed information from an incoming HTTP request.
// It is exported so it can be used by other packages (like main).
type Request struct {
	Method  string
	Target  string
	Version string
	Headers map[string]string
	Body    []byte
}

// It parses the request. It coordinates the parsing process
// by calling helper functions for each part of the request.
func Parse(r io.Reader) (*Request, error) {
	reader := bufio.NewReader(r)

	method, target, version, err := parseRequestLine(reader)
	if err != nil {
		return nil, err
	}

	headers, err := parseHeaders(reader)
	if err != nil {
		return nil, err
	}

	body, err := parseBody(reader, headers)
	if err != nil {
		return nil, err
	}

	return &Request{
		Method:  method,
		Target:  target,
		Version: version,
		Headers: headers,
		Body:    body,
	}, nil
}

// parseRequestLine handles the first line of the HTTP request.
func parseRequestLine(reader *bufio.Reader) (method, target, version string, err error) {
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", &ParseError{StatusCode: 400, Msg: "failed to read request line"}
	}

	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) != 3 {
		return "", "", "", &ParseError{StatusCode: 400, Msg: fmt.Sprintf("malformed request line: %q", requestLine)}
	}

	return parts[0], parts[1], parts[2], nil
}

// parseHeaders reads the block of header lines.
func parseHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)
	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, &ParseError{StatusCode: 400, Msg: "failed to read header line"}
		}
		headerLine = strings.TrimSpace(headerLine)
		if headerLine == "" {
			break // Empty line signifies the end of headers.
		}

		key, value, found := strings.Cut(headerLine, ":")
		if !found {
			log.Printf("ignoring malformed header: %s", headerLine)
			continue
		}
		headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return headers, nil
}

// parseBody reads the request body, if one is specified by Content-Length.
func parseBody(reader *bufio.Reader, headers map[string]string) ([]byte, error) {
	contentLengthStr, ok := headers["Content-Length"]
	if !ok {
		return nil, nil // No body to read.
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, &ParseError{StatusCode: 400, Msg: fmt.Sprintf("invalid Content-Length value: %q", contentLengthStr)}
	}

	if contentLength == 0 {
		return nil, nil // Body is explicitly empty.
	}

	body := make([]byte, contentLength)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return nil, &ParseError{StatusCode: 400, Msg: "failed to read full request body"}
	}

	return body, nil
}

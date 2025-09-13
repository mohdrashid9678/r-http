package response

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// Some sensible status text
var statusText = map[int]string{
	200: "OK",
	201: "Created",
	204: "No Content",
	400: "Bad Request",
	404: "Not Found",
	500: "Internal Server Error",
}

// New creates a basic Response object. It automatically sets the StatusText,
// a default Content-Type, and calculates the Content-Length header.
func New(statusCode int, body []byte) *Response {
	resp := &Response{
		StatusCode: statusCode,
		StatusText: statusText[statusCode], // Look up the text for the code.
		Headers:    make(map[string]string),
		Body:       body,
	}

	// Use a fallback for unknown status codes.
	if resp.StatusText == "" {
		resp.StatusText = "Status Unknown"
	}

	// Set essential default headers.
	resp.Headers["Content-Length"] = strconv.Itoa(len(body))
	resp.Headers["Content-Type"] = "text/plain"

	return resp
}

// Write sends the formatted HTTP response to the provided io.Writer.
// It uses a bufio.Writer for improved I/O performance.
func (r *Response) Write(w io.Writer) error {
	writer := bufio.NewWriter(w)

	// 1. Write the status line (e.g., HTTP/1.1 200 OK\r\n).
	if _, err := fmt.Fprintf(writer, "HTTP/1.1 %d %s\r\n", r.StatusCode, r.StatusText); err != nil {
		return err
	}

	// 2. Write all the headers.
	for key, value := range r.Headers {
		if _, err := fmt.Fprintf(writer, "%s: %s\r\n", key, value); err != nil {
			return err
		}
	}

	// 3. Write the blank line separator.
	if _, err := writer.WriteString("\r\n"); err != nil {
		return err
	}

	// 4. Write the body, if it exists.
	if len(r.Body) > 0 {
		if _, err := writer.Write(r.Body); err != nil {
			return err
		}
	}

	// 5. Flush the buffer to ensure all data is sent to the underlying writer.
	return writer.Flush()
}

package response

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mohdrashid9678/rhttp/httperrors"
)

// Response can now be created from an io.Reader for streaming.
type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       io.Reader
}

var statusText = map[int]string{
	200: "OK", 201: "Created", 400: "Bad Request",
	404: "Not Found", 500: "Internal Server Error",
}

// New creates a response with a streaming body.
func New(statusCode int, body io.Reader) *Response {
	return &Response{
		StatusCode: statusCode,
		StatusText: statusText[statusCode],
		Headers:    make(map[string]string),
		Body:       body,
	}
}

// Text is a helper to create a plain text response.
func Text(statusCode int, text string) (*Response, error) {
	resp := New(statusCode, strings.NewReader(text))
	resp.Headers["Content-Type"] = "text/plain; charset=utf-8"
	resp.Headers["Content-Length"] = strconv.Itoa(len(text))
	return resp, nil
}

// JSON is a helper to create a JSON response.
func JSON(statusCode int, v interface{}) (*Response, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	resp := New(statusCode, bytes.NewReader(data))
	resp.Headers["Content-Type"] = "application/json; charset=utf-8"
	resp.Headers["Content-Length"] = strconv.Itoa(len(data))
	return resp, nil
}

// Error is a helper to create a response from an error.
func Error(err error) (*Response, error) {
	var httpErr *httperrors.HTTPError
	if errors.As(err, &httpErr) {
		return Text(httpErr.StatusCode, httpErr.Message)
	}
	// Fallback for unexpected errors.
	return Text(500, "Internal Server Error")
}

// Write sends the response to the client. It now supports streaming bodies.
func (r *Response) Write(w io.Writer) error {
	writer := bufio.NewWriter(w)
	fmt.Fprintf(writer, "HTTP/1.1 %d %s\r\n", r.StatusCode, r.StatusText)
	for k, v := range r.Headers {
		fmt.Fprintf(writer, "%s: %s\r\n", k, v)
	}
	writer.WriteString("\r\n")
	if r.Body != nil {
		if _, err := io.Copy(writer, r.Body); err != nil {
			return err
		}
	}
	return writer.Flush()
}

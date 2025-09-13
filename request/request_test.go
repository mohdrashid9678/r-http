package request

import (
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParse is a table-driven test that covers various valid and invalid request formats.
func TestParse(t *testing.T) {
	testCases := []struct {
		name            string   // A descriptive name for the test case.
		rawRequest      string   // The raw HTTP request string to be parsed.
		expectErr       bool     // True if we expect Parse to return an error.
		expectedRequest *Request // The expected Request struct if parsing succeeds.
		expectedBody    []byte   // The expected body content.
	}{
		{
			name: "Good GET Request",
			rawRequest: "GET /path/to/resource HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"User-Agent: AwesomeClient/1.0\r\n" +
				"Accept: */*\r\n\r\n",
			expectErr: false,
			expectedRequest: &Request{
				Method:  "GET",
				Target:  "/path/to/resource",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Host":       "localhost:42069",
					"User-Agent": "AwesomeClient/1.0",
					"Accept":     "*/*",
				},
			},
			expectedBody: []byte{},
		},
		{
			name: "Good POST Request with Body",
			rawRequest: "POST /api/users HTTP/1.1\r\n" +
				"Host: api.example.com\r\n" +
				"Content-Type: application/json\r\n" +
				"Content-Length: 27\r\n\r\n" +
				`{"username":"test","age":30}`,
			expectErr: false,
			expectedRequest: &Request{
				Method:  "POST",
				Target:  "/api/users",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Host":           "api.example.com",
					"Content-Type":   "application/json",
					"Content-Length": "27",
				},
			},
			expectedBody: []byte(`{"username":"test","age":30}`),
		},
		{
			name: "Request with extra whitespace in headers",
			rawRequest: "GET / HTTP/1.1\r\n" +
				"Host:    localhost:42069 \r\n" + // Extra whitespace
				"\r\n",
			expectErr: false,
			expectedRequest: &Request{
				Method:  "GET",
				Target:  "/",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Host": "localhost:42069",
				},
			},
			expectedBody: []byte{},
		},
		{
			name:       "Malformed Request Line - Too few parts",
			rawRequest: "GET /\r\n\r\n",
			expectErr:  true,
		},
		{
			name: "Malformed Header - No colon",
			rawRequest: "GET / HTTP/1.1\r\n" +
				"Host localhost:42069\r\n" +
				"\r\n",
			expectErr: false, // The current implementation skips malformed headers.
			expectedRequest: &Request{
				Method:  "GET",
				Target:  "/",
				Version: "HTTP/1.1",
				Headers: map[string]string{}, // Empty because the bad header was skipped.
			},
			expectedBody: []byte{},
		},
		{
			name: "Invalid Content-Length value",
			rawRequest: "POST /submit HTTP/1.1\r\n" +
				"Content-Length: not-a-number\r\n\r\n" +
				"some body data",
			expectErr: false, // Parse doesn't fail, just treats body as empty.
			expectedRequest: &Request{
				Method:  "POST",
				Target:  "/submit",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Content-Length": "not-a-number",
				},
			},
			expectedBody: []byte{},
		},
		{
			name: "Empty Body with Content-Length 0",
			rawRequest: "POST /submit HTTP/1.1\r\n" +
				"Content-Length: 0\r\n\r\n",
			expectErr: false,
			expectedRequest: &Request{
				Method:  "POST",
				Target:  "/submit",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Content-Length": "0",
				},
			},
			expectedBody: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use net.Pipe to create an in-memory connection for testing.
			// clientConn simulates the client, and serverConn simulates the server.
			clientConn, serverConn := net.Pipe()

			// Start a goroutine to write the raw request to the client side of the pipe.
			// This simulates a client sending data.
			go func() {
				defer clientConn.Close()
				_, err := clientConn.Write([]byte(tc.rawRequest))
				require.NoError(t, err)
			}()

			// Call Parse with the server side of the pipe, which now implements net.Conn.
			r, err := Parse(serverConn)

			if tc.expectErr {
				require.Error(t, err, "Expected an error for this test case")
				return // Test is complete.
			}

			// If we're here, no error was expected.
			require.NoError(t, err, "Did not expect an error for this test case")
			require.NotNil(t, r, "Parsed request should not be nil")
			defer r.Body.Close()

			// Assert individual fields for clarity and simplicity.
			expected := tc.expectedRequest
			assert.Equal(t, expected.Method, r.Method, "Method does not match")
			assert.Equal(t, expected.Target, r.Target, "Target does not match")
			assert.Equal(t, expected.Version, r.Version, "Version does not match")
			assert.Equal(t, expected.Headers, r.Headers, "Headers do not match")

			// Read and compare the streaming body separately.
			bodyBytes, readErr := io.ReadAll(r.Body)
			require.NoError(t, readErr, "Reading the request body should not produce an error")
			assert.Equal(t, tc.expectedBody, bodyBytes, "Body content does not match")
		})
	}
}

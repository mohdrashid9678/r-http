package request

import (
	"strings"
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
				Body: nil,
			},
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
				Body: []byte(`{"username":"test","age":30}`),
			},
		},
		{
			name: "Request with extra whitespace in headers",
			rawRequest: "GET / HTTP/1.1\r\n" +
				"Host:    localhost:42069 \r\n" + // Extra whitespace
				"  User-Agent : MyClient  \r\n" + // Extra whitespace
				"\r\n",
			expectErr: false,
			expectedRequest: &Request{
				Method:  "GET",
				Target:  "/",
				Version: "HTTP/1.1",
				Headers: map[string]string{
					"Host":       "localhost:42069",
					"User-Agent": "MyClient",
				},
				Body: nil,
			},
		},
		{
			name:       "Malformed Request Line - Too few parts",
			rawRequest: "GET / HTTP/1.1oops\r\n\r\n",
			expectErr:  true,
		},
		{
			name:       "Malformed Header - No colon",
			rawRequest: "GET / HTTP/1.1\r\n" + "Host localhost:42069\r\n" + "\r\n",
			expectErr:  false, // The current implementation logs and skips malformed headers, so no error is returned.
			expectedRequest: &Request{
				Method:  "GET",
				Target:  "/",
				Version: "HTTP/1.1",
				Headers: map[string]string{}, // Empty because the bad header was skipped.
				Body:    nil,
			},
		},
		{
			name: "Invalid Content-Length value",
			rawRequest: "POST /submit HTTP/1.1\r\n" +
				"Content-Length: not-a-number\r\n\r\n" +
				"some body data",
			expectErr: true,
		},
		{
			name: "Incomplete Body - Content-Length is too high",
			rawRequest: "POST /submit HTTP/1.1\r\n" +
				"Content-Length: 100\r\n\r\n" +
				"body is shorter than 100 bytes",
			expectErr: true,
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
				Body: nil, // Should be nil or empty slice, our implementation results in nil.
			},
		},
	}

	// Loop over all test cases and run them as subtests.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Parse(strings.NewReader(tc.rawRequest))

			if tc.expectErr {
				// We expected an error.
				require.Error(t, err, "Parse() should have returned an error")
				// Check if it's our custom error type.
				var parseErr *ParseError
				assert.ErrorAs(t, err, &parseErr, "Error should be of type *ParseError")
			} else {
				// We did not expect an error.
				require.NoError(t, err, "Parse() should not have returned an error")
				require.NotNil(t, r, "Parsed request should not be nil")
				// Use assert.Equal to compare the expected and actual structs.
				// This gives a nice diff if they don't match.
				assert.Equal(t, tc.expectedRequest, r)
			}
		})
	}
}

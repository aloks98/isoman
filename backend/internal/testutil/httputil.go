package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockHTTPServer represents a mock HTTP server for testing
type MockHTTPServer struct {
	Server         *httptest.Server
	RequestCount   int
	LastRequestURL string
	ResponseCode   int
	ResponseBody   string
	ResponseDelay  int // in milliseconds
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer() *MockHTTPServer {
	mock := &MockHTTPServer{
		ResponseCode: http.StatusOK,
		ResponseBody: "test content",
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.RequestCount++
		mock.LastRequestURL = r.URL.String()

		w.WriteHeader(mock.ResponseCode)
		fmt.Fprint(w, mock.ResponseBody)
	}))

	return mock
}

// URL returns the base URL of the mock server
func (m *MockHTTPServer) URL() string {
	return m.Server.URL
}

// Close shuts down the mock server
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// SetResponse sets the response code and body for the mock server
func (m *MockHTTPServer) SetResponse(code int, body string) {
	m.ResponseCode = code
	m.ResponseBody = body
}

// Reset resets the request count
func (m *MockHTTPServer) Reset() {
	m.RequestCount = 0
	m.LastRequestURL = ""
}

// NewMockDownloadServer creates a mock server that simulates file downloads
func NewMockDownloadServer(t *testing.T, fileSize int, statusCode int) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			w.WriteHeader(statusCode)
			return
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		w.WriteHeader(http.StatusOK)

		// Write fake content
		content := make([]byte, fileSize)
		for i := range content {
			content[i] = byte('A' + (i % 26))
		}
		w.Write(content)
	}))
}

// NewMockChecksumServer creates a mock server that returns checksum files
func NewMockChecksumServer(t *testing.T, checksum, filename string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s  %s\n", checksum, filename)
	}))
}

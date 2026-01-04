package vncproxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"
)

// Mock WebSocket adapter for testing
type MockWebSocketAdapter struct {
	readData  []byte
	writeData bytes.Buffer
	closed    bool
	addr      string
	request   *http.Request
}

func NewMockWebSocketAdapter(readData []byte, addr string, req *http.Request) *MockWebSocketAdapter {
	return &MockWebSocketAdapter{
		readData: readData,
		addr:     addr,
		request:  req,
	}
}

func (m *MockWebSocketAdapter) Read(p []byte) (n int, err error) {
	if len(m.readData) == 0 {
		return 0, io.EOF
	}
	n = copy(p, m.readData)
	m.readData = m.readData[n:]
	return n, nil
}

func (m *MockWebSocketAdapter) Write(p []byte) (n int, err error) {
	return m.writeData.Write(p)
}

func (m *MockWebSocketAdapter) Close() error {
	m.closed = true
	return nil
}

func (m *MockWebSocketAdapter) RemoteAddr() string {
	return m.addr
}

func (m *MockWebSocketAdapter) Request() *http.Request {
	return m.request
}

func (m *MockWebSocketAdapter) SetBinaryMode() error {
	return nil
}

func TestMockWebSocketAdapter(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	testData := []byte("test data")
	
	adapter := NewMockWebSocketAdapter(testData, "127.0.0.1:8080", req)
	
	// Test Read
	buf := make([]byte, 20)
	n, err := adapter.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(testData) {
		t.Fatalf("Read returned %d bytes, expected %d", n, len(testData))
	}
	if !bytes.Equal(buf[:n], testData) {
		t.Fatalf("Read data mismatch: got %v, expected %v", buf[:n], testData)
	}
	
	// Test Write
	writeData := []byte("write test")
	n, err = adapter.Write(writeData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(writeData) {
		t.Fatalf("Write returned %d bytes, expected %d", n, len(writeData))
	}
	if !bytes.Equal(adapter.writeData.Bytes(), writeData) {
		t.Fatalf("Write data mismatch: got %v, expected %v", adapter.writeData.Bytes(), writeData)
	}
	
	// Test RemoteAddr
	if adapter.RemoteAddr() != "127.0.0.1:8080" {
		t.Fatalf("RemoteAddr mismatch: got %s, expected 127.0.0.1:8080", adapter.RemoteAddr())
	}
	
	// Test Request
	if adapter.Request() != req {
		t.Fatalf("Request mismatch")
	}
	
	// Test SetBinaryMode
	if err := adapter.SetBinaryMode(); err != nil {
		t.Fatalf("SetBinaryMode failed: %v", err)
	}
	
	// Test Close
	if err := adapter.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if !adapter.closed {
		t.Fatalf("Adapter not marked as closed")
	}
}

func TestNetWebSocketAdapter(t *testing.T) {
	// Create a test server to establish a WebSocket connection
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		adapter := NewNetWebSocketAdapter(ws)
		
		// Test SetBinaryMode
		if err := adapter.SetBinaryMode(); err != nil {
			t.Errorf("SetBinaryMode failed: %v", err)
			return
		}
		
		// Test Read/Write by echoing data
		buf := make([]byte, 1024)
		n, err := adapter.Read(buf)
		if err != nil {
			t.Errorf("Read failed: %v", err)
			return
		}
		
		_, err = adapter.Write(buf[:n])
		if err != nil {
			t.Errorf("Write failed: %v", err)
			return
		}
	}))
	defer server.Close()
	
	// Connect to the test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()
	
	// Send test data
	testData := []byte("hello websocket adapter")
	if _, err := ws.Write(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	
	// Read echoed data
	buf := make([]byte, 1024)
	n, err := ws.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	
	if !bytes.Equal(buf[:n], testData) {
		t.Fatalf("Echoed data mismatch: got %v, expected %v", buf[:n], testData)
	}
}

func TestProxyWithMockAdapter(t *testing.T) {
	// Create a proxy with default configuration
	proxy := New(&Config{
		LogLevel: InfoLevel,
		TokenHandler: func(r *http.Request) (addr string, err error) {
			return "127.0.0.1:5901", nil
		},
	})
	
	// Create a mock adapter
	req := httptest.NewRequest("GET", "/test", nil)
	testData := []byte("vnc test data")
	adapter := NewMockWebSocketAdapter(testData, "127.0.0.1:8080", req)
	
	// Test that the proxy can use the WebSocket handler function
	handler := proxy.Handler()
	if handler == nil {
		t.Fatalf("Handler should not be nil")
	}
	
	// Verify adapter interface compliance
	if adapter.RemoteAddr() != "127.0.0.1:8080" {
		t.Fatalf("Adapter RemoteAddr mismatch")
	}
	
	if adapter.Request() != req {
		t.Fatalf("Adapter Request mismatch")
	}
	
	// This would normally connect to a VNC server, but we're just testing the interface
	// The actual connection will fail, but we can verify the adapter interface works
	// handler(adapter) // We skip this as it would try to connect to a real VNC server
}
package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProxy_Valid(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType ProxyType
		expectedAddr string
		expectedPort int
	}{
		{
			name:         "http with scheme",
			input:        "http://127.0.0.1:8080",
			expectedType: HTTP,
			expectedAddr: "127.0.0.1",
			expectedPort: 8080,
		},
		{
			name:         "https with scheme",
			input:        "https://192.168.1.1:443",
			expectedType: HTTPS,
			expectedAddr: "192.168.1.1",
			expectedPort: 443,
		},
		{
			name:         "socks5 with scheme",
			input:        "socks5://10.0.0.1:1080",
			expectedType: SOCKS5,
			expectedAddr: "10.0.0.1",
			expectedPort: 1080,
		},
		{
			name:         "without scheme defaults to http",
			input:        "127.0.0.1:8080",
			expectedType: HTTP,
			expectedAddr: "127.0.0.1",
			expectedPort: 8080,
		},
		{
			name:         "with whitespace",
			input:        "  http://127.0.0.1:8080  ",
			expectedType: HTTP,
			expectedAddr: "127.0.0.1",
			expectedPort: 8080,
		},
		{
			name:         "domain address",
			input:        "http://proxy.example.com:8080",
			expectedType: HTTP,
			expectedAddr: "proxy.example.com",
			expectedPort: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxy(tt.input)
			assert.NoError(t, err)
			assert.NotNil(t, proxy)
			assert.Equal(t, tt.expectedType, proxy.Type)
			assert.Equal(t, tt.expectedAddr, proxy.Address)
			assert.Equal(t, tt.expectedPort, proxy.Port)
		})
	}
}

func TestParseProxy_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   ",
		},
		{
			name:  "no port",
			input: "http://127.0.0.1",
		},
		{
			name:  "invalid format",
			input: "not-a-proxy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxy(tt.input)
			assert.Error(t, err)
			assert.Nil(t, proxy)
		})
	}
}

func TestProxy_URL(t *testing.T) {
	tests := []struct {
		name     string
		proxy    Proxy
		expected string
	}{
		{
			name:     "http proxy",
			proxy:    Proxy{Type: HTTP, Address: "127.0.0.1", Port: 8080},
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "https proxy",
			proxy:    Proxy{Type: HTTPS, Address: "192.168.1.1", Port: 443},
			expected: "https://192.168.1.1:443",
		},
		{
			name:     "socks5 proxy",
			proxy:    Proxy{Type: SOCKS5, Address: "10.0.0.1", Port: 1080},
			expected: "socks5://10.0.0.1:1080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.proxy.URL()
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestProxyHandler_Next(t *testing.T) {
	proxies := []Proxy{
		{Type: HTTP, Address: "127.0.0.1", Port: 8080},
		{Type: HTTP, Address: "127.0.0.2", Port: 8081},
		{Type: HTTP, Address: "127.0.0.3", Port: 8082},
	}

	handler := NewProxyHandler(proxies)

	seen := make(map[string]bool)
	for i := 0; i < 6; i++ {
		proxy := handler.Next()
		assert.NotNil(t, proxy)
		key := proxy.URL()
		seen[key] = true
	}

	assert.Len(t, seen, 3, "Should cycle through all proxies")

	for _, p := range proxies {
		assert.True(t, seen[p.URL()], "Proxy %s should be seen", p.URL())
	}
}

func TestProxyHandler_Next_Empty(t *testing.T) {
	handler := NewProxyHandler(nil)
	proxy := handler.Next()
	assert.Nil(t, proxy)

	handler = NewProxyHandler([]Proxy{})
	proxy = handler.Next()
	assert.Nil(t, proxy)
}

func TestProxyHandler_GetProxyURL(t *testing.T) {
	proxies := []Proxy{
		{Type: HTTP, Address: "127.0.0.1", Port: 8080},
	}

	handler := NewProxyHandler(proxies)

	url := handler.GetProxyURL()
	assert.Equal(t, "http://127.0.0.1:8080", url)

	url = handler.GetProxyURL()
	assert.Equal(t, "http://127.0.0.1:8080", url)
}

func TestProxyHandler_GetProxyURL_Empty(t *testing.T) {
	handler := NewProxyHandler(nil)
	url := handler.GetProxyURL()
	assert.Empty(t, url)

	handler = NewProxyHandler([]Proxy{})
	url = handler.GetProxyURL()
	assert.Empty(t, url)
}

func TestProxyURLToURL(t *testing.T) {
	tests := []struct {
		name     string
		proxyURL string
		wantNil  bool
	}{
		{
			name:     "valid http url",
			proxyURL: "http://127.0.0.1:8080",
			wantNil:  false,
		},
		{
			name:     "empty url",
			proxyURL: "",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := ProxyURLToURL(tt.proxyURL)
			if tt.wantNil {
				assert.Nil(t, u)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, u)
			}
		})
	}
}

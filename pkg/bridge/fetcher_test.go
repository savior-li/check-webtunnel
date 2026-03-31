package bridge

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetcher_NewFetcher(t *testing.T) {
	fetcher := NewFetcher("https://example.com", 10*time.Second)
	assert.NotNil(t, fetcher)
	assert.Equal(t, "https://example.com", fetcher.url)
	assert.Equal(t, 10*time.Second, fetcher.timeout)
}

func TestFetcher_SetProxy(t *testing.T) {
	tests := []struct {
		name     string
		proxyStr string
		wantErr  bool
	}{
		{
			name:     "empty proxy",
			proxyStr: "",
			wantErr:  false,
		},
		{
			name:     "http proxy",
			proxyStr: "http://127.0.0.1:8080",
			wantErr:  false,
		},
		{
			name:     "https proxy",
			proxyStr: "https://127.0.0.1:8080",
			wantErr:  false,
		},
		{
			name:     "socks5 proxy",
			proxyStr: "socks5://127.0.0.1:1080",
			wantErr:  false,
		},
		{
			name:     "proxy without scheme",
			proxyStr: "127.0.0.1:8080",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewFetcher("https://example.com", 5*time.Second)
			err := fetcher.SetProxy(tt.proxyStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFetcher_Fetch_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
			<html>
				<body>
					webtunnel 192.168.1.1:443 ABCD1234 url=https://example.com <br/>
					webtunnel [2001:db8::1]:443 EFGH5678 url=https://test.com <br/>
				</body>
			</html>
		`
		w.Write([]byte(html))
	}))
	defer ts.Close()

	fetcher := NewFetcher(ts.URL, 5*time.Second)
	bridges, err := fetcher.Fetch()

	assert.NoError(t, err)
	assert.Len(t, bridges, 2)
	assert.Equal(t, "192.168.1.1", bridges[0].Address)
	assert.Equal(t, 443, bridges[0].Port)
	assert.Equal(t, "ABCD1234", bridges[0].Fingerprint)
}

func TestFetcher_Fetch_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	fetcher := NewFetcher(ts.URL, 5*time.Second)
	bridges, err := fetcher.Fetch()

	assert.Error(t, err)
	assert.Nil(t, bridges)
	assert.Contains(t, err.Error(), "unexpected status code: 404")
}

func TestFetcher_Fetch_NetworkError(t *testing.T) {
	fetcher := NewFetcher("http://127.0.0.1:99999", 1*time.Second)
	bridges, err := fetcher.Fetch()

	assert.Error(t, err)
	assert.Nil(t, bridges)
}

func TestFetcher_Parse(t *testing.T) {
	html := `
		<html>
			<body>
				webtunnel 192.168.1.1:443 ABCD1234 url=https://test1.com <br/>
				webtunnel 10.0.0.1:8080 EFGH5678 url=https://test2.com <br/>
				webtunnel [2001:db8::1]:9000 ABCDEF url=https://test3.com <br/>
			</body>
		</html>
	`

	fetcher := NewFetcher("https://example.com", 5*time.Second)
	bridges, err := fetcher.Parse(html)

	assert.NoError(t, err)
	assert.Len(t, bridges, 3)

	assert.Equal(t, "webtunnel", bridges[0].Transport)
	assert.Equal(t, "192.168.1.1", bridges[0].Address)
	assert.Equal(t, 443, bridges[0].Port)
	assert.Equal(t, "ABCD1234", bridges[0].Fingerprint)

	assert.Equal(t, "10.0.0.1", bridges[1].Address)
	assert.Equal(t, 8080, bridges[1].Port)

	assert.Equal(t, "2001:db8::1", bridges[2].Address)
	assert.Equal(t, 9000, bridges[2].Port)
	assert.Equal(t, "ABCDEF", bridges[2].Fingerprint)
}

func TestFetcher_Parse_Empty(t *testing.T) {
	fetcher := NewFetcher("https://example.com", 5*time.Second)

	tests := []struct {
		name string
		html string
	}{
		{"empty html", ""},
		{"no bridges", "<html><body>No bridges here</body></html>"},
		{"invalid format", "<html><body>invalid bridge format</body></html>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridges, err := fetcher.Parse(tt.html)
			assert.NoError(t, err)
			assert.Len(t, bridges, 0)
		})
	}
}

func TestFetcher_ParseVariations(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []struct {
			addr string
			port int
			fp   string
		}
	}{
		{
			name: "IPv4 webtunnel",
			html: "webtunnel 192.168.1.1:443 ABCD url=https://test.com",
			expected: []struct {
				addr string
				port int
				fp   string
			}{{addr: "192.168.1.1", port: 443, fp: "ABCD"}},
		},
		{
			name: "IPv6 webtunnel",
			html: "webtunnel [2001:db8::1]:443 ABCD url=https://test.com",
			expected: []struct {
				addr string
				port int
				fp   string
			}{{addr: "2001:db8::1", port: 443, fp: "ABCD"}},
		},
		{
			name: "multiple bridges",
			html: strings.Join([]string{
				"webtunnel 192.168.1.1:443 ABCD url=https://test1.com",
				"webtunnel 10.0.0.1:8080 EFGH url=https://test2.com",
				"webtunnel [::1]:9000 1234 url=https://test3.com",
			}, "\n"),
			expected: []struct {
				addr string
				port int
				fp   string
			}{
				{addr: "192.168.1.1", port: 443, fp: "ABCD"},
				{addr: "10.0.0.1", port: 8080, fp: "EFGH"},
				{addr: "::1", port: 9000, fp: "1234"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewFetcher("https://example.com", 5*time.Second)
			bridges, err := fetcher.Parse(tt.html)
			assert.NoError(t, err)
			assert.Len(t, bridges, len(tt.expected))

			for i, exp := range tt.expected {
				assert.Equal(t, exp.addr, bridges[i].Address)
				assert.Equal(t, exp.port, bridges[i].Port)
				assert.Equal(t, exp.fp, bridges[i].Fingerprint)
			}
		})
	}
}

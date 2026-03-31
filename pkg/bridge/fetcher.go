package bridge

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type Fetcher struct {
	url     string
	timeout time.Duration
	client  *http.Client
}

func NewFetcher(url string, timeout time.Duration) *Fetcher {
	return &Fetcher{
		url:     url,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout * time.Second,
		},
	}
}

func (f *Fetcher) SetProxy(proxyURL string) error {
	if proxyURL == "" {
		f.client.Transport = nil
		return nil
	}

	proxyURL = strings.TrimPrefix(proxyURL, "http://")
	proxyURL = strings.TrimPrefix(proxyURL, "https://")
	socks5Prefix := strings.TrimPrefix(proxyURL, "socks5://")

	isSocks5 := socks5Prefix != proxyURL
	proxyURL = socks5Prefix

	if !strings.Contains(proxyURL, "://") {
		if isSocks5 {
			proxyURL = "socks5://" + proxyURL
		} else {
			proxyURL = "http://" + proxyURL
		}
	}

	if isSocks5 {
		return f.setSocks5Proxy(proxyURL)
	}

	proxyURLObj, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	f.client.Transport = &http.Transport{
		Proxy:               http.ProxyURL(proxyURLObj),
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 30 * time.Second,
	}

	return nil
}

func (f *Fetcher) setSocks5Proxy(proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid socks5 proxy URL: %w", err)
	}

	var auth *proxy.Auth
	if parsedURL.User != nil {
		if p, ok := parsedURL.User.Password(); ok {
			auth = &proxy.Auth{
				User:     parsedURL.User.Username(),
				Password: p,
			}
		} else {
			auth = &proxy.Auth{
				User: parsedURL.User.Username(),
			}
		}
	}

	dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, &net.Dialer{
		Timeout:  f.timeout,
		Deadline: time.Now().Add(f.timeout),
	})
	if err != nil {
		return fmt.Errorf("create SOCKS5 dialer failed: %w", err)
	}

	f.client.Transport = &http.Transport{
		Dial:                dialer.Dial,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 30 * time.Second,
	}

	return nil
}

func (f *Fetcher) Fetch() ([]Bridge, error) {
	resp, err := f.client.Get(f.url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	return f.Parse(string(body))
}

func (f *Fetcher) Parse(html string) ([]Bridge, error) {
	var bridges []Bridge

	lines := strings.Split(html, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "webtunnel") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// parts[0] = "webtunnel"
		// parts[1] = "192.168.1.1:443" 或 "[IPv6]:port"
		// parts[2] = fingerprint
		// parts[3+] = url=xxx, ver=yyy

		addrPort := parts[1]
		fingerprint := parts[2]

		var addr string
		var port int

		if strings.HasPrefix(addrPort, "[") {
			// IPv6 格式: [addr]:port
			end := strings.LastIndex(addrPort, "]")
			if end > 0 {
				addr = addrPort[1:end]
				fmt.Sscanf(addrPort[end+2:], "%d", &port)
			}
		} else {
			// IPv4 格式: addr:port
			lastColon := strings.LastIndex(addrPort, ":")
			if lastColon > 0 {
				addr = addrPort[:lastColon]
				fmt.Sscanf(addrPort[lastColon+1:], "%d", &port)
			}
		}

		if addr == "" || port == 0 {
			continue
		}

		// 提取 url（如果存在）
		var extra string
		for i := 3; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "url=") {
				extra = strings.TrimPrefix(parts[i], "url=")
				extra = strings.TrimSuffix(extra, "<br/>")
				break
			}
		}

		bridge := &Bridge{
			Transport:    "webtunnel",
			Address:      addr,
			Port:         port,
			Fingerprint:  fingerprint,
			Extra:        extra,
			DiscoveredAt: time.Now(),
			IsAvailable:  -1,
		}
		bridge.Hash = bridge.GenerateHash()

		bridges = append(bridges, *bridge)
	}

	return bridges, nil
}

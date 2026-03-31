package bridge

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
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
		Proxy: http.ProxyURL(proxyURLObj),
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
		Timeout: f.timeout,
	})
	if err != nil {
		return fmt.Errorf("create SOCKS5 dialer failed: %w", err)
	}

	f.client.Transport = &http.Transport{
		Dial: dialer.Dial,
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

	bridgeRegex := regexp.MustCompile(`Bridge\s+(webtunnel\s+)?([^\s]+):(\d+)(?:\s+fingerprint\s+([^\s]+))?`)
	matches := bridgeRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		transport := "webtunnel"
		address := match[2]
		port := 0
		fingerprint := ""

		fmt.Sscanf(match[3], "%d", &port)
		if len(match) > 4 {
			fingerprint = match[4]
		}

		bridge := &Bridge{
			Transport:    transport,
			Address:      address,
			Port:         port,
			Fingerprint:  fingerprint,
			DiscoveredAt: time.Now(),
			IsAvailable:  -1,
		}
		bridge.Hash = bridge.GenerateHash()

		bridges = append(bridges, *bridge)
	}

	return bridges, nil
}

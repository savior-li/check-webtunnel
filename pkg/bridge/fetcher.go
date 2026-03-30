package bridge

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
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

	transport := &http.Transport{
		Proxy: http.ProxyURL(nil),
	}

	proxyURL = strings.TrimPrefix(proxyURL, "http://")
	proxyURL = strings.TrimPrefix(proxyURL, "https://")
	proxyURL = strings.TrimPrefix(proxyURL, "socks5://")

	if !strings.Contains(proxyURL, "://") {
		proxyURL = "http://" + proxyURL
	}

	transport.Proxy = nil
	f.client.Transport = transport

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

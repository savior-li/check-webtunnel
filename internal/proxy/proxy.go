package proxy

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type ProxyClient struct {
	dialer proxy.Dialer
}

func NewProxyClient(proxyType, address string, port int, username, password string) (*ProxyClient, error) {
	proxyURL := fmt.Sprintf("%s://%s:%d", proxyType, address, port)
	if username != "" && password != "" {
		proxyURL = fmt.Sprintf("%s://%s:%s@%s:%d", proxyType, username, password, address, port)
	}

	var dialer proxy.Dialer
	var err error

	switch strings.ToLower(proxyType) {
	case "socks5", "socks":
		dialer, err = proxy.SOCKS5("tcp", fmt.Sprintf("%s:%d", address, port), nil, proxy.Direct)
	case "http", "https":
		baseDialer := &net.Dialer{
			Timeout: 10 * time.Second,
		}
		dialer, err = proxy.FromURL(parseProxyURL(proxyURL), baseDialer)
	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", proxyType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create proxy dialer: %w", err)
	}

	return &ProxyClient{dialer: dialer}, nil
}

func parseProxyURL(proxyURL string) *url.URL {
	u, _ := url.Parse(proxyURL)
	return u
}

func (p *ProxyClient) Dial(network, addr string) (net.Conn, error) {
	return p.dialer.Dial(network, addr)
}

func ValidateProxy(proxyType, address string, port int, username, password string) error {
	client, err := NewProxyClient(proxyType, address, port, username, password)
	if err != nil {
		return err
	}

	conn, err := client.Dial("tcp", "1.1.1.1:80")
	if err != nil {
		return fmt.Errorf("proxy connection failed: %w", err)
	}
	conn.Close()

	return nil
}

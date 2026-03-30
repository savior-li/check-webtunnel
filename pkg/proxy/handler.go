package proxy

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

type ProxyType string

const (
	HTTP    ProxyType = "http"
	HTTPS   ProxyType = "https"
	SOCKS5  ProxyType = "socks5"
	SOCKS5H ProxyType = "socks5h"
)

type Proxy struct {
	Type    ProxyType
	Address string
	Port    int
}

func (p *Proxy) URL() string {
	return fmt.Sprintf("%s://%s:%d", p.Type, p.Address, p.Port)
}

func ParseProxy(proxyStr string) (*Proxy, error) {
	proxyStr = strings.TrimSpace(proxyStr)
	if proxyStr == "" {
		return nil, fmt.Errorf("empty proxy string")
	}

	var p Proxy

	if strings.HasPrefix(proxyStr, "socks5h://") {
		p.Type = SOCKS5H
		proxyStr = strings.TrimPrefix(proxyStr, "socks5h://")
	} else if strings.HasPrefix(proxyStr, "socks5://") {
		p.Type = SOCKS5
		proxyStr = strings.TrimPrefix(proxyStr, "socks5://")
	} else if strings.HasPrefix(proxyStr, "socks://") {
		p.Type = SOCKS5
		proxyStr = strings.TrimPrefix(proxyStr, "socks://")
	} else if strings.HasPrefix(proxyStr, "https://") {
		p.Type = HTTPS
		proxyStr = strings.TrimPrefix(proxyStr, "https://")
	} else if strings.HasPrefix(proxyStr, "http://") {
		p.Type = HTTP
		proxyStr = strings.TrimPrefix(proxyStr, "http://")
	} else {
		p.Type = HTTP
	}

	parts := strings.Split(proxyStr, ":")
	if len(parts) >= 2 {
		p.Address = parts[0]
		fmt.Sscanf(parts[1], "%d", &p.Port)
	} else {
		return nil, fmt.Errorf("invalid proxy format")
	}

	return &p, nil
}

type ProxyHandler struct {
	proxies []Proxy
	index   int
	mu      sync.Mutex
}

func NewProxyHandler(proxies []Proxy) *ProxyHandler {
	return &ProxyHandler{
		proxies: proxies,
		index:   0,
	}
}

func (h *ProxyHandler) Next() *Proxy {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.proxies) == 0 {
		return nil
	}

	proxy := h.proxies[h.index]
	h.index = (h.index + 1) % len(h.proxies)
	return &proxy
}

func (h *ProxyHandler) GetProxyURL() string {
	proxy := h.Next()
	if proxy == nil {
		return ""
	}
	return proxy.URL()
}

func ProxyURLToURL(proxyURL string) (*url.URL, error) {
	if proxyURL == "" {
		return nil, nil
	}
	return url.Parse(proxyURL)
}

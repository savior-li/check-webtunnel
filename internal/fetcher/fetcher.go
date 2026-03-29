package fetcher

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"tor-bridge-collector/pkg/models"
)

type Fetcher struct {
	client  *http.Client
	baseURL string
}

type BridgeData struct {
	Address   string
	Port      int
	Transport string
	Hash      string
	IPv6      string
}

func New(timeout int) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		baseURL: "https://bridges.torproject.org/bridges?transport=webtunnel",
	}
}

func NewWithProxy(proxyURL string, timeout int) (*Fetcher, error) {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	return &Fetcher{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(proxy),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		baseURL: "https://bridges.torproject.org/bridges?transport=webtunnel",
	}, nil
}

func (f *Fetcher) Fetch() ([]models.Bridge, error) {
	resp, err := f.client.Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return f.parseBridges(string(body))
}

func (f *Fetcher) parseBridges(html string) ([]models.Bridge, error) {
	var bridges []models.Bridge
	now := time.Now()

	bridgeRegex := regexp.MustCompile(`Bridge\s+(\w+)\s+([^\s]+):(\d+)\s+([a-zA-Z0-9]+)`)
	matches := bridgeRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) != 5 {
			continue
		}

		transport := match[1]
		address := match[2]
		port := 443
		fmt.Sscanf(match[3], "%d", &port)
		hash := match[4]

		bridge := models.Bridge{
			Transport: transport,
			Address:   address,
			Port:      port,
			Hash:      hash,
			FirstSeen: now,
			LastSeen:  now,
		}
		bridge.Hash = bridge.CalculateHash()
		bridges = append(bridges, bridge)
	}

	ipv6Regex := regexp.MustCompile(`Bridge\s+(\w+)\s+([^\s]+):(\d+)\s+([a-zA-Z0-9]+)\s+\[([0-9a-f:]+)\]`)
	ipv6Matches := ipv6Regex.FindAllStringSubmatch(html, -1)

	for _, match := range ipv6Matches {
		if len(match) != 6 {
			continue
		}

		for i := range bridges {
			if bridges[i].Address == match[2] && bridges[i].Hash == match[4] {
				bridges[i].IPv6 = match[5]
				break
			}
		}
	}

	return bridges, nil
}

func (f *Fetcher) FetchWithRetry(retry int) ([]models.Bridge, error) {
	var lastErr error
	for i := 0; i <= retry; i++ {
		bridges, err := f.Fetch()
		if err == nil {
			return bridges, nil
		}
		lastErr = err
		time.Sleep(time.Duration(2<<uint(i)) * time.Second)
	}
	return nil, fmt.Errorf("fetch failed after %d retries: %w", retry, lastErr)
}

func BridgesToLines(bridges []models.Bridge) []string {
	var lines []string
	for _, b := range bridges {
		line := b.ToTorrcLine()
		lines = append(lines, line)
	}
	return lines
}

func FormatAsTorrc(bridges []models.Bridge) string {
	var buf bytes.Buffer
	for _, b := range bridges {
		buf.WriteString(b.ToTorrcLine())
		buf.WriteString("\n")
	}
	return buf.String()
}

func FormatAsJSON(bridges []models.Bridge) string {
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, b := range bridges {
		data := b.ToJSON()
		buf.WriteString(fmt.Sprintf("  {\"hash\": \"%s\", \"transport\": \"%s\", \"address\": \"%s\", \"port\": %d, \"is_valid\": %v}",
			data["hash"], data["transport"], data["address"], data["port"], data["is_valid"]))
		if i < len(bridges)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("]\n")
	return buf.String()
}

func FormatAsCSV(bridges []models.Bridge) string {
	var buf bytes.Buffer
	buf.WriteString("hash,transport,address,port,is_valid,avg_latency,success_rate\n")
	for _, b := range bridges {
		buf.WriteString(fmt.Sprintf("%s,%s,%s,%d,%v,%.2f,%.2f\n",
			b.Hash, b.Transport, b.Address, b.Port, b.IsValid, b.AvgLatency, b.SuccessRate))
	}
	return buf.String()
}

func FilterNewBridges(existing, fetched []models.Bridge) []models.Bridge {
	existingMap := make(map[string]bool)
	for _, b := range existing {
		existingMap[b.Hash] = true
	}

	var newBridges []models.Bridge
	for _, b := range fetched {
		if !existingMap[b.Hash] {
			newBridges = append(newBridges, b)
		}
	}
	return newBridges
}

func SanitizeAddress(address string) string {
	address = strings.TrimSpace(address)
	address = strings.TrimPrefix(address, "https://")
	address = strings.TrimPrefix(address, "http://")
	if idx := strings.Index(address, "/"); idx != -1 {
		address = address[:idx]
	}
	return address
}

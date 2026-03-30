package bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Bridge struct {
	ID           int64     `json:"id"`
	Hash         string    `json:"hash"`
	Transport    string    `json:"transport"`
	Address      string    `json:"address"`
	Port         int       `json:"port"`
	Fingerprint  string    `json:"fingerprint,omitempty"`
	DiscoveredAt time.Time `json:"discovered_at"`
	LastValidAt  time.Time `json:"last_validated_at,omitempty"`
	IsAvailable  int       `json:"is_available"`
	ResponseTime int       `json:"response_time_ms,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BridgeResponse struct {
	Bridges   []Bridge  `json:"bridges"`
	FetchedAt time.Time `json:"fetched_at"`
}

func (b *Bridge) GenerateHash() string {
	data := fmt.Sprintf("%s:%d:%s", b.Address, b.Port, b.Transport)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func ParseBridgeLine(line string) (*Bridge, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	var address string
	var port int
	var fingerprint string

	parts := strings.Fields(line)

	if len(parts) >= 2 {
		addrPort := strings.Split(parts[0], ":")
		if len(addrPort) == 2 {
			address = addrPort[0]
			fmt.Sscanf(addrPort[1], "%d", &port)
		}
	}

	if len(parts) >= 3 && strings.HasPrefix(parts[2], " fingerprint ") {
		fingerprint = strings.Trim(parts[2], " fingerprint ")
	} else if len(parts) >= 3 {
		fingerprint = parts[2]
	}

	transport := "webtunnel"
	if len(parts) >= 1 && parts[0] == "webtunnel" {
		transport = "webtunnel"
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

	return bridge, nil
}

func FormatTorrcLine(bridge *Bridge) string {
	if bridge.Fingerprint != "" {
		return fmt.Sprintf("Bridge %s %s:%d fingerprint=%s",
			bridge.Transport, bridge.Address, bridge.Port, bridge.Fingerprint)
	}
	return fmt.Sprintf("Bridge %s %s:%d",
		bridge.Transport, bridge.Address, bridge.Port)
}

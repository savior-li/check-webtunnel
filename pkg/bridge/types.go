package bridge

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
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

type FileImporter struct {
	importer interface {
		Parse(filePath string) ([]Bridge, error)
	}
}

func NewFileImporter(format, transport string) *FileImporter {
	var impl interface {
		Parse(filePath string) ([]Bridge, error)
	}

	switch strings.ToLower(format) {
	case "csv":
		impl = &csvImporterAdapter{}
	default:
		impl = &textImporterAdapter{transport: transport}
	}

	return &FileImporter{importer: impl}
}

func (fi *FileImporter) Import(filePath string) ([]Bridge, error) {
	return fi.importer.Parse(filePath)
}

type textImporterAdapter struct {
	transport string
}

func (t *textImporterAdapter) Parse(filePath string) ([]Bridge, error) {
	lines, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var bridges []Bridge
	for _, line := range strings.Split(string(lines), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		var address string
		var port int
		var fingerprint string
		transport := t.transport

		if strings.HasPrefix(parts[0], "Bridge") && len(parts) >= 3 {
			if len(parts) >= 2 {
				transport = parts[1]
			}
			addrPort := strings.Split(parts[2], ":")
			if len(addrPort) == 2 {
				address = addrPort[0]
				fmt.Sscanf(addrPort[1], "%d", &port)
			}
			for i := 3; i < len(parts); i++ {
				if strings.HasPrefix(parts[i], "fingerprint=") {
					fingerprint = strings.TrimPrefix(parts[i], "fingerprint=")
				}
			}
		} else {
			addrPort := strings.Split(parts[0], ":")
			if len(addrPort) == 2 {
				address = addrPort[0]
				fmt.Sscanf(addrPort[1], "%d", &port)
			}
			if len(parts) >= 2 {
				fingerprint = parts[1]
			}
		}

		if address == "" || port == 0 {
			continue
		}

		b := Bridge{
			Transport:    transport,
			Address:      address,
			Port:         port,
			Fingerprint:  fingerprint,
			DiscoveredAt: time.Now(),
			IsAvailable:  -1,
		}
		b.Hash = b.GenerateHash()
		bridges = append(bridges, b)
	}

	return bridges, nil
}

type csvImporterAdapter struct{}

func (c *csvImporterAdapter) Parse(filePath string) ([]Bridge, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("csv file is empty or has no data rows")
	}

	header := make(map[string]int)
	for i, col := range records[0] {
		header[strings.TrimSpace(strings.ToLower(col))] = i
	}

	transportIdx, ok := header["transport"]
	if !ok {
		transportIdx = -1
	}
	addrIdx, ok := header["address"]
	if !ok {
		return nil, fmt.Errorf("csv missing required column: address")
	}
	portIdx, ok := header["port"]
	if !ok {
		return nil, fmt.Errorf("csv missing required column: port")
	}
	fingerprintIdx, ok := header["fingerprint"]
	if !ok {
		fingerprintIdx = -1
	}

	var bridges []Bridge
	for _, record := range records[1:] {
		if len(record) == 0 {
			continue
		}

		transport := "webtunnel"
		if transportIdx >= 0 && transportIdx < len(record) {
			transport = strings.TrimSpace(record[transportIdx])
		}

		address := strings.TrimSpace(record[addrIdx])
		portStr := strings.TrimSpace(record[portIdx])
		var port int
		fmt.Sscanf(portStr, "%d", &port)

		fingerprint := ""
		if fingerprintIdx >= 0 && fingerprintIdx < len(record) {
			fingerprint = strings.TrimSpace(record[fingerprintIdx])
		}

		if address == "" || port == 0 {
			continue
		}

		b := Bridge{
			Transport:    transport,
			Address:      address,
			Port:         port,
			Fingerprint:  fingerprint,
			DiscoveredAt: time.Now(),
			IsAvailable:  -1,
		}
		b.Hash = b.GenerateHash()
		bridges = append(bridges, b)
	}

	return bridges, nil
}

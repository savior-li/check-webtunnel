package bridge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTextImporter_Parse(t *testing.T) {
	tmpDir := t.TempDir()

	textContent := `192.168.1.1:443 ABC123
192.168.1.2:443 DEF456
# Comment line
Bridge webtunnel 10.0.0.1:443 fingerprint=GHI789
`

	textFile := filepath.Join(tmpDir, "bridges.txt")
	err := os.WriteFile(textFile, []byte(textContent), 0644)
	if err != nil {
		t.Fatalf("create test file failed: %v", err)
	}

	importer := NewFileImporter("text", "webtunnel")
	bridges, err := importer.Import(textFile)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if len(bridges) != 3 {
		t.Errorf("expected 3 bridges, got %d", len(bridges))
	}
}

func TestCSVImporter_Parse(t *testing.T) {
	tmpDir := t.TempDir()

	csvContent := `transport,address,port,fingerprint
webtunnel,192.168.1.1,443,ABC123
webtunnel,192.168.1.2,443,DEF456
`

	csvFile := filepath.Join(tmpDir, "bridges.csv")
	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("create test file failed: %v", err)
	}

	importer := NewFileImporter("csv", "webtunnel")
	bridges, err := importer.Import(csvFile)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if len(bridges) != 2 {
		t.Errorf("expected 2 bridges, got %d", len(bridges))
	}

	if bridges[0].Transport != "webtunnel" {
		t.Errorf("expected transport webtunnel, got %s", bridges[0].Transport)
	}

	if bridges[0].Address != "192.168.1.1" {
		t.Errorf("expected address 192.168.1.1, got %s", bridges[0].Address)
	}

	if bridges[0].Port != 443 {
		t.Errorf("expected port 443, got %d", bridges[0].Port)
	}
}

func TestBridge_GenerateHash(t *testing.T) {
	b := Bridge{
		Address:   "192.168.1.1",
		Port:      443,
		Transport: "webtunnel",
	}

	hash := b.GenerateHash()
	if len(hash) != 32 {
		t.Errorf("expected hash length 32, got %d", len(hash))
	}

	b2 := Bridge{
		Address:   "192.168.1.1",
		Port:      443,
		Transport: "webtunnel",
	}
	hash2 := b2.GenerateHash()

	if hash != hash2 {
		t.Errorf("same bridge should have same hash")
	}

	b3 := Bridge{
		Address:   "192.168.1.2",
		Port:      443,
		Transport: "webtunnel",
	}
	hash3 := b3.GenerateHash()

	if hash == hash3 {
		t.Errorf("different bridges should have different hash")
	}
}

func TestParseBridgeLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{
			name:    "simple format",
			line:    "192.168.1.1:443 ABC123",
			wantErr: false,
		},
		{
			name:    "torrc format",
			line:    "Bridge webtunnel 192.168.1.1:443 fingerprint=ABC123",
			wantErr: false,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseBridgeLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBridgeLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

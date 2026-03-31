package bridge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBridge_GenerateHash(t *testing.T) {
	bridge1 := &Bridge{
		Address:   "192.168.1.1",
		Port:      443,
		Transport: "webtunnel",
	}
	bridge2 := &Bridge{
		Address:   "192.168.1.1",
		Port:      443,
		Transport: "webtunnel",
	}
	bridge3 := &Bridge{
		Address:   "192.168.1.2",
		Port:      443,
		Transport: "webtunnel",
	}

	hash1 := bridge1.GenerateHash()
	hash2 := bridge2.GenerateHash()
	hash3 := bridge3.GenerateHash()

	assert.Equal(t, hash1, hash2, "Same input should produce same hash")
	assert.NotEqual(t, hash1, hash3, "Different input should produce different hash")
	assert.Len(t, hash1, 32, "Hash should be 32 characters (16 bytes hex)")
}

func TestParseBridgeLine_Valid(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedAddr string
		expectedPort int
		expectedFp   string
	}{
		{
			name:         "address with port and fingerprint",
			input:        "192.168.1.1:443 fingerprint ABCD1234",
			expectedAddr: "192.168.1.1",
			expectedPort: 443,
			expectedFp:   "ABCD1234",
		},
		{
			name:         "single token not parsed",
			input:        "10.0.0.1:8080",
			expectedAddr: "",
			expectedPort: 0,
			expectedFp:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge, err := ParseBridgeLine(tt.input)
			assert.NoError(t, err)
			assert.NotNil(t, bridge)
			assert.Equal(t, tt.expectedAddr, bridge.Address)
			assert.Equal(t, tt.expectedPort, bridge.Port)
			assert.Equal(t, tt.expectedFp, bridge.Fingerprint)
			assert.Equal(t, "webtunnel", bridge.Transport)
		})
	}
}

func TestParseBridgeLine_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty line",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "single token",
			input:   "192.168.1.1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge, err := ParseBridgeLine(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bridge)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bridge)
			}
		})
	}
}

func TestParseBridgeLine_MissingFields(t *testing.T) {
	bridge, err := ParseBridgeLine("192.168.1.1")
	assert.NoError(t, err)
	assert.NotNil(t, bridge)
	assert.Empty(t, bridge.Address)
}

func TestFormatTorrcLine_WithFingerprint(t *testing.T) {
	bridge := &Bridge{
		Transport:   "webtunnel",
		Address:     "192.168.1.1",
		Port:        443,
		Fingerprint: "ABCD1234EFGH",
	}

	line := FormatTorrcLine(bridge)
	assert.Contains(t, line, "Bridge webtunnel")
	assert.Contains(t, line, "192.168.1.1:443")
	assert.Contains(t, line, "fingerprint=ABCD1234EFGH")
}

func TestFormatTorrcLine_WithoutFingerprint(t *testing.T) {
	bridge := &Bridge{
		Transport: "webtunnel",
		Address:   "10.0.0.1",
		Port:      8080,
	}

	line := FormatTorrcLine(bridge)
	assert.Equal(t, "Bridge webtunnel 10.0.0.1:8080", line)
}

func TestBridge_Timestamps(t *testing.T) {
	bridge := &Bridge{
		Address:      "192.168.1.1",
		Port:         443,
		Transport:    "webtunnel",
		DiscoveredAt: time.Now(),
		IsAvailable:  -1,
	}

	assert.False(t, bridge.DiscoveredAt.IsZero())
	assert.Equal(t, -1, bridge.IsAvailable)
}

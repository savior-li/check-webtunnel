package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Bridge struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Hash        string    `gorm:"uniqueIndex;size:64" json:"hash"`
	Transport   string    `gorm:"size:32" json:"transport"`
	Address     string    `gorm:"size:256" json:"address"`
	Port        int       `gorm:"port" json:"port"`
	IPv6        string    `gorm:"size:64" json:"ipv6,omitempty"`
	Extort      string    `gorm:"size:512" json:"extort,omitempty"`
	FirstSeen   time.Time `gorm:"first_seen" json:"first_seen"`
	LastSeen    time.Time `gorm:"last_seen" json:"last_seen"`
	IsValid     bool      `gorm:"is_valid;default:false" json:"is_valid"`
	AvgLatency  float64   `gorm:"avg_latency;default:0" json:"avg_latency"`
	SuccessRate float64   `gorm:"success_rate;default:0" json:"success_rate"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ValidationHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	BridgeID    uint      `gorm:"index" json:"bridge_id"`
	Bridge      Bridge    `gorm:"foreignKey:BridgeID" json:"-"`
	TestedAt    time.Time `gorm:"tested_at" json:"tested_at"`
	Latency     float64   `gorm:"latency" json:"latency"`
	IsReachable bool      `gorm:"is_reachable" json:"is_reachable"`
	ErrorMsg    string    `gorm:"size:512" json:"error_msg,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (b *Bridge) CalculateHash() string {
	data := fmt.Sprintf("%s:%s:%d", b.Transport, b.Address, b.Port)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (b *Bridge) ToTorrcLine() string {
	return fmt.Sprintf("Bridge %s %s:%d %s", b.Transport, b.Address, b.Port, b.Hash)
}

func (b *Bridge) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"hash":         b.Hash,
		"transport":    b.Transport,
		"address":      b.Address,
		"port":         b.Port,
		"ipv6":         b.IPv6,
		"is_valid":     b.IsValid,
		"avg_latency":  b.AvgLatency,
		"success_rate": b.SuccessRate,
		"first_seen":   b.FirstSeen,
		"last_seen":    b.LastSeen,
	}
}

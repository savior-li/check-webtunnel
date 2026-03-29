package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"tor-bridge-collector/pkg/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	lock sync.Mutex
)

type Storage struct {
	db *gorm.DB
}

func New(dbPath string) (*Storage, error) {
	lock.Lock()
	defer lock.Unlock()

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	database, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := database.AutoMigrate(&models.Bridge{}, &models.ValidationHistory{}); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	db = database
	return &Storage{db: database}, nil
}

func InitDB(dbPath string) error {
	_, err := New(dbPath)
	return err
}

func (s *Storage) CreateBridge(bridge *models.Bridge) error {
	bridge.Hash = bridge.CalculateHash()
	result := s.db.FirstOrCreate(bridge, models.Bridge{Hash: bridge.Hash})
	return result.Error
}

func (s *Storage) UpdateBridge(bridge *models.Bridge) error {
	return s.db.Save(bridge).Error
}

func (s *Storage) GetBridgeByHash(hash string) (*models.Bridge, error) {
	var bridge models.Bridge
	result := s.db.Where("hash = ?", hash).First(&bridge)
	if result.Error != nil {
		return nil, result.Error
	}
	return &bridge, nil
}

func (s *Storage) GetAllBridges(filter *models.BridgeFilter) ([]models.Bridge, int64, error) {
	var bridges []models.Bridge
	var total int64

	query := s.db.Model(&models.Bridge{})

	if filter != nil {
		if filter.Transport != "" {
			query = query.Where("transport = ?", filter.Transport)
		}
		if filter.IsValid != nil {
			query = query.Where("is_valid = ?", *filter.IsValid)
		}
		if filter.MinLatency > 0 {
			query = query.Where("avg_latency >= ?", filter.MinLatency)
		}
		if filter.MaxLatency > 0 {
			query = query.Where("avg_latency <= ?", filter.MaxLatency)
		}
	}

	query.Count(&total)

	if filter != nil && filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	result := query.Order("last_seen DESC").Find(&bridges)
	return bridges, total, result.Error
}

func (s *Storage) GetBridgeByID(id uint) (*models.Bridge, error) {
	var bridge models.Bridge
	result := s.db.First(&bridge, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &bridge, nil
}

func (s *Storage) CreateValidationHistory(history *models.ValidationHistory) error {
	return s.db.Create(history).Error
}

func (s *Storage) GetValidationHistory(bridgeID uint, limit int) ([]models.ValidationHistory, error) {
	var histories []models.ValidationHistory
	result := s.db.Where("bridge_id = ?", bridgeID).
		Order("tested_at DESC").
		Limit(limit).
		Find(&histories)
	return histories, result.Error
}

func (s *Storage) GetStats() (*models.Stats, error) {
	var stats models.Stats

	var total, valid int64
	s.db.Model(&models.Bridge{}).Count(&total)
	s.db.Model(&models.Bridge{}).Where("is_valid = ?", true).Count(&valid)

	stats.TotalBridges = total
	stats.ValidBridges = valid
	stats.InvalidBridges = total - valid

	var avgLatency float64
	s.db.Model(&models.Bridge{}).Where("is_valid = ?", true).Select("COALESCE(AVG(avg_latency), 0)").Scan(&avgLatency)
	stats.AvgLatency = avgLatency

	var totalHistory, reachable int64
	s.db.Model(&models.ValidationHistory{}).Count(&totalHistory)
	s.db.Model(&models.ValidationHistory{}).Where("is_reachable = ?", true).Count(&reachable)
	stats.TotalValidations = totalHistory
	if totalHistory > 0 {
		stats.SuccessRate = float64(reachable) / float64(totalHistory) * 100
	}

	return &stats, nil
}

func (s *Storage) DeleteBridge(id uint) error {
	return s.db.Delete(&models.Bridge{}, id).Error
}

func (s *Storage) GetDB() *gorm.DB {
	return s.db
}

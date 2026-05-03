package storage

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"vpsping/internal/models"
)

type Storage struct {
	db *gorm.DB
}

func New(databasePath string) (*Storage, error) {
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err := db.AutoMigrate(&models.VPSServer{}, &models.PingResult{}, &models.Statistic{}); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *Storage) CreateVPSServer(vps *models.VPSServer) error {
	return s.db.Create(vps).Error
}

func (s *Storage) GetVPSServer(id uint) (*models.VPSServer, error) {
	var vps models.VPSServer
	err := s.db.First(&vps, id).Error
	if err != nil {
		return nil, err
	}
	return &vps, nil
}

func (s *Storage) GetVPSServerByName(name string) (*models.VPSServer, error) {
	var vps models.VPSServer
	err := s.db.Where("name = ?", name).First(&vps).Error
	if err != nil {
		return nil, err
	}
	return &vps, nil
}

func (s *Storage) ListVPSServers() ([]models.VPSServer, error) {
	var vpsList []models.VPSServer
	err := s.db.Find(&vpsList).Error
	return vpsList, err
}

func (s *Storage) ListEnabledVPSServers() ([]models.VPSServer, error) {
	var vpsList []models.VPSServer
	err := s.db.Where("enabled = ?", true).Find(&vpsList).Error
	return vpsList, err
}

func (s *Storage) UpdateVPSServer(vps *models.VPSServer) error {
	return s.db.Save(vps).Error
}

func (s *Storage) DeleteVPSServer(id uint) error {
	if err := s.db.Where("vps_id = ?", id).Delete(&models.PingResult{}).Error; err != nil {
		return fmt.Errorf("删除 Ping 结果失败: %w", err)
	}

	if err := s.db.Where("vps_id = ?", id).Delete(&models.Statistic{}).Error; err != nil {
		return fmt.Errorf("删除统计数据失败: %w", err)
	}

	if err := s.db.Delete(&models.VPSServer{}, id).Error; err != nil {
		return fmt.Errorf("删除 VPS 失败: %w", err)
	}

	return nil
}

func (s *Storage) SavePingResult(result *models.PingResult) error {
	return s.db.Create(result).Error
}

func (s *Storage) GetPingResults(vpsID uint, start, end time.Time) ([]models.PingResult, error) {
	var results []models.PingResult
	err := s.db.Where("vps_id = ? AND timestamp BETWEEN ? AND ?", vpsID, start, end).
		Order("timestamp ASC").
		Find(&results).Error
	return results, err
}

func (s *Storage) GetLatestPingResult(vpsID uint) (*models.PingResult, error) {
	var result models.PingResult
	err := s.db.Where("vps_id = ?", vpsID).
		Order("timestamp DESC").
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Storage) SaveStatistic(stat *models.Statistic) error {
	return s.db.Create(stat).Error
}

func (s *Storage) GetStatistics(vpsID uint, start, end time.Time) ([]models.Statistic, error) {
	var stats []models.Statistic
	err := s.db.Where("vps_id = ? AND period_start BETWEEN ? AND ?", vpsID, start, end).
		Order("period_start ASC").
		Find(&stats).Error
	return stats, err
}

func (s *Storage) SyncVPSFromConfig(vpsConfigs []struct {
	Name    string
	Host    string
	Enabled bool
}) error {
	for _, cfg := range vpsConfigs {
		var vps models.VPSServer
		result := s.db.Where("name = ?", cfg.Name).First(&vps)

		if result.Error == gorm.ErrRecordNotFound {
			vps = models.VPSServer{
				Name:    cfg.Name,
				Host:    cfg.Host,
				Enabled: cfg.Enabled,
			}
			if err := s.CreateVPSServer(&vps); err != nil {
				return fmt.Errorf("创建 VPS %s 失败: %w", cfg.Name, err)
			}
		} else if result.Error == nil {
			vps.Host = cfg.Host
			vps.Enabled = cfg.Enabled
			if err := s.UpdateVPSServer(&vps); err != nil {
				return fmt.Errorf("更新 VPS %s 失败: %w", cfg.Name, err)
			}
		} else {
			return result.Error
		}
	}

	return nil
}

func (s *Storage) CleanupOldData(before time.Time) error {
	if err := s.db.Where("timestamp < ?", before).Delete(&models.PingResult{}).Error; err != nil {
		return err
	}
	return s.db.Where("period_end < ?", before).Delete(&models.Statistic{}).Error
}

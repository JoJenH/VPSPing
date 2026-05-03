package models

import "time"

type VPSServer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	Host      string    `gorm:"not null" json:"host"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PingResult struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	VPSSID      uint      `gorm:"column:vps_id;index;not null" json:"vps_id"`
	Timestamp   time.Time `gorm:"index;not null" json:"timestamp"`
	LatencyAvg  float64   `json:"latency_avg"`
	LatencyMin  float64   `json:"latency_min"`
	LatencyMax  float64   `json:"latency_max"`
	TTL         int       `json:"ttl"`
	PacketsSent int       `json:"packets_sent"`
	PacketsRecv int       `json:"packets_recv"`
	PacketLoss  float64   `json:"packet_loss"`
	VPSServer   VPSServer `gorm:"foreignKey:VPSSID" json:"vps_server"`
}

type Statistic struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	VPSSID         uint      `gorm:"column:vps_id;index;not null" json:"vps_id"`
	PeriodStart    time.Time `gorm:"not null" json:"period_start"`
	PeriodEnd      time.Time `gorm:"not null" json:"period_end"`
	AvgLatency     float64   `json:"avg_latency"`
	MinLatency     float64   `json:"min_latency"`
	MaxLatency     float64   `json:"max_latency"`
	AvgTTL         int       `json:"avg_ttl"`
	TotalPackets   int       `json:"total_packets"`
	LostPackets    int       `json:"lost_packets"`
	PacketLossRate float64   `json:"packet_loss_rate"`
	VPSServer      VPSServer `gorm:"foreignKey:VPSSID" json:"vps_server"`
}

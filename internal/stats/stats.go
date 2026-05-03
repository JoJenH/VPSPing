package stats

import (
	"math"
	"time"

	"vpsping/internal/models"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

type StatsResult struct {
	AvgLatency float64
	MinLatency float64
	MaxLatency float64
	AvgTTL     int
	PacketLoss float64
	TotalTests int
}

func (c *Calculator) CalculateFromResults(results []models.PingResult) *StatsResult {
	if len(results) == 0 {
		return &StatsResult{}
	}

	var sumLatency, minLatency, maxLatency float64
	var sumTTL, validTTLCount int
	var totalSent, totalRecv int

	minLatency = math.MaxFloat64
	maxLatency = 0

	for _, r := range results {
		if r.LatencyAvg > 0 {
			sumLatency += r.LatencyAvg
			if r.LatencyMin < minLatency && r.LatencyMin > 0 {
				minLatency = r.LatencyMin
			}
			if r.LatencyMax > maxLatency {
				maxLatency = r.LatencyMax
			}
		}

		if r.TTL > 0 {
			sumTTL += r.TTL
			validTTLCount++
		}

		totalSent += r.PacketsSent
		totalRecv += r.PacketsRecv
	}

	result := &StatsResult{
		TotalTests: len(results),
	}

	if sumLatency > 0 {
		result.AvgLatency = sumLatency / float64(len(results))
		result.MinLatency = minLatency
		result.MaxLatency = maxLatency
	}

	if validTTLCount > 0 {
		result.AvgTTL = sumTTL / validTTLCount
	}

	if totalSent > 0 {
		result.PacketLoss = float64(totalSent-totalRecv) / float64(totalSent) * 100
	}

	return result
}

func (c *Calculator) CalculateStatistics(vpsID uint, results []models.PingResult, periodStart, periodEnd time.Time) *models.Statistic {
	stats := c.CalculateFromResults(results)

	return &models.Statistic{
		VPSSID:         vpsID,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		AvgLatency:     stats.AvgLatency,
		MinLatency:     stats.MinLatency,
		MaxLatency:     stats.MaxLatency,
		AvgTTL:         stats.AvgTTL,
		TotalPackets:   stats.TotalTests,
		PacketLossRate: stats.PacketLoss,
	}
}

func (c *Calculator) FilterByTimeRange(results []models.PingResult, duration time.Duration) []models.PingResult {
	if duration <= 0 {
		return results
	}

	cutoff := time.Now().Add(-duration)
	filtered := make([]models.PingResult, 0)

	for _, r := range results {
		if r.Timestamp.After(cutoff) {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

func (c *Calculator) GetLatencyDataPoints(results []models.PingResult) []LatencyPoint {
	points := make([]LatencyPoint, 0, len(results))

	for _, r := range results {
		if r.LatencyAvg > 0 {
			points = append(points, LatencyPoint{
				Timestamp: r.Timestamp,
				Value:     r.LatencyAvg,
			})
		}
	}

	return points
}

type LatencyPoint struct {
	Timestamp time.Time
	Value     float64
}

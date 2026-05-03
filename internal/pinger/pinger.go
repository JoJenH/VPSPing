package pinger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"vpsping/internal/models"

	probing "github.com/prometheus-community/pro-bing"
)

type Pinger struct {
	count      int
	timeout    time.Duration
	privileged bool
}

type PingResult struct {
	VPSName     string
	Host        string
	LatencyAvg  float64
	LatencyMin  float64
	LatencyMax  float64
	TTL         int
	PacketsSent int
	PacketsRecv int
	PacketLoss  float64
	Error       error
}

func NewPinger(count int, timeout time.Duration, privileged bool) *Pinger {
	return &Pinger{
		count:      count,
		timeout:    timeout,
		privileged: privileged,
	}
}

func (p *Pinger) Ping(host string) (*PingResult, error) {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return nil, fmt.Errorf("创建 pinger 失败: %w", err)
	}

	pinger.Count = p.count
	pinger.Timeout = p.timeout
	pinger.SetPrivileged(p.privileged)

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout+time.Second*2)
	defer cancel()

	done := make(chan struct{})
	var stats *probing.Statistics

	go func() {
		if err := pinger.Run(); err != nil {
			stats = &probing.Statistics{
				PacketsSent: p.count,
				PacketsRecv: 0,
			}
		} else {
			stats = pinger.Statistics()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return nil, fmt.Errorf("ping 超时")
	}

	result := &PingResult{
		PacketsSent: stats.PacketsSent,
		PacketsRecv: stats.PacketsRecv,
		PacketLoss:  float64(stats.PacketsSent-stats.PacketsRecv) / float64(stats.PacketsSent) * 100,
	}

	if stats.PacketsRecv > 0 {
		result.LatencyAvg = float64(stats.AvgRtt.Milliseconds())
		result.LatencyMin = float64(stats.MinRtt.Milliseconds())
		result.LatencyMax = float64(stats.MaxRtt.Milliseconds())
	}

	return result, nil
}

func (p *Pinger) PingMultiple(vpsList []struct {
	Name string
	Host string
}) []*PingResult {
	var wg sync.WaitGroup
	results := make([]*PingResult, len(vpsList))

	for i, vps := range vpsList {
		wg.Add(1)
		go func(idx int, name, host string) {
			defer wg.Done()

			result := &PingResult{
				VPSName: name,
				Host:    host,
			}

			pingResult, err := p.Ping(host)
			if err != nil {
				result.Error = err
				results[idx] = result
				return
			}

			result.LatencyAvg = pingResult.LatencyAvg
			result.LatencyMin = pingResult.LatencyMin
			result.LatencyMax = pingResult.LatencyMax
			result.TTL = pingResult.TTL
			result.PacketsSent = pingResult.PacketsSent
			result.PacketsRecv = pingResult.PacketsRecv
			result.PacketLoss = pingResult.PacketLoss

			results[idx] = result
		}(i, vps.Name, vps.Host)
	}

	wg.Wait()
	return results
}

func (r *PingResult) ToModel(vpsID uint) *models.PingResult {
	return &models.PingResult{
		VPSSID:      vpsID,
		Timestamp:   time.Now(),
		LatencyAvg:  r.LatencyAvg,
		LatencyMin:  r.LatencyMin,
		LatencyMax:  r.LatencyMax,
		TTL:         r.TTL,
		PacketsSent: r.PacketsSent,
		PacketsRecv: r.PacketsRecv,
		PacketLoss:  r.PacketLoss,
	}
}

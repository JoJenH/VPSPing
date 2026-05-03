package scheduler

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"vpsping/internal/config"
	"vpsping/internal/models"
	"vpsping/internal/output"
	"vpsping/internal/pinger"
	"vpsping/internal/stats"
	"vpsping/internal/storage"
)

type Scheduler struct {
	config  *config.Config
	storage *storage.Storage
	pinger  *pinger.Pinger
	output  *output.Output
	calc    *stats.Calculator
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func New(cfg *config.Config, store *storage.Storage, out *output.Output) *Scheduler {
	p := pinger.NewPinger(
		cfg.Ping.Count,
		cfg.GetTimeout(),
		cfg.Ping.Privileged,
	)

	return &Scheduler{
		config:  cfg,
		storage: store,
		pinger:  p,
		output:  out,
		calc:    stats.NewCalculator(),
	}
}

func (s *Scheduler) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.setupSignalHandler()

	s.output.WriteLog("启动 VPS 延迟监控")

	s.runTest()

	ticker := time.NewTicker(s.config.GetInterval())
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.output.WriteLog("停止 VPS 延迟监控")
			return nil
		case <-ticker.C:
			s.runTest()
		}
	}
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) runTest() {
	s.wg.Add(1)
	defer s.wg.Done()

	vpsList, err := s.storage.ListEnabledVPSServers()
	if err != nil {
		s.output.WriteLog("获取 VPS 列表失败: %v", err)
		return
	}

	if len(vpsList) == 0 {
		s.output.WriteLog("没有启用的 VPS")
		return
	}

	testList := make([]struct {
		Name string
		Host string
	}, len(vpsList))

	for i, vps := range vpsList {
		testList[i].Name = vps.Name
		testList[i].Host = vps.Host
	}

	results := s.pinger.PingMultiple(testList)

	s.output.PrintTable(results)

	for i, result := range results {
		if result.Error != nil {
			s.output.WriteLog("VPS %s (%s) 测试失败: %v", result.VPSName, result.Host, result.Error)
			continue
		}

		pingResult := result.ToModel(vpsList[i].ID)
		if err := s.storage.SavePingResult(pingResult); err != nil {
			s.output.WriteLog("保存 VPS %s 测试结果失败: %v", result.VPSName, err)
		}

		s.output.WriteLog("VPS %s: 平均延迟 %.2fms, TTL %d, 丢包率 %.1f%%",
			result.VPSName, result.LatencyAvg, result.TTL, result.PacketLoss)
	}

	if err := s.output.WriteJSON(results); err != nil {
		s.output.WriteLog("写入 JSON 文件失败: %v", err)
	}
}

func (s *Scheduler) RunOnce() error {
	s.runTest()
	return nil
}

func (s *Scheduler) setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n接收到中断信号，正在关闭...")
		s.Stop()
		os.Exit(0)
	}()
}

func (s *Scheduler) GetStats(vpsName string, duration time.Duration) (*stats.StatsResult, error) {
	vps, err := s.storage.GetVPSServerByName(vpsName)
	if err != nil {
		return nil, fmt.Errorf("VPS %s 不存在", vpsName)
	}

	results, err := s.storage.GetPingResults(vps.ID, time.Now().Add(-duration), time.Now())
	if err != nil {
		return nil, err
	}

	return s.calc.CalculateFromResults(results), nil
}

func (s *Scheduler) GetChartData(vpsName string, duration time.Duration) ([]stats.LatencyPoint, error) {
	vps, err := s.storage.GetVPSServerByName(vpsName)
	if err != nil {
		return nil, fmt.Errorf("VPS %s 不存在", vpsName)
	}

	results, err := s.storage.GetPingResults(vps.ID, time.Now().Add(-duration), time.Now())
	if err != nil {
		return nil, err
	}

	return s.calc.GetLatencyDataPoints(results), nil
}

func (s *Scheduler) ListVPS() ([]models.VPSServer, error) {
	return s.storage.ListVPSServers()
}

func (s *Scheduler) AddVPS(name, host string, enabled bool) error {
	vps := &models.VPSServer{
		Name:    name,
		Host:    host,
		Enabled: enabled,
	}
	return s.storage.CreateVPSServer(vps)
}

func (s *Scheduler) RemoveVPS(name string) error {
	vps, err := s.storage.GetVPSServerByName(name)
	if err != nil {
		return err
	}
	return s.storage.DeleteVPSServer(vps.ID)
}

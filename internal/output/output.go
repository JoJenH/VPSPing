package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vpsping/internal/pinger"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type Output struct {
	logFile    string
	jsonFile   string
	logHandle  *os.File
	jsonHandle *os.File
}

type JSONOutput struct {
	Timestamp  string               `json:"timestamp"`
	Results    []JSONResult         `json:"results"`
	Statistics JSONOutputStatistics `json:"statistics"`
}

type JSONResult struct {
	VPSName     string  `json:"vps_name"`
	Host        string  `json:"host"`
	LatencyAvg  float64 `json:"latency_avg"`
	LatencyMin  float64 `json:"latency_min"`
	LatencyMax  float64 `json:"latency_max"`
	TTL         int     `json:"ttl"`
	PacketsSent int     `json:"packets_sent"`
	PacketsRecv int     `json:"packets_recv"`
	PacketLoss  float64 `json:"packet_loss"`
	Error       string  `json:"error,omitempty"`
}

type JSONOutputStatistics struct {
	TotalVPS     int `json:"total_vps"`
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
}

func New(logFile, jsonFile string) (*Output, error) {
	o := &Output{
		logFile:  logFile,
		jsonFile: jsonFile,
	}

	if err := o.initLogFile(); err != nil {
		return nil, err
	}

	if err := o.initJSONFile(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Output) initLogFile() error {
	dir := filepath.Dir(o.logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	file, err := os.OpenFile(o.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	o.logHandle = file
	return nil
}

func (o *Output) initJSONFile() error {
	dir := filepath.Dir(o.jsonFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建 JSON 输出目录失败: %w", err)
	}

	file, err := os.OpenFile(o.jsonFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开 JSON 文件失败: %w", err)
	}

	o.jsonHandle = file
	return nil
}

func (o *Output) Close() error {
	if o.logHandle != nil {
		o.logHandle.Close()
	}
	if o.jsonHandle != nil {
		o.jsonHandle.Close()
	}
	return nil
}

func (o *Output) WriteLog(format string, args ...interface{}) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
	_, err := o.logHandle.WriteString(message)
	return err
}

func (o *Output) WriteJSON(results []*pinger.PingResult) error {
	jsonOutput := JSONOutput{
		Timestamp: time.Now().Format(time.RFC3339),
		Results:   make([]JSONResult, 0, len(results)),
		Statistics: JSONOutputStatistics{
			TotalVPS: len(results),
		},
	}

	for _, r := range results {
		result := JSONResult{
			VPSName:     r.VPSName,
			Host:        r.Host,
			LatencyAvg:  r.LatencyAvg,
			LatencyMin:  r.LatencyMin,
			LatencyMax:  r.LatencyMax,
			TTL:         r.TTL,
			PacketsSent: r.PacketsSent,
			PacketsRecv: r.PacketsRecv,
			PacketLoss:  r.PacketLoss,
		}

		if r.Error != nil {
			result.Error = r.Error.Error()
			jsonOutput.Statistics.FailedCount++
		} else {
			jsonOutput.Statistics.SuccessCount++
		}

		jsonOutput.Results = append(jsonOutput.Results, result)
	}

	data, err := json.MarshalIndent(jsonOutput, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %w", err)
	}

	if err := o.jsonHandle.Truncate(0); err != nil {
		return err
	}
	if _, err := o.jsonHandle.Seek(0, 0); err != nil {
		return err
	}
	_, err = o.jsonHandle.Write(data)
	return err
}

func (o *Output) PrintTable(results []*pinger.PingResult) {
	fmt.Printf("\n[%s] Testing VPS servers...\n\n", time.Now().Format("2006-01-02 15:04:05"))

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"VPS Name", "Avg (ms)", "Min (ms)", "Max (ms)", "TTL", "Loss (%)", "Status"})

	for _, r := range results {
		row := []string{
			r.VPSName,
			formatLatency(r.LatencyAvg),
			formatLatency(r.LatencyMin),
			formatLatency(r.LatencyMax),
			fmt.Sprintf("%d", r.TTL),
			fmt.Sprintf("%.1f", r.PacketLoss),
		}

		if r.Error != nil {
			row = append(row, color.RedString("FAILED"))
		} else if r.PacketLoss > 0 {
			row = append(row, color.YellowString("LOSS"))
		} else {
			row = append(row, color.GreenString("OK"))
		}

		table.Append(row)
	}

	table.Render()
	fmt.Println()
}

func formatLatency(latency float64) string {
	if latency == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f", latency)
}

func (o *Output) PrintStats(vpsName string, avgLatency, minLatency, maxLatency float64, avgTTL int, packetLoss float64, totalTests int) {
	fmt.Printf("\nStatistics for %s (Last %d tests):\n", color.CyanString(vpsName), totalTests)
	fmt.Printf("  Average Latency: %.2f ms\n", avgLatency)
	fmt.Printf("  Minimum Latency: %.2f ms\n", minLatency)
	fmt.Printf("  Maximum Latency: %.2f ms\n", maxLatency)
	fmt.Printf("  Average TTL: %d\n", avgTTL)
	fmt.Printf("  Packet Loss: %.2f%%\n", packetLoss)
	fmt.Println()
}

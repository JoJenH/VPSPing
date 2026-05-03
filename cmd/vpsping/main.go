package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"vpsping/internal/config"
	"vpsping/internal/output"
	"vpsping/internal/scheduler"
	"vpsping/internal/stats"
	"vpsping/internal/storage"
)

var (
	configFile string
	version    = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "vpsping",
	Short: "VPS 延迟监控工具",
	Long:  "一个轻量级的命令行工具，用于周期性测试多个 VPS 服务器的网络延迟",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "启动持续监控",
	Long:  "启动持续监控，按照配置的间隔时间定期测试所有 VPS 的延迟",
	RunE:  runMonitor,
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "执行一次测试",
	Long:  "对所有启用的 VPS 执行一次延迟测试",
	RunE:  runTest,
}

var statsCmd = &cobra.Command{
	Use:   "stats [vps-name]",
	Short: "显示统计信息",
	Long:  "显示指定 VPS 或所有 VPS 的统计信息",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStats,
}

var chartCmd = &cobra.Command{
	Use:   "chart [vps-name]",
	Short: "显示延迟趋势图",
	Long:  "显示指定 VPS 或所有 VPS 的延迟趋势图",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runChart,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有 VPS",
	Long:  "列出配置中的所有 VPS 及其状态",
	RunE:  runList,
}

var addCmd = &cobra.Command{
	Use:   "add <name> <host>",
	Short: "添加新 VPS",
	Long:  "添加一个新的 VPS 到监控列表",
	Args:  cobra.ExactArgs(2),
	RunE:  runAdd,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "删除 VPS",
	Long:  "从监控列表中删除指定的 VPS",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Long:  "创建一个示例配置文件",
	RunE:  runInit,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "配置文件路径")
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(chartCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(initCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadConfig() (*config.Config, error) {
	if configFile != "" {
		return config.Load(configFile)
	}
	return config.Load("")
}

func initStorage(cfg *config.Config) (*storage.Storage, error) {
	return storage.New(cfg.Storage.Database)
}

func initOutput(cfg *config.Config) (*output.Output, error) {
	return output.New(cfg.Storage.LogFile, cfg.Storage.JsonOutput)
}

func runMonitor(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer store.Close()

	if err := syncVPSFromConfig(store, cfg); err != nil {
		return err
	}

	out, err := initOutput(cfg)
	if err != nil {
		return fmt.Errorf("初始化输出失败: %w", err)
	}
	defer out.Close()

	sched := scheduler.New(cfg, store, out)
	return sched.Start()
}

func runTest(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer store.Close()

	if err := syncVPSFromConfig(store, cfg); err != nil {
		return err
	}

	out, err := initOutput(cfg)
	if err != nil {
		return fmt.Errorf("初始化输出失败: %w", err)
	}
	defer out.Close()

	sched := scheduler.New(cfg, store, out)
	return sched.RunOnce()
}

func runStats(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer store.Close()

	out, err := initOutput(cfg)
	if err != nil {
		return fmt.Errorf("初始化输出失败: %w", err)
	}
	defer out.Close()

	sched := scheduler.New(cfg, store, out)

	duration := cfg.GetTimeRange()

	if len(args) > 0 {
		vpsName := args[0]
		result, err := sched.GetStats(vpsName, duration)
		if err != nil {
			return err
		}
		out.PrintStats(vpsName, result.AvgLatency, result.MinLatency, result.MaxLatency,
			result.AvgTTL, result.PacketLoss, result.TotalTests)
	} else {
		vpsList, err := sched.ListVPS()
		if err != nil {
			return err
		}

		for _, vps := range vpsList {
			result, err := sched.GetStats(vps.Name, duration)
			if err != nil {
				fmt.Printf("获取 %s 统计信息失败: %v\n", vps.Name, err)
				continue
			}
			out.PrintStats(vps.Name, result.AvgLatency, result.MinLatency, result.MaxLatency,
				result.AvgTTL, result.PacketLoss, result.TotalTests)
		}
	}

	return nil
}

func runChart(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer store.Close()

	sched := scheduler.New(cfg, store, nil)

	chart := stats.NewChart(cfg.Display.ChartWidth, cfg.Display.ChartHeight)
	duration := cfg.GetTimeRange()

	if len(args) > 0 {
		vpsName := args[0]
		points, err := sched.GetChartData(vpsName, duration)
		if err != nil {
			return err
		}
		title := fmt.Sprintf("Latency Trend for %s (Last %s)", vpsName, stats.FormatDuration(duration))
		fmt.Println(chart.DrawLineChart(points, title))
	} else {
		vpsList, err := sched.ListVPS()
		if err != nil {
			return err
		}

		dataSets := make(map[string][]stats.LatencyPoint)
		for _, vps := range vpsList {
			points, err := sched.GetChartData(vps.Name, duration)
			if err != nil {
				continue
			}
			if len(points) > 0 {
				dataSets[vps.Name] = points
			}
		}

		title := fmt.Sprintf("Latency Trend (Last %s)", stats.FormatDuration(duration))
		fmt.Println(chart.DrawMultiLineChart(dataSets, title))
	}

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer store.Close()

	sched := scheduler.New(cfg, store, nil)

	vpsList, err := sched.ListVPS()
	if err != nil {
		return err
	}

	fmt.Println("\nConfigured VPS Servers:")
	fmt.Println("┌─────────────────────┬──────────────────────┬─────────┐")
	fmt.Println("│ Name                │ Host                 │ Enabled │")
	fmt.Println("├─────────────────────┼──────────────────────┼─────────┤")

	for _, vps := range vpsList {
		enabled := "No"
		if vps.Enabled {
			enabled = "Yes"
		}
		fmt.Printf("│ %-19s │ %-20s │ %-7s │\n", vps.Name, vps.Host, enabled)
	}

	fmt.Println("└─────────────────────┴──────────────────────┴─────────┘")
	fmt.Printf("\nTotal: %d VPS servers\n", len(vpsList))

	return nil
}

func runAdd(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer store.Close()

	sched := scheduler.New(cfg, store, nil)

	name := args[0]
	host := args[1]

	if err := sched.AddVPS(name, host, true); err != nil {
		return fmt.Errorf("添加 VPS 失败: %w", err)
	}

	fmt.Printf("成功添加 VPS: %s (%s)\n", name, host)
	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer store.Close()

	sched := scheduler.New(cfg, store, nil)

	name := args[0]

	if err := sched.RemoveVPS(name); err != nil {
		return fmt.Errorf("删除 VPS 失败: %w", err)
	}

	fmt.Printf("成功删除 VPS: %s\n", name)
	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := configFile
	if configPath == "" {
		configPath = "config.yaml"
	}

	if err := config.CreateDefaultConfigFile(configPath); err != nil {
		return err
	}

	fmt.Printf("已创建示例配置文件: %s\n", configPath)
	fmt.Println("请编辑配置文件，添加你的 VPS 信息")
	return nil
}

func syncVPSFromConfig(store *storage.Storage, cfg *config.Config) error {
	vpsConfigs := make([]struct {
		Name    string
		Host    string
		Enabled bool
	}, len(cfg.VPS))

	for i, vps := range cfg.VPS {
		vpsConfigs[i].Name = vps.Name
		vpsConfigs[i].Host = vps.Host
		vpsConfigs[i].Enabled = vps.Enabled
	}

	return store.SyncVPSFromConfig(vpsConfigs)
}

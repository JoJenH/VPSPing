package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type VPSConfig struct {
	Name    string `mapstructure:"name"`
	Host    string `mapstructure:"host"`
	Enabled bool   `mapstructure:"enabled"`
}

type PingConfig struct {
	Interval   string `mapstructure:"interval"`
	Count      int    `mapstructure:"count"`
	Timeout    string `mapstructure:"timeout"`
	Privileged bool   `mapstructure:"privileged"`
}

type StorageConfig struct {
	Database   string `mapstructure:"database"`
	LogFile    string `mapstructure:"log_file"`
	JsonOutput string `mapstructure:"json_output"`
}

type DisplayConfig struct {
	ChartWidth  int    `mapstructure:"chart_width"`
	ChartHeight int    `mapstructure:"chart_height"`
	TimeRange   string `mapstructure:"time_range"`
}

type Config struct {
	VPS     []VPSConfig   `mapstructure:"vps"`
	Ping    PingConfig    `mapstructure:"ping"`
	Storage StorageConfig `mapstructure:"storage"`
	Display DisplayConfig `mapstructure:"display"`
}

func DefaultConfig() *Config {
	return &Config{
		VPS: []VPSConfig{},
		Ping: PingConfig{
			Interval:   "15m",
			Count:      4,
			Timeout:    "5s",
			Privileged: true,
		},
		Storage: StorageConfig{
			Database:   "./data/vpsping.db",
			LogFile:    "./logs/vpsping.log",
			JsonOutput: "./output/results.json",
		},
		Display: DisplayConfig{
			ChartWidth:  80,
			ChartHeight: 20,
			TimeRange:   "24h",
		},
	}
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("$HOME/.vpsping")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("配置文件未找到: %s", configPath)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if len(config.VPS) == 0 {
		return fmt.Errorf("配置文件中至少需要一个 VPS")
	}

	for i, vps := range config.VPS {
		if vps.Name == "" {
			return fmt.Errorf("VPS #%d 缺少名称", i+1)
		}
		if vps.Host == "" {
			return fmt.Errorf("VPS #%d (%s) 缺少主机地址", i+1, vps.Name)
		}
	}

	if config.Ping.Count <= 0 {
		return fmt.Errorf("ping count 必须大于 0")
	}

	if _, err := time.ParseDuration(config.Ping.Interval); err != nil {
		return fmt.Errorf("无效的 interval 格式: %w", err)
	}

	if _, err := time.ParseDuration(config.Ping.Timeout); err != nil {
		return fmt.Errorf("无效的 timeout 格式: %w", err)
	}

	return nil
}

func Save(configPath string, config *Config) error {
	v := viper.New()

	v.Set("vps", config.VPS)
	v.Set("ping", config.Ping)
	v.Set("storage", config.Storage)
	v.Set("display", config.Display)

	if configPath == "" {
		configPath = "config.yaml"
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

func CreateDefaultConfigFile(configPath string) error {
	config := DefaultConfig()

	config.VPS = []VPSConfig{
		{
			Name:    "Example-VPS-1",
			Host:    "example.com",
			Enabled: true,
		},
		{
			Name:    "Example-VPS-2",
			Host:    "192.168.1.1",
			Enabled: false,
		},
	}

	if configPath == "" {
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("配置文件已存在: %s", configPath)
	}

	return Save(configPath, config)
}

func (c *Config) GetInterval() time.Duration {
	d, _ := time.ParseDuration(c.Ping.Interval)
	return d
}

func (c *Config) GetTimeout() time.Duration {
	d, _ := time.ParseDuration(c.Ping.Timeout)
	return d
}

func (c *Config) GetTimeRange() time.Duration {
	d, _ := time.ParseDuration(c.Display.TimeRange)
	return d
}

package pkg

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ServerCfg 单个协议网关服务配置；注释原因：与 config.yaml servers.* 节点对应；注释人：Cursor
type ServerCfg struct {
	Enabled    bool   `mapstructure:"enabled"`
	Addr       string `mapstructure:"addr"`
	ReadBuffer int    `mapstructure:"read_buffer"` // TCP 读缓冲大小（字节）
}

// MQTTCfg MQTT Broker 配置；注释原因：与 config.yaml mqtt 节点对应；注释人：Cursor
type MQTTCfg struct {
	Enabled bool   `mapstructure:"enabled"`
	Addr    string `mapstructure:"addr"`
}

// LogCfg 日志配置；注释原因：与 config.yaml log 节点对应；注释人：Cursor
type LogCfg struct {
	Level string `mapstructure:"level"` // debug | info | warn | error
}

// AppConfig 全局应用配置；注释原因：统一管理所有模块配置，viper.Unmarshal 填充；注释人：Cursor
type AppConfig struct {
	Log     LogCfg                `mapstructure:"log"`
	Servers map[string]ServerCfg  `mapstructure:"servers"`
	MQTT    MQTTCfg               `mapstructure:"mqtt"`
}

// LoadConfig 使用 viper 加载配置；注释原因：支持文件 + 环境变量两种配置源，文件不存在时回退默认值；注释人：Cursor
// 配置文件搜索路径：./config.yaml、/etc/jt808/config.yaml
// 环境变量规则：JT808_<KEY>（点替换为下划线），如 JT808_LOG_LEVEL=debug
func LoadConfig() (*AppConfig, error) {
	v := viper.New()

	// 默认值（文件缺失时兜底）
	v.SetDefault("log.level", "info")
	v.SetDefault("servers.jt808.enabled", true)
	v.SetDefault("servers.jt808.addr", ":1808")
	v.SetDefault("servers.jt808.read_buffer", 4096)
	v.SetDefault("mqtt.enabled", true)
	v.SetDefault("mqtt.addr", ":1883")

	// 配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/jt808/")

	// 环境变量覆盖（优先级高于文件）；注释原因：容器/K8s 部署时通过 ENV 注入；注释人：Cursor
	v.SetEnvPrefix("JT808")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在，使用默认值，不报错
		slog.Warn("未找到 config.yaml，使用内置默认配置")
	} else {
		slog.Info("加载配置文件", slog.String("file", v.ConfigFileUsed()))
	}

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return &cfg, nil
}

// InitLogger 根据配置初始化全局 slog；注释原因：统一日志级别控制，需在所有服务启动前调用；注释人：Cursor
func InitLogger(cfg LogCfg) {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
	slog.Info("日志初始化完成", slog.String("level", level.String()))
}

// BuildServers 根据配置构建启用的服务列表；注释原因：新增协议只需在此处注册，main.go 无需感知；注释人：Cursor
func BuildServers(cfg *AppConfig) []GatewayServer {
	var servers []GatewayServer
	if c, ok := cfg.Servers["jt808"]; ok && c.Enabled {
		servers = append(servers, NewJT808Server(c))
	}
	// TODO: 在此添加 JT1078、GB32960 等协议服务
	// if c, ok := cfg.Servers["jt1078"]; ok && c.Enabled {
	//     servers = append(servers, NewJT1078Server(c))
	// }
	return servers
}

package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 包含应用程序的所有配置
type Config struct {
	// Server 包含HTTP服务器配置
	Server struct {
		Address string `mapstructure:"address"` // 服务器监听地址，如 :8080
		Mode    string `mapstructure:"mode"`    // Gin模式: debug, release, test
	} `mapstructure:"server"`

	// MongoDB 包含MongoDB连接配置
	MongoDB struct {
		URI         string        `mapstructure:"uri"`           // MongoDB连接URI
		Database    string        `mapstructure:"database"`      // 数据库名称
		Timeout     time.Duration `mapstructure:"timeout"`       // 连接超时(秒)
		MaxPoolSize uint64        `mapstructure:"max_pool_size"` // 最大连接池大小
	} `mapstructure:"mongodb"`

	// JWT 包含JWT认证配置
	JWT struct {
		Secret     string        `mapstructure:"secret"`     // JWT签名密钥
		Expiration time.Duration `mapstructure:"expiration"` // 令牌过期时间(小时)
	} `mapstructure:"jwt"`
}

// Load 从配置文件加载配置
func Load() (*Config, error) {
	// 设置默认值
	setDefaults()

	// 设置配置文件名
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 添加配置文件路径
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}

// setDefaults 设置配置默认值
func setDefaults() {
	// 服务器默认设置
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.mode", "release")

	// MongoDB默认设置 - 使用确认可工作的参数
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017/?directConnection=true")
	viper.SetDefault("mongodb.database", "minds_iolite")
	viper.SetDefault("mongodb.timeout", 20) // 20秒
	viper.SetDefault("mongodb.max_pool_size", 100)

	// JWT默认设置
	viper.SetDefault("jwt.expiration", 24) // 24小时
}

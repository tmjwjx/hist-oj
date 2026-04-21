package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	HojAPI   HojAPIConfig   `mapstructure:"hoj_api"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Rating   RatingConfig   `mapstructure:"rating"`
	COS      COSConfig      `mapstructure:"cos"`
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type HojAPIConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type JWTConfig struct {
	Secret string
}

type RatingConfig struct {
	InitialRating int `mapstructure:"initial_rating"`
	KFactor       int `mapstructure:"k_factor"`
	CheckInterval int `mapstructure:"check_interval"`
}

type COSConfig struct {
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	CdnDomain string `mapstructure:"cdn_domain"` // CDN加速域名（可选，开启后可节省约70%流量费用）
}

// LogConfig 日志配置（简化，使用默认值）
type LogConfig struct {
}

var GlobalConfig *Config

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("../configs")
		viper.AddConfigPath("../../configs")
	}

	// 支持环境变量
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

func setDefaults() {
	viper.SetDefault("server.port", 8888)
	viper.SetDefault("server.mode", "release")

	viper.SetDefault("database.port", 3306)
	viper.SetDefault("jwt.secret", "default")
	viper.SetDefault("rating.initial_rating", 0)
	viper.SetDefault("rating.k_factor", 32)
	viper.SetDefault("rating.check_interval", 5)
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s&interpolateParams=true",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

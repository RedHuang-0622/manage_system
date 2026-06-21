package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Casbin CasbinConfig `mapstructure:"casbin"`
	Log    LogConfig    `mapstructure:"log"`
	CORS   CORSConfig   `mapstructure:"cors"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type ServerConfig struct {
	Port              int    `mapstructure:"port"`
	Mode              string `mapstructure:"mode"`
	ReadTimeout       int    `mapstructure:"read_timeout"`        // 读取完整请求超时(秒)
	ReadHeaderTimeout int    `mapstructure:"read_header_timeout"` // 仅读请求头超时(秒)，防 Slowloris
	WriteTimeout      int    `mapstructure:"write_timeout"`       // 写入响应超时(秒)
	IdleTimeout       int    `mapstructure:"idle_timeout"`        // Keep-Alive 空闲超时(秒)
	MaxHeaderBytes    int    `mapstructure:"max_header_bytes"`    // 请求头最大字节数
}

type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
	Issuer string `mapstructure:"issuer"`
}

type CasbinConfig struct {
	ModelPath string `mapstructure:"model_path"`
}

type LogConfig struct {
	Path       string `mapstructure:"path"`
	Level      string `mapstructure:"level"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// DSN 构建 MySQL 连接字符串
func (m MySQLConfig) DSN() string {
	collation := "utf8mb4_unicode_ci"
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database, m.Charset, collation)
}

func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("LAB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	setDefaults(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 10 // 10s 覆盖所有 API
	}
	if cfg.Server.ReadHeaderTimeout == 0 {
		cfg.Server.ReadHeaderTimeout = 5 // 5s 切断慢速 Header 攻击
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 10
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 30 // 30s 回收僵死连接
	}
	if cfg.Server.MaxHeaderBytes == 0 {
		cfg.Server.MaxHeaderBytes = 1 << 20 // 1MB
	}
	if cfg.MySQL.Charset == "" {
		cfg.MySQL.Charset = "utf8mb4"
	}
	if cfg.MySQL.MaxIdleConns == 0 {
		cfg.MySQL.MaxIdleConns = 10
	}
	if cfg.MySQL.MaxOpenConns == 0 {
		cfg.MySQL.MaxOpenConns = 100
	}
	if cfg.MySQL.ConnMaxLifetime == 0 {
		cfg.MySQL.ConnMaxLifetime = 3600
	}
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 100
	}
	if cfg.Redis.MinIdleConns == 0 {
		cfg.Redis.MinIdleConns = 10
	}
	if cfg.JWT.Expire == 0 {
		cfg.JWT.Expire = 7200
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "lab-system"
	}
}

func validate(cfg *Config) error {
	if cfg.MySQL.Host == "" {
		return fmt.Errorf("mysql.host 不能为空")
	}
	if cfg.MySQL.Database == "" {
		return fmt.Errorf("mysql.database 不能为空")
	}
	if cfg.MySQL.User == "" {
		return fmt.Errorf("mysql.user 不能为空")
	}
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key-at-least-32-chars" || len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("jwt.secret 必须设置且长度 ≥ 32 字符，请通过环境变量 LAB_JWT_SECRET 设置")
	}
	return nil
}

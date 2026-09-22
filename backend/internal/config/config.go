package config

import (
	"os"
	"strconv"
	"time"
)

// Config 集中管理全部服务配置，所有配置均通过环境变量注入。
type Config struct {
	ServerPort  string
	RunMode     string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	JWTSecret   string
	JWTExpire   time.Duration
	RedisAddr   string
	RedisPass   string
	RedisDB     int
	RateLimit   int
	AllowOrigins []string
}

// Load 从环境变量读取配置，并为缺失项提供默认值。
func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		RunMode:     getEnv("GIN_MODE", "release"),
		DBHost:      getEnv("DB_HOST", "127.0.0.1"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "medasset_user"),
		DBPassword:  getEnv("DB_PASSWORD", "medasset_pwd"),
		DBName:      getEnv("DB_NAME", "medasset_db"),
		JWTSecret:   getEnv("JWT_SECRET", "change_me_to_a_long_random_string"),
		JWTExpire:   time.Duration(getEnvInt("JWT_EXPIRE_HOURS", 24)) * time.Hour,
		RedisAddr:   getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:     getEnvInt("REDIS_DB", 0),
		RateLimit:   getEnvInt("RATE_LIMIT_PER_SECOND", 100),
		AllowOrigins: []string{"*"},
	}
}

// DSN 返回 MySQL 连接串。
func (c *Config) DSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName +
		"?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s"
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

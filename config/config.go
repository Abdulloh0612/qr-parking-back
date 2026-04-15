package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Telegram TelegramConfig
	App      AppConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DBConfig struct {
	URL             string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type TelegramConfig struct {
	BotToken   string
	WebhookURL string
}

type AppConfig struct {
	BaseURL     string
	Environment string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "8090")
	viper.SetDefault("SERVER_READ_TIMEOUT", "10s")
	viper.SetDefault("SERVER_WRITE_TIMEOUT", "10s")
	viper.SetDefault("DB_MAX_CONNS", 20)
	viper.SetDefault("DB_MIN_CONNS", 5)
	viper.SetDefault("DB_MAX_CONN_LIFETIME", "1h")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("JWT_ACCESS_TTL", "15m")
	viper.SetDefault("JWT_REFRESH_TTL", "720h")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_BASE_URL", "http://localhost:8090")

	_ = viper.ReadInConfig()

	readTimeout, _ := time.ParseDuration(viper.GetString("SERVER_READ_TIMEOUT"))
	writeTimeout, _ := time.ParseDuration(viper.GetString("SERVER_WRITE_TIMEOUT"))
	maxConnLifetime, _ := time.ParseDuration(viper.GetString("DB_MAX_CONN_LIFETIME"))
	accessTTL, _ := time.ParseDuration(viper.GetString("JWT_ACCESS_TTL"))
	refreshTTL, _ := time.ParseDuration(viper.GetString("JWT_REFRESH_TTL"))

	return &Config{
		Server: ServerConfig{
			Port:         viper.GetString("SERVER_PORT"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		DB: DBConfig{
			URL:             viper.GetString("DATABASE_URL"),
			MaxConns:        viper.GetInt("DB_MAX_CONNS"),
			MinConns:        viper.GetInt("DB_MIN_CONNS"),
			MaxConnLifetime: maxConnLifetime,
		},
		Redis: RedisConfig{
			URL:      viper.GetString("REDIS_URL"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:          viper.GetString("JWT_SECRET"),
			AccessTokenTTL:  accessTTL,
			RefreshTokenTTL: refreshTTL,
		},
		Telegram: TelegramConfig{
			BotToken:   viper.GetString("TG_BOT_TOKEN"),
			WebhookURL: viper.GetString("TG_WEBHOOK_URL"),
		},
		App: AppConfig{
			BaseURL:     viper.GetString("APP_BASE_URL"),
			Environment: viper.GetString("APP_ENV"),
		},
	}, nil
}

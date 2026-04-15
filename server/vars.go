package server

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Env var keys
const (
	AppNameVar    = "APP_NAME"
	AppEnvVar     = "APP_ENV"
	AppBaseURLVar = "APP_BASE_URL"

	ServerPortVar         = "SERVER_PORT"
	ServerReadTimeoutVar  = "SERVER_READ_TIMEOUT"
	ServerWriteTimeoutVar = "SERVER_WRITE_TIMEOUT"

	DatabaseURLVar        = "DATABASE_URL"
	DBMaxConnsVar         = "DB_MAX_CONNS"
	DBMinConnsVar         = "DB_MIN_CONNS"
	DBMaxConnLifetimeVar  = "DB_MAX_CONN_LIFETIME"

	RedisURLVar      = "REDIS_URL"
	RedisPasswordVar = "REDIS_PASSWORD"
	RedisDBVar       = "REDIS_DB"

	JWTSecretVar      = "JWT_SECRET"
	JWTAccessTTLVar   = "JWT_ACCESS_TTL"
	JWTRefreshTTLVar  = "JWT_REFRESH_TTL"

	TGBotTokenVar    = "TG_BOT_TOKEN"
	TGBotUsernameVar = "TG_BOT_USERNAME"
	TGWebhookURLVar  = "TG_WEBHOOK_URL"
)

// LoadEnvVars loads environment variables into memory to avoid repeated syscalls.
// Values are initialized once and only read afterwards.
func LoadEnvVars() map[string]string {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: could not load .env file, using environment variables")
	}

	keys := []string{
		AppNameVar,
		AppEnvVar,
		AppBaseURLVar,
		ServerPortVar,
		ServerReadTimeoutVar,
		ServerWriteTimeoutVar,
		DatabaseURLVar,
		DBMaxConnsVar,
		DBMinConnsVar,
		DBMaxConnLifetimeVar,
		RedisURLVar,
		RedisPasswordVar,
		RedisDBVar,
		JWTSecretVar,
		JWTAccessTTLVar,
		JWTRefreshTTLVar,
		TGBotTokenVar,
		TGBotUsernameVar,
		TGWebhookURLVar,
	}

	vars := make(map[string]string, len(keys))
	for _, k := range keys {
		vars[k] = os.Getenv(k)
	}

	// Apply defaults for missing values
	defaults := map[string]string{
		ServerPortVar:         "8090",
		ServerReadTimeoutVar:  "10s",
		ServerWriteTimeoutVar: "10s",
		DBMaxConnsVar:         "20",
		DBMinConnsVar:         "5",
		DBMaxConnLifetimeVar:  "1h",
		RedisDBVar:            "0",
		JWTAccessTTLVar:       "15m",
		JWTRefreshTTLVar:      "720h",
		AppEnvVar:             "development",
		AppBaseURLVar:         "http://localhost:8090",
	}
	for k, def := range defaults {
		if vars[k] == "" {
			vars[k] = def
		}
	}

	return vars
}

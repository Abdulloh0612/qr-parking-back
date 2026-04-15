package server

import (
	"context"

	"qr-parking/db"
	jwtpkg "qr-parking/pkg/jwt"
	"qr-parking/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Essentials holds the core initialized dependencies shared across the app.
type Essentials struct {
	Pool   *pgxpool.Pool
	Redis  *redis.Client
	Logger *zap.Logger
	JWTMgr *jwtpkg.Manager
	Vars   map[string]string
}

// NewEssentials initializes all core dependencies.
func NewEssentials() Essentials {
	vars := LoadEnvVars()

	zapLogger, err := logger.New(vars[AppEnvVar])
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	pool := db.GetPoolInstance(
		context.Background(),
		vars[DatabaseURLVar],
		vars[DBMaxConnsVar],
		vars[DBMinConnsVar],
		1,
	)

	rdb := db.GetRedisInstance(vars[RedisURLVar], vars[RedisPasswordVar], vars[RedisDBVar])

	jwtMgr := jwtpkg.NewManagerFromVars(
		vars[JWTSecretVar],
		vars[JWTAccessTTLVar],
		vars[JWTRefreshTTLVar],
	)

	return Essentials{
		Pool:   pool,
		Redis:  rdb,
		Logger: zapLogger,
		JWTMgr: jwtMgr,
		Vars:   vars,
	}
}

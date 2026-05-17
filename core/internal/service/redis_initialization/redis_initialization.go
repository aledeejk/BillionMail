package redis_initialization

import (
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"strconv"
	"time"
)

// InitRedis initialize redis configuration
func InitRedis() (err error) {
	passwd := public.MustGetDockerEnv("REDISPASS", "")

	db := 1
	dbEnv := public.MustGetDockerEnv("REDISDB", "1")
	if dbVal, parseErr := strconv.Atoi(dbEnv); parseErr == nil {
		db = dbVal
	}

	host := public.MustGetDockerEnv("REDIS_HOST", "127.0.0.1")
	port := public.MustGetDockerEnv("REDIS_PORT", "6379")
	if public.IsRunningInContainer() {
		host = public.MustGetDockerEnv("REDIS_HOST", "redis")
	}
	address := fmt.Sprintf("%s:%s", host, port)

	// Initialize Redis configuration
	gredis.SetConfig(&gredis.Config{
		Address: address,
		Db:      db,
		Pass:    passwd,
	})

	// Testing redis connection until it successful
	connectionOK := false
	for i := 0; i < 10; i++ {
		k := "bm_test_connection"
		if err := g.Redis().SetEX(context.Background(), k, k, 1); err != nil {
			g.Log().Error(context.Background(), "Redis connection test failed: ", err, " Waiting for 5 seconds before retrying...")
			time.Sleep(5 * time.Second)
			continue
		}

		if _, err := g.Redis().Del(context.Background(), k); err != nil {
			g.Log().Error(context.Background(), "Redis connection test failed: ", err, " Waiting for 5 seconds before retrying...")
			time.Sleep(5 * time.Second)
			continue
		}

		connectionOK = true
		g.Log().Debug(context.Background(), "Redis connection test successful")
		break
	}

	if !connectionOK {
		return fmt.Errorf("redis connection test failed after 10 attempts")
	}

	return nil
}

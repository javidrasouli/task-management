// @title           Task Management API
// @version         1.0
// @description     RESTful API for creating and managing tasks with priority and status tracking.
// @host            localhost:8080
// @BasePath        /api/v1

package main

import (
	"context"
	"fmt"
	"task-management/internal/cache"
	"task-management/internal/config"
	"task-management/internal/delivery/handlers"
	"task-management/internal/delivery/router"
	"task-management/internal/migration"
	"task-management/internal/repository"
	"task-management/internal/usecase"
	"task-management/internal/utils/logutil"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}

	log, err := logutil.New(cfg.App.Env)
	if err != nil {
		panic(fmt.Sprintf("init logger: %v", err))
	}
	defer log.Sync() //nolint:errcheck

	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatal("connect postgres", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("ping postgres", zap.Error(err))
	}

	if err := migration.Up(context.Background(), pool); err != nil {
		log.Fatal("run migrations", zap.Error(err))
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	repo := repository.NewPostgresRepository(pool)
	cacheLayer := cache.NewRedisCache(redisClient)
	uc := usecase.NewTaskUsecase(repo, cacheLayer)
	h := handlers.NewTaskHandler(uc)
	r := router.Setup(h, log)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("server listening", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}

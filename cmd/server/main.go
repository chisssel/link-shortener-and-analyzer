package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/artem/url-shortener/internal/config"
	"github.com/artem/url-shortener/internal/handler"
	"github.com/artem/url-shortener/internal/middleware"
	"github.com/artem/url-shortener/internal/repository"
	"github.com/artem/url-shortener/internal/service"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	pgPool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer pgPool.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	linkRepo := repository.NewPostgresRepo(pgPool)
	cacheRepo := repository.NewRedisRepo(rdb, cfg.CacheTTL)
	svc := service.NewShortenerService(linkRepo, cacheRepo)

	r := gin.Default()

	r.Use(middleware.RateLimiter(rdb))
	r.Use(middleware.CORS())

	handler.RegisterRoutes(r, svc, linkRepo)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

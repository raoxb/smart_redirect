package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/raoxb/smart_redirect/internal/config"
	"github.com/raoxb/smart_redirect/internal/database"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/local.yaml", "Path to configuration file")
	flag.Parse()
	
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	db, err := database.NewPostgresDB(&cfg.Database.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redisClient.Close()
	
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})
	
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}
	
	go func() {
		log.Printf("Server starting on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	sqlDB, _ := db.DB()
	sqlDB.Close()
	
	log.Println("Server exited")
}
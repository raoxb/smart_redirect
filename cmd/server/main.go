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
	"github.com/raoxb/smart_redirect/internal/api"
	"github.com/raoxb/smart_redirect/internal/config"
	"github.com/raoxb/smart_redirect/internal/database"
	"github.com/raoxb/smart_redirect/internal/middleware"
	"github.com/raoxb/smart_redirect/internal/services"
	"github.com/raoxb/smart_redirect/pkg/auth"
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
	
	redirectHandler := api.NewRedirectHandler(db, redisClient)
	jwtManager := auth.NewJWTManager(cfg.Security.JWTSecret, cfg.Security.JWTExpireHours)
	authHandler := api.NewAuthHandler(db, jwtManager)
	linkHandler := api.NewLinkHandler(db, redisClient)
	userHandler := api.NewUserHandler(db)
	statsHandler := api.NewStatsHandler(db, redisClient)
	batchHandler := api.NewBatchHandler(db, redisClient)
	templateHandler := api.NewTemplateHandler(db)
	monitorHandler := api.NewMonitorHandler(db, redisClient)
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})
	
	router.GET("/v1/:bu/:link_id", middleware.RateLimitMiddleware(redisClient, 1000, time.Hour), redirectHandler.HandleRedirect)
	
	apiV1 := router.Group("/api/v1")
	apiV1.Use(middleware.RateLimitMiddleware(redisClient, 100, time.Hour))
	{
		apiV1.POST("/auth/login", authHandler.Login)
		apiV1.POST("/auth/register", authHandler.Register)
		
		authGroup := apiV1.Group("/")
		authGroup.Use(middleware.AuthMiddleware(jwtManager))
		{
			authGroup.GET("/auth/profile", authHandler.GetProfile)
			
			authGroup.POST("/links", linkHandler.CreateLink)
			authGroup.GET("/links", linkHandler.ListLinks)
			authGroup.GET("/links/:link_id", linkHandler.GetLink)
			authGroup.PUT("/links/:link_id", linkHandler.UpdateLink)
			authGroup.DELETE("/links/:link_id", linkHandler.DeleteLink)
			
			authGroup.POST("/links/:link_id/targets", linkHandler.CreateTarget)
			authGroup.GET("/links/:link_id/targets", linkHandler.GetTargets)
			authGroup.PUT("/targets/:target_id", linkHandler.UpdateTarget)
			authGroup.DELETE("/targets/:target_id", linkHandler.DeleteTarget)
			
			authGroup.GET("/stats/links/:link_id", statsHandler.GetLinkStats)
			authGroup.GET("/stats/links/:link_id/hourly", statsHandler.GetHourlyStats)
			authGroup.GET("/stats/system", statsHandler.GetSystemStats)
			authGroup.GET("/stats/realtime", statsHandler.GetRealtimeStats)
			
			authGroup.POST("/batch/links", batchHandler.BatchCreateLinks)
			authGroup.PUT("/batch/links", batchHandler.BatchUpdateLinks)
			authGroup.DELETE("/batch/links", batchHandler.BatchDeleteLinks)
			authGroup.POST("/batch/import", batchHandler.ImportLinksFromCSV)
			authGroup.GET("/batch/export", batchHandler.ExportLinksToCSV)
			
			authGroup.POST("/templates", templateHandler.CreateTemplate)
			authGroup.GET("/templates", templateHandler.ListTemplates)
			authGroup.GET("/templates/:id", templateHandler.GetTemplate)
			authGroup.PUT("/templates/:id", templateHandler.UpdateTemplate)
			authGroup.DELETE("/templates/:id", templateHandler.DeleteTemplate)
			authGroup.POST("/templates/create-links", templateHandler.CreateLinksFromTemplate)
			
			adminGroup := authGroup.Group("/")
			adminGroup.Use(middleware.AdminOnly())
			{
				adminGroup.POST("/users", userHandler.CreateUser)
				adminGroup.GET("/users", userHandler.ListUsers)
				adminGroup.GET("/users/:id", userHandler.GetUser)
				adminGroup.PUT("/users/:id", userHandler.UpdateUser)
				adminGroup.DELETE("/users/:id", userHandler.DeleteUser)
				adminGroup.POST("/users/:id/links", userHandler.AssignLink)
				adminGroup.GET("/users/:id/links", userHandler.GetUserLinks)
				
				adminGroup.GET("/stats/ip/:ip", statsHandler.GetIPInfo)
				adminGroup.POST("/stats/ip/:ip/block", statsHandler.BlockIP)
				adminGroup.DELETE("/stats/ip/:ip/block", statsHandler.UnblockIP)
				
				adminGroup.GET("/monitor/alerts", monitorHandler.GetActiveAlerts)
				adminGroup.POST("/monitor/alerts/:id/acknowledge", monitorHandler.AcknowledgeAlert)
				adminGroup.POST("/monitor/alerts/:id/resolve", monitorHandler.ResolveAlert)
				adminGroup.GET("/monitor/config", monitorHandler.GetMonitoringConfig)
				adminGroup.PUT("/monitor/config", monitorHandler.UpdateMonitoringConfig)
				adminGroup.GET("/monitor/health", monitorHandler.GetHealthStatus)
			}
		}
	}
	
	// Start monitoring service
	monitorService := services.NewMonitorService(db, redisClient)
	monitorCtx, cancelMonitor := context.WithCancel(context.Background())
	go monitorService.StartMonitoring(monitorCtx)
	log.Println("Monitoring service started")
	
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
	
	// Stop monitoring service
	cancelMonitor()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	sqlDB, _ := db.DB()
	sqlDB.Close()
	
	log.Println("Server exited")
}
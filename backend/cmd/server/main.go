package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/guvi-internship/polling-backend/internal/config"
	"github.com/guvi-internship/polling-backend/internal/db"
	"github.com/guvi-internship/polling-backend/internal/handlers"
	"github.com/guvi-internship/polling-backend/internal/middleware"
	"github.com/guvi-internship/polling-backend/internal/ws"
)

func main() {
	cfg := config.Load()

	mongo, err := db.ConnectMongo(cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("mongo connection failed: %v", err)
	}
	log.Println("connected to MongoDB")

	redisClient, err := db.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	log.Println("connected to Redis")

	hub := ws.NewHub(redisClient)
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)

	app := &handlers.App{
		Mongo:     mongo,
		Redis:     redisClient,
		Hub:       hub,
		JWTSecret: cfg.JWTSecret,
		TokenTTLH: cfg.TokenTTLHours,
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigin},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		api.POST("/auth/signup", app.Signup)
		api.POST("/auth/login", app.Login)

		api.GET("/polls/:id", app.GetPoll)
		

		authed := api.Group("/")
		authed.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			authed.POST("polls", app.CreatePoll)
			authed.GET("polls", app.ListMyPolls)
			authed.PATCH("polls/:id/close", app.ClosePoll)
			authed.POST("polls/:id/vote", app.CastVote)
		}
	}

	router.GET("/ws/polls/:id", app.PollSocket)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	cancel() // stop the Redis pub/sub hub

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	_ = mongo.Client.Disconnect(shutdownCtx)
}

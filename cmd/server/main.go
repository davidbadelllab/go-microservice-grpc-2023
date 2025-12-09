package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/davidbadelllab/go-microservice-grpc-2023/internal/config"
	"github.com/davidbadelllab/go-microservice-grpc-2023/internal/repository"
	"github.com/davidbadelllab/go-microservice-grpc-2023/internal/server"
	"github.com/davidbadelllab/go-microservice-grpc-2023/internal/service"
	"github.com/davidbadelllab/go-microservice-grpc-2023/pkg/cache"
	"github.com/davidbadelllab/go-microservice-grpc-2023/pkg/database"
	"github.com/davidbadelllab/go-microservice-grpc-2023/pkg/logger"
	pb "github.com/davidbadelllab/go-microservice-grpc-2023/proto"
)

func main() {
	// Initialize logger
	log := logger.New()
	slog.SetDefault(log)

	slog.Info("starting gRPC server",
		slog.String("service", "user-service"),
		slog.String("version", "1.0.0"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize database
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize cache
	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		slog.Error("failed to connect to redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer redisClient.Close()

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Initialize service
	userService := service.NewUserService(userRepo, redisClient)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.LoggingInterceptor,
			server.MetricsInterceptor,
			server.RecoveryInterceptor,
		),
	)

	// Register services
	userServer := server.NewUserServer(userService)
	pb.RegisterUserServiceServer(grpcServer, userServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("user-service", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		slog.Info("metrics server starting", slog.Int("port", cfg.MetricsPort))
		if err := http.ListenAndServe(":9090", nil); err != nil {
			slog.Error("metrics server failed", slog.String("error", err.Error()))
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		slog.Error("failed to listen", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		slog.Info("gRPC server listening", slog.String("address", cfg.GRPCAddress))
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully stop gRPC server
	grpcServer.GracefulStop()

	// Close database connection
	db.Close()

	slog.Info("server stopped", slog.String("context", ctx.Err().Error()))
}

package main

import (
	"context"
	"fileStorage/internal/api"
	"fileStorage/internal/config"
	"fileStorage/internal/database"
	"fileStorage/internal/logger"
	gRPC "fileStorage/internal/proto"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv := grpc.NewServer()
	cfg := config.NewConfig()
	zapLog, err := logger.NewLogger(cfg.LogFilePath, &cfg.Level)
	if err != nil {
		log.Fatal("failed to create logs")
	}
	server := &api.Server{Cfg: &cfg.ServerConfig, Logger: zapLog}
	gRPC.RegisterFileStorageServer(srv, server)
	wg.Add(1)
	go database.Broadcast(ctx, &wg, &cfg.DbConfig, srv, zapLog)
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		zapLog.Fatal("main", zap.String("description", "failed to listen"), zap.String("error", err.Error()))
	}
	zapLog.Info("main", zap.String("description", fmt.Sprintf("server listening at %v", lis.Addr())), zap.String("error", "nil"))
	if err := srv.Serve(lis); err != nil {
		zapLog.Fatal("main", zap.String("description", "failed to listen"), zap.String("error", err.Error()))
	}
	wg.Wait()
}

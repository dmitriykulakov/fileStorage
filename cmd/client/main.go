package main

import (
	"fileStorage/internal/api"
	"fileStorage/internal/config"
	gRPC "fileStorage/internal/proto"
	"log"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	cfg := config.NewConfig()
	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error connecting to server: %v", err.Error())
	}
	defer conn.Close()
	client := gRPC.NewFileStorageClient(conn)
	api.Auth(&cfg.ServerConfig, &client)
}

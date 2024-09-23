package main

import (
	enter "fileStorage/cmd/client/auth"
	m "fileStorage/cmd/client/mainMenu"
	"fileStorage/config"
	gRPC "fileStorage/proto"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.ConfigLoad()
	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}
	defer conn.Close()
	client := gRPC.NewFileStorageClient(conn)
	if resp, ok := enter.Auth(&client); ok {
		m.MainMenu(&cfg, &client, resp)
	}
}

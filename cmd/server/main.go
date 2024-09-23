package main

import (
	"context"
	db "fileStorage/cmd/server/database"
	api "fileStorage/cmd/server/serverAPI"
	"fileStorage/config"
	gRPC "fileStorage/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var srv *grpc.Server

func main() {
	var wg sync.WaitGroup
	cfg := config.ConfigLoad()
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv = grpc.NewServer()
	log.Printf("server listening at %v", lis.Addr())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	wg.Add(1)
	go broadcast(ctx, &wg)
	gRPC.RegisterFileStorageServer(srv, &api.Server{})
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	wg.Wait()
	log.Println("The server is stopped")
}

func broadcast(ctx context.Context, wg *sync.WaitGroup) {
	dataBase := db.ConnectToDB()
	for {
		select {
		case <-ctx.Done():
			srv.Stop()
			wg.Done()
			return
		case client := <-api.LoginCh:
			api.DbResponse <- client.Login(dataBase)
		case client := <-api.RegCh:
			api.DbResponse <- client.Reg(dataBase)
		default:
			time.Sleep(time.Duration(time.Millisecond))
		}
	}
}

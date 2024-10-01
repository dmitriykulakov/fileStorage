package main

import (
	"context"
	db "fileStorage/cmd/server/database"
	l "fileStorage/cmd/server/log"
	api "fileStorage/cmd/server/serverAPI"
	"fileStorage/config"
	gRPC "fileStorage/proto"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

var srv *grpc.Server

func main() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go l.Logger(ctx, &wg)
	go broadcast(ctx, &wg)
	wg.Add(2)

	cfg := config.ConfigLoad()
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		l.ChLog <- l.Log{Message: fmt.Sprintf("failed to listen: %v", err), Level: "fatal"}
	}

	srv = grpc.NewServer()
	l.ChLog <- l.Log{Message: fmt.Sprintf("server listening at %v", lis.Addr()), Level: ""}

	gRPC.RegisterFileStorageServer(srv, &api.Server{})
	if err := srv.Serve(lis); err != nil {
		l.ChLog <- l.Log{Message: fmt.Sprintf("failed to serve: %v", err), Level: "fatal"}
	}

	wg.Wait()
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

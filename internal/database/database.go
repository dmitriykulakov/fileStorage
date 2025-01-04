package database

import (
	"context"
	"errors"
	"fileStorage/internal/config"
	"fileStorage/internal/models"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

var LoginCh = make(chan models.Clients)
var RegCh = make(chan models.Clients)
var AuthResponseCh = make(chan error)

type db interface {
	login(client *models.Clients) error
	reg(client *models.Clients) error
}

func Broadcast(ctx context.Context, wg *sync.WaitGroup, cfg *config.DbConfig, srv *grpc.Server, zapLog *zap.Logger) {
	defer wg.Done()
	var db db
	db, err := connectToPgDB(cfg)
	if err != nil {
		zapLog.Fatal("broadcast", zap.String("description", "failed to connect to database"), zap.String("error", err.Error()))
	}
	for {
		select {
		case <-ctx.Done():
			srv.Stop()
			return
		case client := <-LoginCh:
			select {
			case AuthResponseCh <- db.login(&client):
			case <-time.After(time.Microsecond * 200):
				AuthResponseCh <- errors.New("server is not responding")
			}
		case client := <-RegCh:
			select {
			case AuthResponseCh <- db.reg(&client):
			case <-time.After(time.Microsecond * 200):
				AuthResponseCh <- errors.New("server is not responding")
			}
		default:
			time.Sleep(time.Duration(time.Millisecond))
		}
	}
}

func StoragePassword(password string) ([]byte, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return passHash, nil
}

func comparePassword(passHash []byte, databasePass []byte) bool {
	if err := bcrypt.CompareHashAndPassword(passHash, databasePass); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

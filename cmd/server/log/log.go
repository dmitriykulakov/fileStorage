package log

import (
	"context"
	"log"
	"os"
	"sync"
	"time"
)

type Log struct {
	Message string
	Level   string
}

var ChLog = make(chan Log)

const logFile = "./cmd/server/log/server.log"

func Logger(ctx context.Context, wg *sync.WaitGroup) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal("error to open the server.log file", err)
	}
	defer file.Close()
	for {
		select {
		case <-ctx.Done():
			if _, err = file.WriteString(time.Now().String() + " " + "server stopped\n"); err != nil {
				log.Fatal(err)
			}
			log.Println("server stopped")
			wg.Done()
			return

		case Log := <-ChLog:
			if _, err = file.WriteString(time.Now().String() + " " + Log.Message + "\n"); err != nil {
				log.Fatal("error to write the server.log file")
			}
			if Log.Level == "fatal" {
				log.Fatalf(Log.Message)
				return
			}
			log.Println(Log.Message)
		default:
			time.Sleep(time.Duration(time.Millisecond))
		}
	}
}

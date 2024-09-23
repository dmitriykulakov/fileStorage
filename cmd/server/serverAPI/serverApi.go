package serverAPI

import (
	"context"
	db "fileStorage/cmd/server/database"
	"fileStorage/config"
	gRPC "fileStorage/proto"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const (
	storagePath = "./cmd/server/serverStorage/"
)

var LoginCh = make(chan db.Clients)
var RegCh = make(chan db.Clients)
var DbResponse = make(chan string)

type Server struct {
	gRPC.FileStorageServer
}

func (s *Server) LogIn(_ context.Context, r *gRPC.LogInRequest) (*gRPC.LogInResponse, error) {
	LoginCh <- db.Clients{Name: r.Login, Password: r.Password}
	return &gRPC.LogInResponse{Response: <-DbResponse}, nil
}

func (s *Server) Reg(_ context.Context, r *gRPC.RegRequest) (*gRPC.RegResponse, error) {
	RegCh <- db.Clients{Name: r.Login, Password: r.Password}
	return &gRPC.RegResponse{Response: <-DbResponse}, nil
}

func (s *Server) GetFiles(r *gRPC.GetFilesRequest, t gRPC.FileStorage_GetFilesServer) error {
	dir, err := os.Open(storagePath)
	if err != nil {
		log.Fatalf("GetFiles: %v", err)
	}
	defer dir.Close()
	dirContain, err := dir.Readdir(-2)
	if err != nil {
		log.Fatalf("GetFiles: %s", err)
	}
	for _, file := range dirContain {
		resp := &gRPC.GetFilesResponse{
			Response: file.Name(),
		}
		err := t.Send(resp)
		if err != nil {
			return err
		}
	}
	log.Printf("GetFiles: user %s - OK", r.Client)
	return nil
}

func (s *Server) GetFile(r *gRPC.GetFileRequest, t gRPC.FileStorage_GetFileServer) error {
	cfg := config.ConfigLoad()
	file, err := os.Open(storagePath + r.Filename)
	if err != nil {
		err := fmt.Errorf("the file \"%s\" is not exist: %v", r.Filename, err)
		log.Printf("GetFile: user %s: %v", r.Client, err)
		return err
	}
	buf := make([]byte, cfg.MaxByteSend)
	pos, err := file.Read(buf)
	for err != io.EOF {
		if pos < cfg.MaxByteSend {
			buf = buf[:pos]
		}
		resp := &gRPC.GetFileResponse{
			Filename: r.Filename,
			FileData: buf,
		}
		err = t.Send(resp)
		if err != nil {
			return err
		}
		pos, err = file.Read(buf)
	}
	log.Printf("GetFile: user %s: the file \"%s\" is sent, OK", r.Client, r.Filename)
	return nil
}

func (s *Server) PostFile(t gRPC.FileStorage_PostFileServer) error {
	var fileCreate *os.File
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	flag := true
	flagCreated := false
	defer fileCreate.Close()
	for flag {
		if file, err := t.Recv(); err == nil {
			if !flagCreated {
				if _, err = os.Open((storagePath + file.Filename)); err == nil {
					err := fmt.Errorf("the file \"%s\" is already exist ", file.Filename)
					log.Printf("PostFile: user %s: %v", file.Client, err)
					return err
				}
				fileCreate, err = os.Create(storagePath + file.Filename)
				if err != nil {
					err := fmt.Errorf("error  with create the file \"%s\"", file.Filename)
					log.Printf("PostFile: user %s: %v", file.Client, err)
					return err
				}
				flagCreated = true
				log.Printf("user %s: Created new file %s\n", file.Client, file.Filename)
			}
			if _, err = fileCreate.Write([]byte(file.FileData)); err != nil {
				fmt.Printf("user %s: error with write the file \"%s\": %v\n", file.Client, file.Filename, err)
				return err
			}
		} else {
			flag = false
			if err != io.EOF {
				fmt.Printf("user %s: Файла \"%s\" нет в хранилище\n, %v", file.Client, file.Filename, err)
			}
		}
	}
	return nil
}

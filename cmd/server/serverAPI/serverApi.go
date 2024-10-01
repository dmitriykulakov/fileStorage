package serverAPI

import (
	"context"
	db "fileStorage/cmd/server/database"
	l "fileStorage/cmd/server/log"
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
		l.ChLog <- l.Log{Message: fmt.Sprintf("GetFiles: %s", err), Level: "fatal"}
	}
	defer dir.Close()
	dirContain, err := dir.Readdir(-2)
	if err != nil {
		l.ChLog <- l.Log{Message: fmt.Sprintf("GetFiles: %s", err), Level: "fatal"}
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
	l.ChLog <- l.Log{Message: fmt.Sprintf("GetFiles: user %s - OK", r.Client), Level: ""}
	return nil
}

func (s *Server) GetFile(r *gRPC.GetFileRequest, t gRPC.FileStorage_GetFileServer) error {
	cfg := config.ConfigLoad()
	file, err := os.Open(storagePath + r.Filename)
	if err != nil {
		err := fmt.Errorf("the file \"%s\" is not exist: %v", r.Filename, err)
		l.ChLog <- l.Log{Message: fmt.Sprintf("GetFile: user %s: %v", r.Client, err), Level: ""}
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
	l.ChLog <- l.Log{Message: fmt.Sprintf("GetFile: user %s: the file \"%s\" is sent, OK", r.Client, r.Filename), Level: ""}
	return nil
}

func (s *Server) PostFile(t gRPC.FileStorage_PostFileServer) error {
	var filename string
	var client string
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
					l.ChLog <- l.Log{Message: fmt.Sprintf("PostFile: user %s: %v", file.Client, err), Level: ""}
					return err
				}
				fileCreate, err = os.Create(storagePath + file.Filename)
				if err != nil {
					err := fmt.Errorf("error  with create the file \"%s\"", file.Filename)
					l.ChLog <- l.Log{Message: fmt.Sprintf("PostFile: user %s: %v", file.Client, err), Level: ""}
					return err
				}
				client = file.Client
				filename = file.Filename
				flagCreated = true
				l.ChLog <- l.Log{Message: fmt.Sprintf("PostFile: user %s: Created new file %s", file.Client, file.Filename), Level: ""}
			}
			if _, err = fileCreate.Write([]byte(file.FileData)); err != nil {
				l.ChLog <- l.Log{Message: fmt.Sprintf("PostFile: user %s: error with write the file \"%s\": %v", file.Client, file.Filename, err), Level: ""}
				return err
			}
		} else {
			flag = false
			if err != io.EOF {
				log.Println(err)
				return err
			}
			l.ChLog <- l.Log{Message: fmt.Sprintf("PostFile: user %s: the file %s is saved", client, filename), Level: ""}
			t.SendAndClose(&gRPC.PostFileResponse{
				Response: "the file " + filename + " is saved",
			})
			return nil
		}
	}
	return nil
}

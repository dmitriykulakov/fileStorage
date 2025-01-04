package api

import (
	"context"
	"fileStorage/internal/config"
	"fileStorage/internal/database"
	"fileStorage/internal/jwt"
	"fileStorage/internal/models"
	gRPC "fileStorage/internal/proto"
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
)

type Server struct {
	gRPC.FileStorageServer
	Cfg    *config.ServerConfig
	Logger *zap.Logger
}

func (s *Server) LogIn(_ context.Context, r *gRPC.LogInRequest) (*gRPC.LogInResponse, error) {
	database.LoginCh <- models.Clients{Name: r.Login, HashPassword: r.Password}
	if err := <-database.AuthResponseCh; err != nil {
		s.Logger.Info("api_login", zap.String("error", err.Error()))
		return nil, err
	}
	token, err := jwt.NewToken(r.Login, s.Cfg.TokenTTL)
	if err != nil {
		s.Logger.Info("api_login",
			zap.String("description", fmt.Sprintf("Пользователь %s - ошибка создания токена", r.Login)),
			zap.String("error", err.Error()))
	}
	s.Logger.Info("api_login",
		zap.String("description", fmt.Sprintf("Пользователь %s - успешный вход", r.Login)))
	return &gRPC.LogInResponse{Token: token}, nil
}

func (s *Server) Reg(_ context.Context, r *gRPC.RegRequest) (*gRPC.RegResponse, error) {
	hashPassword, err := database.StoragePassword(r.Password)
	if err != nil {
		s.Logger.Info("api_reg",
			zap.String("description", fmt.Sprintf("Пользователь %s", r.Login)),
			zap.String("error", err.Error()))
		return nil, err
	}
	database.RegCh <- models.Clients{Name: r.Login, HashPassword: (string)(hashPassword)}
	if err := <-database.AuthResponseCh; err != nil {
		s.Logger.Info("api_reg",
			zap.String("error", err.Error()))
		return nil, err
	}
	token, err := jwt.NewToken(r.Login, s.Cfg.TokenTTL)
	if err != nil {
		s.Logger.Info("api_reg",
			zap.String("description", fmt.Sprintf("Пользователь %s - ошибка создания токена", r.Login)),
			zap.String("error", err.Error()))
	}
	if err := os.Mkdir(s.Cfg.StoragePath+r.Login, 0777); err != nil {
		s.Logger.Fatal("api_reg",
			zap.String("description", fmt.Sprintf("Пользователь %s - ошибка создания директории", r.Login)),
			zap.String("error", err.Error()))
		return nil, err
	}
	s.Logger.Info("api_reg",
		zap.String("description", fmt.Sprintf("Пользователь %s - успешная регистрация", r.Login)))
	return &gRPC.RegResponse{Token: token}, nil
}

func (s *Server) GetFiles(r *gRPC.GetFilesRequest, t gRPC.FileStorage_GetFilesServer) error {
	user, err := jwt.CheckToken(r.Token)
	if err != nil {
		s.Logger.Info("api_get_files",
			zap.String("error", err.Error()))
		return err
	}
	dir, err := os.Open(s.Cfg.StoragePath + user)
	if err != nil {
		s.Logger.Info("api_get_files",
			zap.String("description", fmt.Sprintf("user %s", user)),
			zap.String("error", err.Error()))
		return err
	}
	defer dir.Close()
	dirContain, err := dir.Readdir(-2)
	if err != nil {
		s.Logger.Info("api_get_files",
			zap.String("description", fmt.Sprintf("user %s", user)),
			zap.String("error", err.Error()))
		return err
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
	s.Logger.Info("api_get_files",
		zap.String("description", fmt.Sprintf("user %s OK", user)))
	return nil
}

func (s *Server) GetFile(r *gRPC.GetFileRequest, t gRPC.FileStorage_GetFileServer) error {
	user, err := jwt.CheckToken(r.Token)
	if err != nil {
		s.Logger.Info("api_get_file",
			zap.String("error", err.Error()))
		return err
	}
	file, err := os.Open(s.Cfg.StoragePath + user + "/" + r.Filename)
	if err != nil {
		err := fmt.Errorf("the file %s is not exist: %v", r.Filename, err)
		s.Logger.Info("api_get_file",
			zap.String("description", fmt.Sprintf("user %s", user)),
			zap.String("error", err.Error()))
		return err
	}
	buf := make([]byte, s.Cfg.MaxByteSend)
	pos, err := file.Read(buf)
	for err != io.EOF {
		if pos < s.Cfg.MaxByteSend {
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
	s.Logger.Info("api_get_file",
		zap.String("description", fmt.Sprintf("user %s: the file %s is sent, OK", user, r.Filename)))
	return nil
}

func (s *Server) PostFile(t gRPC.FileStorage_PostFileServer) error {
	var filename string
	var fileCreate *os.File
	var user string
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	flag := true
	flagCreated := false
	defer fileCreate.Close()
	for flag {
		if file, err := t.Recv(); err == nil {
			if !flagCreated {
				user, err = jwt.CheckToken(file.Token)
				if err != nil {
					s.Logger.Info("api_post_file",
						zap.String("error", err.Error()))
					return err
				}
				if _, err = os.Open((s.Cfg.StoragePath + user + "/" + file.Filename)); err == nil {
					err := fmt.Errorf("the file %s is already exist ", file.Filename)
					s.Logger.Info("api_post_file",
						zap.String("description", fmt.Sprintf("user %s", user)),
						zap.String("error", err.Error()))
					return err
				}
				fileCreate, err = os.Create(s.Cfg.StoragePath + user + "/" + file.Filename)
				if err != nil {
					err := fmt.Errorf("error  with create the file %s", file.Filename)
					s.Logger.Info("api_post_file",
						zap.String("description", fmt.Sprintf("user %s", user)),
						zap.String("error", err.Error()))
					return err
				}
				filename = file.Filename
				flagCreated = true
				s.Logger.Info("api_post_file",
					zap.String("description", fmt.Sprintf("user %s: Created new file %s", user, file.Filename)))
			}
			if _, err = fileCreate.Write([]byte(file.FileData)); err != nil {
				s.Logger.Info("api_post_file",
					zap.String("description", fmt.Sprintf("user %s: error with write the file %s", user, file.Filename)),
					zap.String("error", err.Error()))
				return err
			}
		} else {
			flag = false
			if err != io.EOF {
				s.Logger.Info("api_post_file",
					zap.String("error", err.Error()))
				return err
			}
			s.Logger.Info("api_post_file",
				zap.String("description", fmt.Sprintf(" user %s: the file %s is saved", user, filename)))
			t.SendAndClose(&gRPC.PostFileResponse{
				Response: "the file " + filename + " is saved",
			})
		}
	}
	return nil
}

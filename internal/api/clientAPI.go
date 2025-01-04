package api

import (
	"context"
	"fileStorage/internal/config"
	gRPC "fileStorage/internal/proto"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Auth(cfg *config.ServerConfig, client *gRPC.FileStorageClient) (string, bool) {
	var result string
	fmt.Println("Программа fileStorage приветствует Вас\nСпасибо, что воспользовались нашим приложением")
	for {
		fmt.Println("Для продолжения необходимо авторизоваться:\n\t1:\tВойти\n\t2:\tЗарегестрироваться\n\t3:\tВыйти")
		fmt.Scan(&result)
		switch result {
		case "1":
			token, ok := login(client)
			if ok {
				mainMenu(cfg, client, token)
			}
			continue
		case "2":
			token, ok := registration(client)
			if ok {
				mainMenu(cfg, client, token)
			}
			continue
		case "3":
			fmt.Println("Good luck")
			return "", false
		default:
			fmt.Printf("The \"%s\" is wrong\n", result)
			continue
		}
	}
}

func login(client *gRPC.FileStorageClient) (string, bool) {
	var login string
	var password string
	for {
		fmt.Println("Введите логин:")
		fmt.Scan(&login)
		fmt.Println("Введите пароль:")
		fmt.Scan(&password)
		resp, err := (*client).LogIn(context.Background(), &gRPC.LogInRequest{Login: login, Password: password})
		if err != nil {
			fmt.Println(err)
			fmt.Println("\t1:\tПопробовать еще раз\n\t2:\tНазад")
			fmt.Scan(&login)
			if login != "1" {
				return "", false
			}
		} else {
			fmt.Printf("Успешный вход под именем %s", login)
			return resp.Token, true
		}
	}
}

func registration(client *gRPC.FileStorageClient) (string, bool) {
	var login string
	var password string
	for {
		fmt.Println("Введите логин:")
		fmt.Scan(&login)
		fmt.Println("Введите пароль:")
		fmt.Scan(&password)
		resp, err := (*client).Reg(context.Background(), &gRPC.RegRequest{Login: login, Password: password})
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("\t1:\t Попробовать еще раз\n\t2\tНазад")
			fmt.Scan(&login)
			if login != "1" {
				return "", false
			}
		} else {
			fmt.Printf("Успешная регистрация под именем %s", login)
			return resp.Token, true
		}
	}
}

const downloadPath = "./cmd/client/download/"

func mainMenu(cfg *config.ServerConfig, client *gRPC.FileStorageClient, token string) {
	var result string
	for {
		fmt.Println("\nГлавное меню:\n\t1:\tПолучить список файлов\n\t2:\tЗагрузить файл в хранилище\n\t3:\tСкачать файл с хранилища\n\t4:\tВыход")
		fmt.Scan(&result)
		switch result {
		case "1":
			if err := getFiles(client, token); err != nil {
				return
			}
		case "2":
			if err := postFile(client, cfg, token); err != nil {
				return
			}
		case "3":
			if err := getFile(client, token); err != nil {
				return
			}
		case "4":
			return
		default:
			fmt.Printf("The \"%s\" is wrong\n", result)
			continue
		}
	}
}

func getFiles(client *gRPC.FileStorageClient, token string) error {
	resp, err := (*client).GetFiles(context.Background(), &gRPC.GetFilesRequest{Token: token})
	if err != nil {
		log.Printf("error sending request: %v", err)
		return err
	}
	flag := true
	fmt.Println("Список файлов:")
	for flag {
		if filename, err := resp.Recv(); err == nil {
			fmt.Println("\t" + filename.Response)
		} else {
			flag = false
			if err != io.EOF {
				log.Print("ERROR:", err)
				return err
			}
		}
	}
	return nil
}

func getFile(client *gRPC.FileStorageClient, token string) error {
	var filename string
	fmt.Println("Введите название файла")
	fmt.Scan(&filename)
	resp, err := (*client).GetFile(context.Background(), &gRPC.GetFileRequest{Filename: filename, Token: token})
	if err != nil {
		fmt.Printf("error sending request: %v\n", err)
		return filepath.ErrBadPattern
	}
	flag := true
	flagCreated := false
	var fileCreate *os.File
	defer fileCreate.Close()
	for flag {
		if file, err := resp.Recv(); err == nil {
			if !flagCreated {
				fileCreate, err = os.Create(downloadPath + file.Filename)
				if err != nil {
					fmt.Printf("Ошибка создания файла: %v\n", err)
					return nil
				}
				flagCreated = true
				fmt.Printf("Файл \"%s\" сохранен в папку downloads", filename)
			}
			if _, err = fileCreate.Write([]byte(file.FileData)); err != nil {
				fmt.Printf("error with write the file \"%s\": %v\n", file.Filename, err)
				return nil
			}
		} else {
			flag = false
			fmt.Println()
			if err != io.EOF {
				fmt.Printf("Файла \"%s\" нет в хранилище\n, %v", filename, err)
			}
		}
	}
	return nil
}

func postFile(client *gRPC.FileStorageClient, cfg *config.ServerConfig, token string) error {
	var filename string
	fmt.Println("Загрузить файл в хранилище:\n\tВВедите полный путь к файлу на вашем компьютере:")
	fmt.Scan(&filename)
	response, err := os.Stat(filename)
	if err != nil || !(response.Mode().IsRegular()) {
		fmt.Printf("the  \"%s\" is not a file: %v\n", filename, err)
		return nil
	}
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("the file \"%s\" is not exist: %v\n", filename, err)
		return nil
	}
	defer file.Close()
	filename = filepath.Base(filename)
	buf := make([]byte, cfg.MaxByteSend)
	pos, errRead := file.Read(buf)
	clientSTR, err := (*client).PostFile(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		resp, err := clientSTR.CloseAndRecv()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(resp.Response)
		}
	}()
	for errRead != io.EOF {
		if pos < cfg.MaxByteSend {
			buf = buf[:pos]
		}
		request := &gRPC.PostFileRequest{
			Filename: filename,
			FileData: buf,
			Token:    token,
		}
		err := clientSTR.Send(request)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return nil
		}
		pos, errRead = file.Read(buf)
	}
	return nil
}

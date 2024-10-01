package main_menu

import (
	"context"
	"fileStorage/config"
	gRPC "fileStorage/proto"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const downloadPath = "./cmd/client/download/"

func MainMenu(cfg *config.HTTPServer, client *gRPC.FileStorageClient, clientName string) {
	var result string
	for {
		fmt.Println("\nГлавное меню:\n\t1:\tПолучить список файлов\n\t2:\tЗагрузить файл в хранилище\n\t3:\tСкачать файл с хранилища\n\t4:\tВыход")
		fmt.Scan(&result)
		switch result {
		case "1":
			getFiles(client, clientName)
		case "2":
			postFile(client, cfg, clientName)
		case "3":
			getFile(client, clientName)
		case "4":
			return
		default:
			fmt.Printf("The \"%s\" is wrong\n", result)
			continue
		}
	}
}

func getFiles(client *gRPC.FileStorageClient, clientName string) {
	resp, err := (*client).GetFiles(context.Background(), &gRPC.GetFilesRequest{Client: clientName})
	if err != nil {
		log.Fatalf("error sending request: %v", err)
	}
	flag := true
	fmt.Println("Список файлов:")
	for flag {
		if filename, err := resp.Recv(); err == nil {
			fmt.Println("\t" + filename.Response)
		} else {
			flag = false
			if err != io.EOF {
				log.Fatal("ERROR:", err)
			}
		}
	}
}

func getFile(client *gRPC.FileStorageClient, clientName string) {
	var filename string
	fmt.Println("Введите название файла")
	fmt.Scan(&filename)
	resp, err := (*client).GetFile(context.Background(), &gRPC.GetFileRequest{Filename: filename, Client: clientName})
	if err != nil {
		fmt.Printf("error sending request: %v\n", err)
		return
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
					return
				}
				flagCreated = true
				fmt.Printf("Файл \"%s\" сохранен в папку downloads", filename)
			}
			if _, err = fileCreate.Write([]byte(file.FileData)); err != nil {
				fmt.Printf("error with write the file \"%s\": %v\n", file.Filename, err)
				return
			}
		} else {
			flag = false
			fmt.Println()
			if err != io.EOF {
				fmt.Printf("Файла \"%s\" нет в хранилище\n, %v", filename, err)
			}
		}
	}
}

func postFile(client *gRPC.FileStorageClient, cfg *config.HTTPServer, clientName string) {
	var filename string
	fmt.Println("Загрузить файл в хранилище:\n\tВВедите полный путь к файлу на вашем компьютере:")
	fmt.Scan(&filename)
	response, err := os.Stat(filename)
	if err != nil || !(response.Mode().IsRegular()) {
		fmt.Printf("the  \"%s\" is not a file: %v\n", filename, err)
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("the file \"%s\" is not exist: %v\n", filename, err)
		return
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
			Client:   clientName,
		}
		err := clientSTR.Send(request)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		}
		pos, errRead = file.Read(buf)
	}
}

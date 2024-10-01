package auth

import (
	"context"
	c "fileStorage/config"
	gRPC "fileStorage/proto"
	"fmt"
)

func Auth(client *gRPC.FileStorageClient) (string, bool) {
	var result string
	fmt.Println("Программа fileStorage приветствует Вас\nСпасибо, что воспользовались нашим приложением")
	for {
		fmt.Println("Для продолжения необходимо авторизоваться:\n\t1:\tВойти\n\t2:\tЗарегестрироваться\n\t3:\tВыйти")
		fmt.Scan(&result)
		switch result {
		case "1":
			user, ok := Login(client)
			if ok {
				return user, true
			}
			continue
		case "2":
			user, ok := Registration(client)
			if ok {
				return user, true
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

func Login(client *gRPC.FileStorageClient) (string, bool) {
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
			return "", false
		}
		if resp.Response == c.LogResp {
			fmt.Println(resp.Response, login)
			return login, true
		} else {
			fmt.Println(resp.Response)
			fmt.Println("\t1:\tПопробовать еще раз\n\t2:\tНазад")
			fmt.Scan(&login)
			if login != "1" {
				return "", false
			}
		}
	}
}

func Registration(client *gRPC.FileStorageClient) (string, bool) {
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
			return "", false
		}
		if resp.Response == c.RegResp {
			fmt.Println(resp.Response, login)
			return login, true
		} else {
			fmt.Println(resp.Response)
			fmt.Println("\t1:\t Попробовать еще раз\n\t2\tНазад")
			fmt.Scan(&login)
			if login != "1" {
				return "", false
			}
		}
	}
}

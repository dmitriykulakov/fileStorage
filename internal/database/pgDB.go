package database

import (
	"fileStorage/internal/config"
	"fileStorage/internal/models"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const LogResp = "Добро пожаловать в fileStorage "
const RegResp = "Вы успешно зарегестрированы под именем "

type pg struct {
	pg *gorm.DB
}

func connectToPgDB(cfg *config.DbConfig) (*pg, error) {
	cfgPG := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port)
	db, err := gorm.Open(postgres.Open(cfgPG), &gorm.Config{})
	for i := 0; i < 10 && err != nil; i++ {
		time.Sleep(time.Second * 5)
		db, err = gorm.Open(postgres.Open(cfgPG), &gorm.Config{})
	}
	if err != nil {
		return nil, fmt.Errorf("ConnectToDB: error to connect %v", err)
	}
	db.AutoMigrate(&models.Clients{})
	return &pg{db}, nil
}

func (db *pg) login(client *models.Clients) error {
	var clients []models.Clients
	db.pg.Table("clients").Where("name = ?", client.Name).Find(&clients)
	if len(clients) == 0 {
		return fmt.Errorf("пользователь \"%s\" не зарегестрирован", client.Name)
	}
	if comparePassword([]byte(clients[0].HashPassword), []byte(client.HashPassword)) {
		return nil
	}
	return fmt.Errorf("неверный пароль для пользователя \"%s\"", client.Name)
}

func (db *pg) reg(client *models.Clients) error {
	var clients []models.Clients
	db.pg.Table("clients").Where("name = ?", client.Name).Find(&clients)
	if len(clients) != 0 {
		return fmt.Errorf("пользователь \"%s\" уже зарегестрирован", client.Name)
	}
	db.pg.Table("clients").Select("Name", "HashPassword").Create(client)
	return nil
}

package database

import (
	l "fileStorage/cmd/server/log"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	c "fileStorage/config"
)

const configPath = "./cmd/server/database/config.yaml"

type Clients struct {
	Name     string `gorm:"primaryKey"`
	Password string
}

type PgConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"dbname"`
}

func ConnectToDB() *gorm.DB {
	var pg PgConfig
	pg.GetConf()
	cfg := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", pg.Host, pg.Username, pg.Password, pg.Database, pg.Port)
	db, err := gorm.Open(postgres.Open(cfg), &gorm.Config{})
	for i := 0; i < 5 && err != nil; i++ {
		time.Sleep(time.Second * 5)
		db, err = gorm.Open(postgres.Open(cfg), &gorm.Config{})
		if err != nil {
			l.ChLog <- l.Log{Message: fmt.Sprintf("ConnectToDB: error to connect, please wait %v", err), Level: "fatal"}
		}
	}
	if err != nil {
		l.ChLog <- l.Log{Message: fmt.Sprintf("ConnectToDB: error to connect %v", err), Level: "fatal"}
	}
	db.AutoMigrate(&Clients{})
	return db
}

func (p *PgConfig) GetConf() *PgConfig {
	conf, err := os.ReadFile(configPath)
	if err != nil {
		l.ChLog <- l.Log{Message: fmt.Sprintf("GetConf: file not found: %v", err), Level: "fatal"}
	}
	err = yaml.Unmarshal(conf, p)
	if err != nil {
		l.ChLog <- l.Log{Message: fmt.Sprintf("GetConf: error unmarshalling yaml file: %v", err), Level: "fatal"}
	}
	return p
}

func (f *Clients) Login(p *gorm.DB) string {
	var clients []Clients
	p.Table("clients").Where("name = ?", f.Name).Find(&clients)
	if len(clients) == 0 {
		l.ChLog <- l.Log{Message: fmt.Sprintf("Login: Попытка зайти под не зарегестрированным пользователем \"%s\"", f.Name), Level: ""}
		return fmt.Sprintf("Пользователь \"%s\" не зарегестрирован", f.Name)
	}
	if clients[0].Password == f.Password {
		l.ChLog <- l.Log{Message: fmt.Sprintf("Login: Пользователь \"%s\" - успешный вход", f.Name), Level: ""}
		return c.LogResp
	}
	l.ChLog <- l.Log{Message: fmt.Sprintf("Login: Пользователь \"%s\" - введен неверный пароль", f.Name), Level: ""}
	return fmt.Sprintf("Неверный пароль для пользователя \"%s\"", f.Name)
}

func (f *Clients) Reg(p *gorm.DB) string {
	var clients []Clients
	p.Table("clients").Where("name = ?", f.Name).Find(&clients)
	if len(clients) != 0 {
		l.ChLog <- l.Log{Message: fmt.Sprintf("Login: Попытка зарегестриваться под существующим пользователем \"%s\"", f.Name), Level: ""}
		return fmt.Sprintf("Пользователь \"%s\" уже зарегестрирован", f.Name)
	}
	p.Table("clients").Select("Name", "Password").Create(f)
	l.ChLog <- l.Log{Message: fmt.Sprintf("Login: Пользователь \"%s\" - зарегестрирован", f.Name), Level: ""}
	return c.RegResp
}

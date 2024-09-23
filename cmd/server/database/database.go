package database

import (
	"fmt"
	"log"
	"os"

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
	if err != nil {
		log.Fatalf("ConnectToDB: error to connect %v", err)
	}
	db.AutoMigrate(&Clients{})
	return db
}

func (p *PgConfig) GetConf() *PgConfig {
	conf, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("GetConf: file not found: %v", err)
	}
	err = yaml.Unmarshal(conf, p)
	if err != nil {
		log.Fatalf("GetConf: error unmarshalling yaml file: %v", err)
	}
	return p
}

func (f *Clients) Login(p *gorm.DB) string {
	var clients []Clients
	p.Table("clients").Where("name = ?", f.Name).Find(&clients)
	if len(clients) == 0 {
		log.Printf("Login: Попытка зайти под незарегестрированным пользователем \"%s\"\n", f.Name)
		return fmt.Sprintf("Пользователь \"%s\" не зарегестрирован", f.Name)
	}
	if clients[0].Password == f.Password {
		log.Printf("Login: Пользователь \"%s\" - успешный вход\n", f.Name)
		return c.LogResp
	}
	log.Printf("Login: Пользователь \"%s\" - введен неверный пароль\n", f.Name)
	return fmt.Sprintf("Неверный пароль для пользвателя \"%s\"", f.Name)
}

func (f *Clients) Reg(p *gorm.DB) string {
	var clients []Clients
	p.Table("clients").Where("name = ?", f.Name).Find(&clients)
	if len(clients) != 0 {
		log.Printf("Login: Попытка зарегестриваться под существующим пользователем \"%s\"\n", f.Name)
		return fmt.Sprintf("Пользователь \"%s\" уже зарегестрирован", f.Name)
	}
	p.Table("clients").Select("Name", "Password").Create(f)
	log.Printf("Login: Пользователь \"%s\" - зарегетрирован\n", f.Name)
	return c.RegResp
}

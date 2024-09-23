package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const congifPath = "./config/server_config.yaml"
const LogResp = "Добро пожаловать в fileStorage"
const RegResp = "Вы успешно зарегестрированы под именем "

type HTTPServer struct {
	Address     string `yaml:"address" env_default:"localhost:3333"`
	MaxByteSend int    `yaml:"maxByteSend" env_default:"100"`
}

func ConfigLoad() HTTPServer {
	conf, err := os.ReadFile(congifPath)
	if err != nil {
		log.Fatalf("ConfigLoad: file not found: %v", err)
	}
	var s HTTPServer
	err = yaml.Unmarshal(conf, &s)
	if err != nil {
		log.Fatalf("ConfigLoad: error unmarshalling yaml file: %v", err)
	}
	return s
}

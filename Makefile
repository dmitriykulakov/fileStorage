.PHONY: run_server run_client

DIRCLIENT=./cmd/client
DIRSERVER=./cmd/server

all: run_server

run_server: 
	sudo docker-compose -f ./docker-compose.yaml --env-file=config.env up --build -d
	go run $(DIRSERVER)/main.go

run_client: 
	go run $(DIRCLIENT)/main.go
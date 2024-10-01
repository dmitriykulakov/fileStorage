.PHONY: run_server run_client

DIRCLIENT=./cmd/client
DIRSERVER=./cmd/server

all: run_server

run_server: 
	sudo docker-compose up --build -d
	go run $(DIRSERVER)/main.go

run_client: 
	go run $(DIRCLIENT)/main.go
	

	
	
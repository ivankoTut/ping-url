.PHONY: run
run:
	@CONFIG_PATH='./config/local.yaml' go run cmd/ping/main.go

up:
	@docker-compose up -d

stop:
	@docker-compose down

status:
	@docker-compose ps

build:
	go build -o ping-url cmd/ping/main.go

start: up run
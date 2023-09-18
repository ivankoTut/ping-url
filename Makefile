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

args = `arg="$(filter-out $@,$(MAKECMDGOALS))" && echo $${arg:-${1}}`
migpq:
	@migrate create -ext sql -dir migrations/postgres $(args)

start: up run
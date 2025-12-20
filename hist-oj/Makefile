.PHONY: build run test clean docker-build docker-push docker-run docker-stop docker-logs deps fmt lint

# 变量
APP_NAME := hist-oj
BINARY := bin/$(APP_NAME)-service
DOCKER_IMAGE := $(APP_NAME):latest

# 本地开发
build:
	go build -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server/main.go

test:
	go test ./...

clean:
	rm -rf bin/
	go clean

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run

# Docker 操作
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-push:
	docker push $(DOCKER_IMAGE)

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart

# 一键部署（构建并启动）
deploy: docker-build docker-run

# 查看状态
status:
	docker-compose ps

# Rating 管理命令
reset-rating:
	go run ./cmd/reset_rating/main.go

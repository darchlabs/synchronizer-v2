# load .env file
include .env
export $(shell sed 's/=.*//' .env)

BIN_FOLDER_PATH=bin
SERVICE_NAME=synchronizer-v2
DOCKER_USER=darchlabs

rm:
	@echo "[rm] Removing..."
	@rm -rf bin

compile: rm
	@echo "[compile] Compiling..."
	@go build -o $(BIN_FOLDER_PATH)/synchronizer cmd/synchronizer/main.go

linux: rm
	@echo "[compile-linux] Compiling..."
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_FOLDER_PATH)/synchronizer-linux cmd/synchronizer/main.go

dev:
	@echo "[dev] Running..."
	@go run cmd/synchronizer/main.go

compose-dev:
	@echo "[compose-dev]: Running docker compose dev mode..."
	@docker-compose -f docker-compose.yml up --build

compose-stop:
	@echo "[compose-dev]: Running docker compose dev mode..."
	@docker-compose -f docker-compose.yml down

docker-login:
	@echo "[docker] Login to docker..."
	@docker login -u $(DOCKER_USER) -p $(DOCKER_PASS)

docker: docker-login
	@echo "[docker] pushing $(REGISTRY_URL)/$(SERVICE_NAME):$(VERSION)"
	@docker buildx build --platform linux/amd64,linux/arm64 --push -t $(DOCKER_USER)/$(SERVICE_NAME):$(VERSION) .

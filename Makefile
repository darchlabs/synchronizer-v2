# load .env file
include .env
export $(shell sed 's/=.*//' .env)

rm:
	@echo "[rm] Removing..."
	@rm -rf bin

compile: rm
	@echo "[compile] Compiling..."

run:
	@echo "[run] Running..."
	@NODE_URL=$(NODE_URL) INTERVAL_SECONDS=$(INTERVAL_SECONDS) DATABASE_URL=$(DATABASE_URL) go run cmd/synchronizer/main.go
	
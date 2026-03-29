APP_NAME := gophermart
CMD_DIR := .

.PHONY: all build run test lint tidy clean sqlc accrual

sqlc:
	sqlc generate -f sqlc/sqlc.yaml
build:
	go build $(CMD_DIR)/cmd/gophermart/main.go
run:
	go run $(CMD_DIR)/cmd/gophermart/main.go
accrual:
	./cmd/accrual/accrual_windows_amd64.exe
lint:
	go vet ./...
tidy:
	go mod tidy
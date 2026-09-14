test:
	go test $$(go list ./... | grep -v /docs) -v

test-coverage:
	go test $$(go list ./... | grep -v /docs) -coverprofile=coverage.out
	go tool cover -html=coverage.out

run:
	cd cmd/api && swag init -g main.go -o ../../docs --parseDependency --parseInternal --dir .,../../internal/adapter/delivery/http/handlers
	go run ./cmd/api
run-without-swagger:
	go run ./cmd/api

APP_NAME=pormade-email-manager
BIN_DIR=bin

.PHONY: all build build-linux clean

all: build

# Build padrão (usa o sistema atual)
build:
	go build -o $(BIN_DIR)/$(APP_NAME) .

# Build para Ubuntu Server (Linux AMD64)
# NOTA: O pacote confluent-kafka-go requer CGO (CGO_ENABLED=1). 
# Fazer cross-compiling de CGO do Windows para o Linux não é nativo do Go e requer um compilador C cruzado (como GCC para linux).
build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 .

# Forma recomendada para compilar para Linux estando no Windows usando Docker
build-linux-docker:
	docker run --rm -v "$$PWD":/usr/src/app -w /usr/src/app golang:1.21 \
		/bin/bash -c "apt-get update && apt-get install -y gcc && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 ."

clean:
	rm -rf $(BIN_DIR)

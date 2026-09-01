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
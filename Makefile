proto:
	protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/inventory/v1/inventory.proto

run-inventory:
	go run services/inventory-service/cmd/api/main.go

run-oms:
	go run services/oms-service/cmd/api/main.go

tidy:
	go mod tidy

test-inventory:
	go test ./services/inventory-service/internal/usecase -v

test-oms:
	go test ./services/oms-service/internal/usecase -v

run-ai:
	go run services/ai-agent-service/cmd/api/main.go
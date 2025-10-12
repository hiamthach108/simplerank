test:
	go test ./... -v

run:
	go run cmd/main.go

.PHONY: test run
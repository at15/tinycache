install: fmt
	go install ./cmd/tinycache

fmt:
	go fmt ./...

test:
	go test ./...

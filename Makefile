install: fmt
	go install ./cmd/tinycache

fmt:
	go fmt ./...

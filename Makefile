IMAGE_VERSION := 0.1

.PHONY: proto

install: fmt
	go install ./cmd/tinycache

fmt:
	go fmt ./...

test:
	go test ./...

docker-build:
	docker build -t tinycache:$(IMAGE_VERSION) .

docker-run:
	docker run -p 8080:8080 tinycache:$(IMAGE_VERSION)

# go install golang.org/x/tools/cmd/godoc@latest
doc:
	@echo "Open http://localhost:6060/pkg/github.com/at15/tinycache/cache/"
	godoc -http=:6060

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/tinycache.proto

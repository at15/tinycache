# tinycache

A in memory KV cache in Go with http API. Supports LRU and TTL.
Toy project, do NOT use in production.

## Usage

See [cmd/tinycache/main.go](cmd/tinycache) for using the server, which uses the library in [server/http.go](server/http.go)

### Server

```bash
# Install to $GOPATH/bin
make install
# Run server on localhost:8080
# View prometheus metrics on http://localhost:8080/stats
tinycache
```

### Client

curl

```bash
# set
curl -X PUT http://localhost:8080/cache/b1/k1 -d "v1"
# set with ttl and policy
curl -X PUT "http://localhost:8080/cache/b1/k1?ttl=1s&policy=lru" -d "v1"

# get
curl -X GET http://localhost:8080/cache/b1/k1

# delete
curl -X DELETE http://localhost:8080/cache/b1/k1
```

### Docker

```bash
make docker-build
make docker-run
```

## Development

### gRPC/Protobuf

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
make proto
```

## TODO

KV

- [x] copy the interface
- [x] in memory cache
  - [x] bucket, max 255 keys is per bucket or entire cache? Should be entire cache otherwise there is no limit on number of buckets.
  - [x] eviction policy, each operation can have different policy in options???
  - [x] ttl (lazy or run in background)
  - [x] test

Server

- [ ] grpc
- [x] http
- [x] metrics, using prometheus
- [ ] client in the cli
- [ ] redis protocol? (if I have time)

## References

- https://github.com/hashicorp/golang-lru has more complex LRU implementation and supports resize.
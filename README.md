# tinycache

A in memory KV cache in Go with http API. Supports LRU and TTL.
Toy project, do NOT use in production.

## Usage

See [cmd/tinycache/main.go](cmd/tinycache) for using the server and client, which uses the library in [server/http.go](server/http.go) and [server/grpc.go](server/grpc.go).

### Server

```bash
# Install to $GOPATH/bin
make install
# Run server on localhost:8080
# View prometheus metrics on http://localhost:8080/stats
tinycache server
# gRPC server
tinycache server --grpc
```

### Client

#### curl

NOTE: Only works for HTTP server.

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

#### REPL

NOTE: Only works for gRPC right now.

```text
tinycache client
TinyCache CLI (type 'help' for commands, 'exit' to quit)
Connected to localhost:8080
> set k1 v1
Usage: set <bucket> <key> <value> [ttl_ms]
> set b1 k1 v1
OK
> set b2 k2 v3
OK
> get b1 k1
v1
> get b2 k3
Error: rpc error: code = Unknown desc = key k3 not found
> del b1 k1
OK
> get b1 k1
Error: rpc error: code = Unknown desc = bucket b1 not found
> exit
```

### Docker

```bash
make docker-build
make docker-run
```

## Development

### gRPC/Protobuf

To regenerate pb files after chainging the proto

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
make proto
```

### How it works

Just using a linked list to track the insertion order and recent usage.
We are using a **single** linked list to track different policies, so
the behavior can be strange when mixed policies are used ... Ideally
the eviction policy should be same for entire cache and not specified
in each operation.

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
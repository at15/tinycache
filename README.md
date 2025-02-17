# tinycache

A in memory kv cache in Go with http and grpc interface. Toy project, do NOT use in production.

## Usage

```bash
# Install to $GOPATH/bin
make install
# Run server on localhost:8080
# View metrics on http://localhost:8080/stats
tinycache
```

curl

```bash
# set
curl -X PUT http://localhost:8080/cache/b1/k1 -d "v1"

# get
curl -X GET http://localhost:8080/cache/b1/k1

# delete
curl -X DELETE http://localhost:8080/cache/b1/k1
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
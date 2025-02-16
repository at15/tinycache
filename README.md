# tinycache

A in memory kv cache in Go with http and grpc interface. Toy project, do NOT use in production.

## Usage

```bash
make install
```

## TODO

KV

- [x] copy the interface
- [ ] in memory cache
  - [ ] bucket, max 255 keys is per bucket or entire cache? Should be entire cache otherwise there is no limit on number of buckets.
  - [ ] eviction policy, each operation can have different policy in options???
  - [ ] ttl (lazy or run in background)
  - [ ] test

Server

- [ ] grpc
- [ ] http
- [ ] metrics
- [ ] client in the cli
- [ ] redis protocol? (if I have time)
# pq-backend

Go backend for the PrivateQuery demo.

## Features

- Single codebase for both Server A and Server B
- SQLite sample dataset initialization at startup
- Catalog version, public catalog list, health check, and DPF evaluation endpoints
- C++ DPF core integration through `cgo`

## Run

Server A:

```bash
go run ./cmd/server
```

Server B:

```bash
PORT=8082 PARTY=1 DB_PATH=data/server_b.db go run ./cmd/server
```

## Build

```bash
go build ./...
```

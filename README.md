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

## Ubuntu 24.04 Release

The repository root contains a Windows build entrypoint for release packaging:

```powershell
pwsh ./scripts/build-release.ps1
```

The backend release output is placed in `deploy/release/backend/` and includes:

- `server-a`
- `server-b`
- `server-a.env`
- `server-b.env`
- `start-server-a.sh`
- `start-server-b.sh`

For the target server ports, use `18081` for Server A and `18082` for Server B.

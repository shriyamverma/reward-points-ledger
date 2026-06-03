# reward-points-ledger
Reward Points Ledger Service

## Project Directory Structure

```
reward-points-ledger/
├── cmd/
│   └── api/
│       └── main.go
├── docs/
│   └── swagger.yaml
├── internal/
│   ├── domain/
│   │   └── models.go
│   ├── handler/
│   │   └── http.go
│   ├── repository/
│   │   └── memory.go
│   └── service/
│       └── ledger.go
│       └── ledger_test.go
├── docker-compose.yaml
├── Dockerfile
├── go.mod
└── go.sum
```

## How to run locally

1. Start the server: `docker compose up --build`
2. Close the server: `docker compose down` 

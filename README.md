# Reward Points Ledger Service

A production-ready, highly resilient, and fully containerized fintech rewards engine built in `Go`. The system tracks member profiles, logs transactional point allocations, and performs running-sum balance calculations using clean architecture principles and a persistent PostgreSQL storage engine.

## ⚡ Quick-Start Shortcuts (Makefile)

The project includes a `Makefile` to abstract complex Docker and Go commands into simple shortcuts:

* Boot up the full environment: `make up`
* Teardown the containers: `make down`
* Wipe database volumes for a clean seed reset: `make clean`
* Run the complete mock unit test suite: `make test`
* View visual HTML test coverage reports: `make cover`
* Audit live database records instantly via console: `make db-audit`
* Open an interactive PostgreSQL CLI session: `make db-shell`
* Generate sequential schema files: `make migrate-create name=migration_name`
* Deploy pending database changes: `make migrate-up`
* Rollback the latest schema layer: `make migrate-down`

## 🛠 Architecture Overview

The system strictly adheres to **Clean Architecture / Domain-Driven Design (DDD)** concepts, ensuring that business logic remains decoupled from delivery mechanisms, databases, and third-party drivers:

* **Domain Layer** (`internal/domain`): Core models and enterprise business logic rules (completely independent).
* **Service Layer** (`internal/service`): Orchestrates application use cases and enforces domain boundaries.
* **Repository Layer** (`internal/repository`): Manages data state and low-level persistence via native drivers.
* **Handler Layer** (`internal/handler`): Handles HTTP routing, JSON marshaling, and middleware delivery mechanisms.

## 📁 Project Directory Structure

```text
reward-points-ledger/
├── .github/
│   └── workflows/
│       ├── ci.yaml                 # Continuous Integration (PR Gates)
│       └── cd.yaml                 # Continuous Delivery (Releases)
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point & orchestration
├── docs/
│   └── swagger.yaml               # Design-first OpenAPI specification
├── internal/
│   ├── domain/
│   │   └── models.go              # Pure domain entities & error interfaces
│   ├── handler/
│   │   ├── http.go                # REST HTTP Controllers
│   │   └── middleware.go          # Custom self-contained CORS middleware
│   ├── repository/
│   │   ├── migrations/                # Schema tracking directory
│   │   │   ├── 000001_init.up.sql     # Baseline tables setup schema
│   │   │   └── 000001_init.down.sql   # Reversal schema pattern
│   │   ├── memory.go              # Legacy/Testing Ephemeral Storage
│   │   ├── postgres.go            # Production PGX persistence engine
│   │   ├── postgres_init.go       # High-resilience DB connection retry lifecycle
│   │   └── postgres_test.go       # Unified pgxmock repository unit tests
│   └── service/
│       ├── ledger.go              # Core ledger business rules service
│       └── ledger_test.go         # Domain service behavior verification tests
├── docker-compose.yaml            # Local Infrastructure Cluster Definition
├── Dockerfile                     # Multi-stage scratch production compiler blueprint
├── go.mod                         # Go Module Specifications (Target: 1.25.11)
├── go.sum                         # Driver Checksums
└── Makefile                       # Developer Workflow Control Center
```

## 🚀 Key Technical Features

* **Resilient Database Bootstrapping**: Implements an isolated, 10-attempt retry loop with explicit `context.WithTimeout` logic to guarantee clean connection pool recovery if the database cluster is lagging during hot-reloads.
* **Unified Query Architecture (`pgx.NamedArgs`)**: Standardizes parameter maps across all lookup pathways, preventing positional argument errors while solving implicit database type deduction constraints.
* **Proactive Database Clock Sync**: Delegates chronological timestamp tracking entirely to the database engine via native `NOW()` structures, preventing time-drift anomalies across concurrent application instances.
* **Gapless Sequence Allocation**: Utilizes database-level Common Table Expressions (CTEs) paired with `WHERE NOT EXISTS` clauses during member registration. This rejects duplicate emails *before* incrementing sequence counters, preserving uninterrupted ID metrics for strict financial audits.
* **Instant Documentation Previews**: Integrates a decoupled local Swagger UI container volumed directly to your local file modifications for zero-downtime documentation editing.

## 💻 How to Run Locally

### Prerequisites

Make sure you have [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed on your machine.

### 1. Build and Launch the Full Stack

Spin up the Go API backend, the PostgreSQL 16 database instance, and the Swagger UI documentation layer simultaneously:

```bash
docker compose up --build

```

* **API Service**: Runs on `http://localhost:8080`
* **Swagger API Documentation**: Accessible at `http://localhost:8081`

### 2. Teardown System Resources

To stop all active services and unmount temporary containers cleanly:

```bash
docker compose down

```

## 🧪 Verification & Automated Testing

### Run All Isolated Unit Tests

Execute the comprehensive mock verification suite covering both the service layer and the unified repository parameters instantly without needing an active database connection:

```bash
go test ./... -v

```

### Manual Database Audit Tracking

To peek under the hood and confirm data schema adjustments or sequence boundaries, query the live database container tracking engine via `psql`:

```bash
docker exec -it rewards_ledger_db psql -U root -d rewards_db -c "SELECT * FROM members ORDER BY member_id ASC;"

```

```text
 member_id |     name      |       email       |          created_at           
-----------+---------------+-------------------+-------------------------------
         1 | Alice Johnson | alice@example.com | 2026-06-04 07:45:12.123456+00

```

```bash
docker exec -it rewards_ledger_db psql -U root -d rewards_db -c "SELECT * FROM rewards;"

```

```text
 reward_id | member_id | point_type_id | points |   description   |          event_date           
-----------+-----------+---------------+--------+-----------------+-------------------------------
         1 |         1 |             1 |    100 | Welcome Bonus   | 2026-06-04 07:45:15.654321+00

```

### 📜 Migration Strategy

Every database structural adjustment requires two corresponding files inside the `internal/repository/migrations` directory:
1. `XXXXXX_name.up.sql`: Applies the delta schema upgrades (e.g., creating tables, adding columns).
2. `XXXXXX_name.down.sql`: Implements a strict rollback strategy to reverse the exact changes made by the companion `up` script.

The schema lifecycle is automatically tracked inside a dedicated metadata table (`schema_migrations`) managed directly inside our live PostgreSQL cluster.

### 🛠️ Migration Workflows

The centralized `Makefile` streamlines migration operations into simple, highly intuitive shortcut targets:

#### 1. Generate New Migration Scaffolding
To create a new sequential up/down pair of empty SQL files, pass the descriptive snake-case `name` variable directly into the engine:
```bash
make migrate-create name=add_member_tier_status
```

*This will instantly generate `000002_add_member_tier_status.up.sql` and `000002_add_member_tier_status.down.sql` within our local source tree.*

#### 2. Apply Pending Migrations

To upgrade your running PostgreSQL schema to the latest structural configuration, execute:

```bash
make migrate-up
```

#### 3. Rollback the Last Applied Schema State

If we need to step backward by a single migration execution block during active feature prototyping or hotfixes, run:

```bash
make migrate-down
```

### 🔍 Operational Status Audit

To audit exactly which migration version step your database engine is currently sitting on, you can execute a direct terminal readout against the inner state metadata:

```bash
docker exec -it rewards_ledger_db psql -U root -d rewards_db -c "SELECT * FROM schema_migrations;"
```

Our system console will display the operational index mapping matrix:

```text
 version | dirty 
---------+-------
       2 | f
(1 row)
```

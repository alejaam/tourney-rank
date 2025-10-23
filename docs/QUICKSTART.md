# TourneyRank Quick Start Guide

Get TourneyRank up and running in 5 minutes!

## Prerequisites

Make sure you have the following installed:

- **Go** 1.21 or higher: [Install Go](https://golang.org/doc/install)
- **Docker** & **Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **Make** (optional but recommended)
- **golang-migrate** (for database migrations): 
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

## Step 1: Clone the Repository

```bash
git clone https://github.com/yourusername/tourney-rank.git
cd tourney-rank
```

## Step 2: Setup Environment

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` if needed (default values work for local development).

## Step 3: Start Infrastructure

Start PostgreSQL and Redis using Docker Compose:

```bash
make docker-up
```

Or manually:

```bash
docker-compose up -d
```

Verify containers are running:

```bash
docker-compose ps
```

You should see:
- `tourneyrank-db` (PostgreSQL) - port 5432
- `tourneyrank-redis` (Redis) - port 6379

## Step 4: Run Database Migrations

Apply database migrations to create tables:

```bash
make migrate-up
```

Or manually:

```bash
migrate -path migrations -database "postgresql://tourneyrank:tourneyrank@localhost:5432/tourneyrank?sslmode=disable" up
```

**Expected output**:
```
1/u create_games_table (123.456ms)
2/u create_users_and_players (234.567ms)
3/u create_tournaments_and_matches (345.678ms)
```

## Step 5: Download Dependencies

```bash
go mod download
```

## Step 6: Run the Application

```bash
make run
```

Or manually:

```bash
go run cmd/service/main.go
```

**Expected output**:
```
TourneyRank starting...
HTTP Server would start on port 8080
WebSocket Server would start on port 8081
```

## Step 7: Verify Installation

### Check Database

Connect to PostgreSQL and verify the default game was created:

```bash
docker exec -it tourneyrank-db psql -U tourneyrank -d tourneyrank
```

Run query:

```sql
SELECT id, name, slug FROM games;
```

Expected result:
```
                  id                  |          name          |  slug   
--------------------------------------+------------------------+---------
 <uuid>                               | Call of Duty: Warzone  | warzone
```

Exit psql:
```sql
\q
```

### Run Tests

```bash
make test
```

Or:

```bash
go test ./... -v
```

All tests should pass! ✅

## What's Next?

### Explore the Codebase

- **Domain Layer**: `internal/domain/` - Business entities and logic
- **Migrations**: `migrations/` - Database schema
- **Main Entry**: `cmd/service/main.go` - Application bootstrap

### Read Documentation

- **Architecture**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System design and patterns
- **Roadmap**: [docs/TODO.md](docs/TODO.md) - Development plan and tasks
- **README**: [README.md](README.md) - Full project documentation

### Development Workflow

```bash
# Run application
make run

# Run tests with coverage
make test-race
make coverage

# Run linter (requires golangci-lint)
make lint

# Format code
make fmt

# Create new migration
make migrate-create NAME=add_new_feature

# Stop infrastructure
make docker-down
```

### Next Development Steps

Based on the roadmap in `docs/TODO.md`, the next priorities are:

1. **Configuration Layer** - Load environment variables
2. **Database Repositories** - Implement data access layer
3. **HTTP Server** - Create REST API endpoints
4. **Authentication** - User login and JWT tokens

See detailed tasks in [docs/TODO.md](docs/TODO.md).

## Common Issues

### Port Already in Use

If PostgreSQL or Redis ports are already in use:

```bash
# Stop existing containers
docker-compose down

# Check what's using the port
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis

# Kill the process or change ports in docker-compose.yml
```

### Migration Errors

If migrations fail:

```bash
# Rollback migrations
make migrate-down

# Verify database is empty
docker exec -it tourneyrank-db psql -U tourneyrank -d tourneyrank -c "\dt"

# Try again
make migrate-up
```

### Module Errors

If you see "no required module provides package":

```bash
go mod tidy
go mod download
```

## Project Structure Overview

```
tourney-rank/
├── cmd/service/          # Application entry point
├── internal/
│   ├── domain/           # Business logic (entities, services)
│   ├── app/              # Use cases (to be implemented)
│   ├── infra/            # Infrastructure (to be implemented)
│   └── config/           # Configuration (to be implemented)
├── migrations/           # Database migrations
├── docs/                 # Documentation
├── web/                  # Frontend (to be implemented)
├── Makefile              # Development commands
├── docker-compose.yml    # Infrastructure setup
└── go.mod                # Go dependencies
```

## Contributing

We welcome contributions! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Submit a pull request

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/yourusername/tourney-rank/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/tourney-rank/discussions)
- **Documentation**: [docs/](docs/)

## Clean Up

When you're done:

```bash
# Stop and remove containers
make docker-down

# Remove volumes (WARNING: deletes all data!)
make docker-clean

# Remove built binary
make clean
```

---

**Congratulations!** 🎉 You have TourneyRank running locally. Happy coding!

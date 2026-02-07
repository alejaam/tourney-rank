# TourneyRank Architecture

TourneyRank follows **Clean Architecture** with the dependency rule:

`infra → usecase → domain` (never the other way around)

## Repository Layout

```
cmd/service/                # Application entry point
internal/
  domain/                   # Entities + domain interfaces (no external deps)
  usecase/                  # Business logic (services + DTOs)
  infra/
    http/                   # Handlers, routing, middleware
    mongodb/                # MongoDB repository implementations
  config/                   # Env config + validation
frontend/                   # React + Vite app
```

## Layers

### Domain (`internal/domain`)

- Entities and domain rules (e.g. games, players, ranking)
- Repository interfaces live here (ports)
- Must not depend on `infra` or frameworks

### Usecase (`internal/usecase`)

- Orchestrates domain logic to deliver business capabilities
- Depends only on `domain`
- Defines request/response DTOs used by handlers

### Infra (`internal/infra`)

- HTTP transport: routing, JSON encode/decode, auth middleware
- Persistence: MongoDB repositories that implement domain interfaces
- Depends on `usecase` and `domain`

## Key Patterns

### Strategy Pattern (Ranking Calculators)

Ranking uses calculators in `internal/domain/ranking` (e.g. `WarzoneCalculator`, `DefaultCalculator`).

### Repository Pattern

## Testing

- Unit tests live alongside packages (e.g. `*_test.go` under `internal/domain/...`).
- Prefer table-driven tests; keep domain tests fast and isolated.

## Extending the system

- Add a new domain concept under `internal/domain/<feature>` (constructors validate inputs).
- Add a usecase service under `internal/usecase/<feature>`.
- Add HTTP handlers under `internal/infra/http/handlers` and wire routes.
    r.games[game.Slug] = game

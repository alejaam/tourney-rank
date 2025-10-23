# TourneyRank Architecture

## Overview

TourneyRank follows **Clean Architecture** principles with clear separation of concerns across layers. The architecture is designed to be:

- **Testable**: Domain logic is independent of external frameworks
- **Maintainable**: Clear boundaries between layers
- **Scalable**: Easy to add new games, features, and integrations
- **Flexible**: Support for multiple games without structural changes

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    Presentation Layer                    │
│                   (HTTP, WebSocket)                      │
├─────────────────────────────────────────────────────────┤
│                   Application Layer                      │
│              (Use Cases / Business Logic)                │
├─────────────────────────────────────────────────────────┤
│                     Domain Layer                         │
│        (Entities, Value Objects, Domain Services)        │
├─────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                    │
│          (Database, Cache, External Services)            │
└─────────────────────────────────────────────────────────┘
```

### 1. Domain Layer (`internal/domain/`)

**Pure business logic with zero external dependencies.**

- **Entities**: Core business objects (Game, Player, Tournament, Team, Match)
- **Value Objects**: Immutable objects (Tier, RankingWeights, StatSchema)
- **Domain Services**: Business logic that doesn't belong to a single entity (Ranking calculation)
- **Interfaces**: Repository contracts defined here, implemented in infrastructure

**Key Characteristics**:
- No imports from other layers
- Only standard library and domain-specific types
- Rich domain models with behavior
- Business rules enforcement

**Example**:
```go
// internal/domain/game/game.go
type Game struct {
    ID             uuid.UUID
    Name           string
    StatSchema     StatSchema
    RankingWeights RankingWeights
}

func (g *Game) UpdateWeights(weights RankingWeights) error {
    // Business rule: weights must sum to 1.0
    if err := validateRankingWeights(weights); err != nil {
        return err
    }
    // ...
}
```

### 2. Application Layer (`internal/app/`)

**Orchestrates domain entities to fulfill use cases.**

- **Use Cases**: Application-specific business rules
- **Service Implementations**: Coordinates repositories and domain services
- **DTOs**: Data Transfer Objects for external communication
- **Ports**: Input/output interfaces

**Key Characteristics**:
- Depends only on Domain layer
- Defines repository interfaces (ports)
- No HTTP, database, or framework code
- Transaction boundaries

**Example**:
```go
// internal/app/stats/service.go
type Service struct {
    playerRepo     player.Repository
    rankingService *ranking.Service
    cache          Cache
}

func (s *Service) RecordMatchStats(ctx context.Context, input MatchStatsInput) error {
    // 1. Validate input
    // 2. Update player stats (domain entity)
    // 3. Recalculate ranking (domain service)
    // 4. Persist changes (repository)
    // 5. Update cache
    // 6. Publish events
}
```

### 3. Infrastructure Layer (`internal/infra/`)

**Concrete implementations of external concerns.**

- **HTTP**: REST API handlers, request/response models
- **WebSocket**: Real-time communication handlers
- **PostgreSQL**: Repository implementations using database
- **Redis**: Caching implementations
- **Discord**: External service integrations

**Key Characteristics**:
- Implements interfaces defined in Domain/Application layers
- Framework-specific code lives here
- External service clients
- Database queries and migrations

**Example**:
```go
// internal/infra/postgres/player_repository.go
type PlayerRepository struct {
    db *sql.DB
}

func (r *PlayerRepository) GetPlayerStats(ctx context.Context, playerID, gameID uuid.UUID) (*player.PlayerStats, error) {
    // SQL query implementation
    // Maps database rows to domain entities
}
```

### 4. Presentation Layer (`internal/infra/http/`, `internal/infra/websocket/`)

**Handles external communication protocols.**

- **REST API**: HTTP handlers, routing, request validation
- **WebSocket**: Real-time updates, pub/sub
- **Middleware**: Authentication, logging, CORS
- **Serialization**: JSON encoding/decoding

## Design Patterns

### Strategy Pattern (Ranking Calculators)

Different games use different ranking algorithms:

```go
type Calculator interface {
    Calculate(ctx context.Context, stats *PlayerStats, game *Game) (float64, error)
    SupportsGame(gameSlug string) bool
}

// Warzone-specific calculator
type WarzoneCalculator struct{}

// Generic fallback calculator
type DefaultCalculator struct{}

// Service selects appropriate strategy
type RankingService struct {
    calculators []Calculator
}
```

**Benefits**:
- Easy to add new games
- Each game has custom ranking logic
- Swappable algorithms

### Repository Pattern

Abstracts data persistence:

```go
// Domain defines the contract
type Repository interface {
    GetPlayerStats(ctx context.Context, playerID, gameID uuid.UUID) (*PlayerStats, error)
    UpdatePlayerStats(ctx context.Context, stats *PlayerStats) error
}

// Infrastructure provides implementation
type PostgreSQLPlayerRepository struct {
    db *sql.DB
}
```

**Benefits**:
- Domain code independent of database
- Easy to mock for testing
- Can swap database implementations

### Event-Driven Architecture

Domain events for decoupling:

```go
type MatchStatsRecorded struct {
    MatchID    uuid.UUID
    PlayerID   uuid.UUID
    Stats      map[string]interface{}
    RecordedAt time.Time
}

// Event handler updates rankings asynchronously
func (h *RankingUpdateHandler) Handle(ctx context.Context, event MatchStatsRecorded) error {
    // Recalculate rankings
    // Update cache
    // Broadcast via WebSocket
}
```

**Benefits**:
- Loose coupling between components
- Easy to add new side effects
- Async processing

## Data Flow

### Example: Recording Match Statistics

```
1. HTTP Request (JSON)
   ↓
2. HTTP Handler (Validation)
   ↓
3. Application Service (RecordMatchStats use case)
   ↓
4. Domain Entity (Player.UpdateStats)
   ↓
5. Domain Service (RankingService.Calculate)
   ↓
6. Repository (Persist to PostgreSQL)
   ↓
7. Cache Update (Redis)
   ↓
8. Event Emission (MatchStatsRecorded)
   ↓
9. WebSocket Broadcast (Leaderboard update)
   ↓
10. HTTP Response (Success)
```

## Multi-Game Support Architecture

### Polymorphic Data Schema

Using PostgreSQL JSONB for flexibility:

```sql
-- Games define their own stat schemas
CREATE TABLE games (
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    stat_schema JSONB,  -- {"kills": {...}, "damage": {...}}
    ranking_weights JSONB  -- {"kd_ratio": 0.4, "avg_kills": 0.3}
);

-- Player stats are flexible per game
CREATE TABLE player_stats (
    id UUID PRIMARY KEY,
    player_id UUID,
    game_id UUID,
    stats JSONB  -- Adapts to game's stat_schema
);
```

### Game Registry Pattern

```go
type GameRegistry struct {
    games map[string]*Game
}

func (r *GameRegistry) Register(game *Game) error {
    r.games[game.Slug] = game
}

func (r *GameRegistry) ValidateStats(gameSlug string, stats map[string]interface{}) error {
    game := r.games[gameSlug]
    return game.ValidateStat(stats)
}
```

## Testing Strategy

### Unit Tests (Domain Layer)

```go
func TestGame_UpdateWeights(t *testing.T) {
    game := NewGame("Warzone", "warzone", ...)
    
    err := game.UpdateWeights(RankingWeights{"kd": 0.5})
    
    assert.Error(t, err) // Weights don't sum to 1.0
}
```

### Integration Tests (Application Layer)

```go
func TestStatsService_RecordMatchStats(t *testing.T) {
    // Use real PostgreSQL via testcontainers
    db := setupTestDB(t)
    defer db.Close()
    
    service := NewStatsService(NewPlayerRepo(db), ...)
    
    err := service.RecordMatchStats(ctx, validInput)
    
    assert.NoError(t, err)
    // Verify database state
}
```

### Table-Driven Tests

```go
tests := []struct {
    name          string
    input         MatchStatsInput
    expectedError error
}{
    {"valid warzone stats", validWarzoneInput, nil},
    {"invalid stats", invalidInput, ErrInvalidStats},
}

for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        // Test logic
    })
}
```

## Database Schema Design

### Flexibility Through JSONB

- **games.stat_schema**: Defines available stats for a game
- **games.ranking_weights**: Configurable ranking formula
- **player_stats.stats**: Aggregated stats (flexible structure)
- **match_stats.stats**: Per-match player stats

### Example: Adding a New Game

```sql
INSERT INTO games (name, slug, stat_schema, ranking_weights) VALUES (
    'Fortnite',
    'fortnite',
    '{
        "eliminations": {"type": "integer", "label": "Eliminations"},
        "placement": {"type": "integer", "label": "Placement"},
        "builds": {"type": "integer", "label": "Builds"}
    }',
    '{
        "eliminations": 0.35,
        "placement": 0.35,
        "builds": 0.30
    }'
);
```

No code changes needed! The Strategy Pattern handles the rest.

## Caching Strategy

### Redis Cache Layers

1. **Leaderboards**: Sorted sets for fast ranking queries
2. **Player Stats**: Hash for quick lookups
3. **Active Tournaments**: Set with TTL

```go
// Cache key patterns
leaderboard:{tournamentID}:{gameID}
player:{playerID}:stats:{gameID}
tournament:{tournamentID}:active
```

### Cache Invalidation

- On `MatchStatsRecorded` event → invalidate affected leaderboards
- On `TournamentEnded` event → remove from active set
- TTL for temporary data

## WebSocket Architecture

### Real-Time Updates

```
Client                    Server
  |                          |
  |--- Connect to /ws ------>|
  |<-- Connection OK --------|
  |                          |
  |-- Subscribe topic:       |
  |   leaderboard:tournament |
  |                          |
  |   [Match stats recorded] |
  |   [Ranking recalculated] |
  |                          |
  |<-- Leaderboard Update ---|
  |   {players: [...]}       |
```

### Topics

- `leaderboard:{tournamentID}`: Tournament leaderboard updates
- `match:{matchID}`: Match events
- `player:{playerID}`: Personal notifications

## Observability

### Structured Logging

```go
logger.Info("match stats recorded",
    "match_id", matchID,
    "player_id", playerID,
    "game", gameSlug,
    "duration_ms", elapsed.Milliseconds(),
)
```

### Metrics

- `matches_recorded_total{game="warzone"}`
- `ranking_calculation_duration_seconds{game="warzone"}`
- `cache_hit_ratio{key_pattern="leaderboard:*"}`

### Tracing

- OpenTelemetry for distributed tracing
- Trace stats submission → ranking calculation → cache update

## Future Extensibility

### Ready for:

1. **New Games**: Add to database, implement Calculator if needed
2. **New Tournament Modes**: Extend Tournament entity, add validation
3. **Stream Overlays**: WebSocket already provides real-time data
4. **Mobile App**: REST API is platform-agnostic
5. **ML Predictions**: Domain events feed training data
6. **Microservices**: Clean boundaries enable easy extraction

## References

- [Clean Architecture - Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design - Eric Evans](https://www.domainlanguage.com/ddd/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

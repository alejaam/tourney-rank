# Git Commit Message Template

## Initial Setup Commit

```
feat: initialize TourneyRank multi-game tournament platform

BREAKING CHANGE: Initial project setup with Clean Architecture

This commit establishes the foundational structure for TourneyRank,
a multi-game tournament management platform with automated ranking.

### What's Included:

**Architecture & Structure**
- Clean Architecture with DDD (domain → app → infra)
- Go module initialized with core dependencies
- Docker Compose setup (PostgreSQL 15 + Redis 7)
- Makefile with 15+ development commands
- Dockerfile for production deployment

**Domain Layer (Pure Business Logic)**
- Game entity with flexible JSONB schema support
- Player entity with multi-platform ID management
- PlayerStats with per-game statistics tracking
- Ranking service with Strategy Pattern
- WarzoneCalculator implementing game-specific ranking algorithm
- DefaultCalculator as fallback for generic games

**Database Schema (3 Migrations)**
- 000001: games table with stat_schema and ranking_weights (JSONB)
- 000002: users, players, player_stats tables
- 000003: tournaments, teams, matches, match_stats, registrations

**Testing**
- Unit tests for Game entity (100% coverage)
- Unit tests for ranking weight validation
- Table-driven tests following Go best practices
- All tests passing (4 test suites)

**Documentation**
- README.md: Comprehensive project overview
- docs/ARCHITECTURE.md: Detailed architecture guide with patterns
- docs/TODO.md: 8-week development roadmap
- docs/QUICKSTART.md: Getting started in 5 minutes
- docs/PROJECT_SUMMARY.md: Technical executive summary
- docs/RESUMEN_EJECUTIVO.md: Executive summary (Spanish)

**Development Tools**
- .gitignore for Go, Node, Docker artifacts
- .env.example with configuration template
- LICENSE (MIT)
- Comprehensive Makefile

### Key Features Ready:

✅ Multi-game architecture (start with Warzone, expand to any game)
✅ Flexible JSONB schemas (add games without code changes)
✅ Strategy Pattern for ranking calculators
✅ Automated ranking with configurable weights
✅ Player tier system (Elite/Advanced/Intermediate/Beginner)
✅ Docker infrastructure ready
✅ Clean separation of concerns (testable domain logic)

### Technical Decisions:

- JSONB for game-agnostic stat storage
- Strategy Pattern for extensible ranking algorithms
- Clean Architecture for maintainability
- Event-Driven design (prepared for future events)
- Repository Pattern for data access abstraction

### Database Pre-Seeded With:

- Call of Duty: Warzone game configuration
  - stat_schema: kills, deaths, damage, contracts, cash
  - ranking_weights: kd_ratio(0.4), avg_kills(0.3), avg_damage(0.2), consistency(0.1)

### Next Steps (Phase 1):

Week 1-2: Configuration layer, PostgreSQL repositories, Authentication, HTTP server

See docs/TODO.md for complete roadmap.

### Verification:

✅ `make build` - Compiles successfully
✅ `make test` - All tests passing
✅ `make docker-up` - Infrastructure starts
✅ `make migrate-up` - Migrations apply cleanly

### Project Stats:

- Go Files: 10
- Lines of Code: ~1,200
- Tests: 4 suites, 100% passing
- Migrations: 3 (with rollback)
- Docker Services: 2 (PostgreSQL, Redis)
- Documentation: 6 comprehensive files
```

---

## Future Commit Convention

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

### Types:
- `feat:` New feature
- `fix:` Bug fix
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `docs:` Documentation changes
- `chore:` Maintenance tasks
- `perf:` Performance improvements
- `ci:` CI/CD changes

### Scope (Optional):
- `feat(ranking):` Feature in ranking system
- `fix(auth):` Bug in authentication
- `refactor(domain):` Refactoring domain layer

### Examples:

```
feat(stats): add match statistics submission endpoint

Implements POST /api/matches/stats to receive player statistics
in JSON format. Validates stats against game schema and persists
to database.

Closes #12
```

```
feat(ranking): implement Fortnite ranking calculator

Adds FortniteCalculator with elimination-based scoring:
- Eliminations: 35%
- Placement: 35%
- Builds: 30%

Updates ranking service to support multiple calculators.
```

```
test(domain): add player stats calculation tests

Adds table-driven tests for:
- K/D ratio calculation
- Stat aggregation
- Tier determination by percentile
```

# Contributing to TourneyRank

Thank you for considering contributing to TourneyRank! This document provides guidelines and best practices for contributing to the project.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Workflow](#development-workflow)
4. [Coding Standards](#coding-standards)
5. [Testing Guidelines](#testing-guidelines)
6. [Commit Convention](#commit-convention)
7. [Pull Request Process](#pull-request-process)
8. [Architecture Guidelines](#architecture-guidelines)

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Maintain professional communication

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make
- Git

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/yourusername/tourney-rank.git
cd tourney-rank

# Setup infrastructure and dependencies
make setup

# Verify setup
make test
```

## Development Workflow

### 1. Create a Branch

```bash
# Feature branch
git checkout -b feat/add-fortnite-calculator

# Bug fix branch
git checkout -b fix/ranking-calculation

# Refactoring branch
git checkout -b refactor/optimize-queries
```

### 2. Make Changes

Follow the [coding standards](#coding-standards) and [architecture guidelines](#architecture-guidelines).

### 3. Run Tests

```bash
# Run all tests
make test

# Run tests with race detector
make test-race

# Generate coverage report
make coverage
```

### 4. Format and Lint

```bash
# Format code
make fmt

# Run linter
make lint

# Vet code
make vet
```

### 5. Commit Changes

Follow the [commit convention](#commit-convention).

### 6. Push and Create PR

```bash
git push origin feat/add-fortnite-calculator
```

Then create a Pull Request on GitHub.

## Coding Standards

### Go Code Style

1. **Follow Go idioms**
   - Use `gofmt` and `goimports`
   - Follow [Effective Go](https://golang.org/doc/effective_go)
   - Use [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

2. **Naming Conventions**
   ```go
   // Good: Exported names are capitalized
   type PlayerStats struct { ... }
   
   // Good: Unexported names are lowercase
   func validateRankingWeights(...) error { ... }
   
   // Good: Interface names end with -er
   type Calculator interface { ... }
   ```

3. **Error Handling**
   ```go
   // Good: Wrap errors with context
   if err := db.Query(...); err != nil {
       return fmt.Errorf("query player stats: %w", err)
   }
   
   // Bad: Swallow errors
   db.Query(...)  // No error handling
   ```

4. **Context Propagation**
   ```go
   // Good: Pass context as first parameter
   func (s *Service) RecordStats(ctx context.Context, stats Stats) error {
       // Use ctx for cancellation, deadlines
   }
   
   // Bad: No context
   func (s *Service) RecordStats(stats Stats) error { ... }
   ```

5. **Package Organization**
   - One concept per package
   - Avoid circular dependencies
   - Keep packages small and focused

### Clean Architecture Rules

1. **Domain Layer** (`internal/domain/`)
   - No external dependencies (only stdlib)
   - Pure business logic
   - Rich entities with behavior
   - Define repository interfaces here

2. **Application Layer** (`internal/app/`)
   - Orchestrates domain entities
   - Implements use cases
   - Depends only on domain layer
   - No HTTP, DB, or framework code

3. **Infrastructure Layer** (`internal/infra/`)
   - Implements domain interfaces
   - Framework-specific code (HTTP, DB, Redis)
   - External service integrations
   - Depends on domain/app layers

### Dependency Direction

```
infra → app → domain
  ↓      ↓       ↓
  ✗      ✗      (no deps)
```

Domain should never import from app or infra!

## Testing Guidelines

### Unit Tests

```go
func TestGame_UpdateWeights(t *testing.T) {
    t.Parallel()  // Run tests in parallel
    
    tests := []struct {
        name          string
        weights       RankingWeights
        expectedError error
    }{
        {
            name:          "valid weights",
            weights:       RankingWeights{"kd": 0.5, "avg": 0.5},
            expectedError: nil,
        },
        // More test cases...
    }
    
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Table-Driven Tests

All tests should use table-driven approach for consistency.

### Test Coverage

- Aim for >80% coverage on domain layer
- 100% coverage on critical business logic (ranking calculations)
- Integration tests for repositories

### Running Tests

```bash
# All tests
make test

# With coverage
make coverage

# With race detector (CI requirement)
make test-race
```

## Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat:` New feature
- `fix:` Bug fix
- `refactor:` Code refactoring (no functional changes)
- `test:` Adding or updating tests
- `docs:` Documentation changes
- `chore:` Maintenance tasks (dependencies, build)
- `perf:` Performance improvements
- `ci:` CI/CD changes

### Scopes (Optional)

- `domain`: Domain layer changes
- `app`: Application layer changes
- `infra`: Infrastructure layer changes
- `ranking`: Ranking system
- `auth`: Authentication
- `stats`: Statistics system
- `db`: Database changes

### Examples

```
feat(ranking): add Fortnite ranking calculator

Implements FortniteCalculator with elimination-based scoring.
Adds support for Fortnite-specific metrics: eliminations,
placement, and builds.

Closes #24
```

```
fix(stats): correct K/D ratio calculation for zero deaths

Handles edge case where player has zero deaths. Returns
kills instead of dividing by zero.

Fixes #31
```

```
refactor(domain): extract tier determination to helper

Moves tier determination logic to DetermineTierByPercentile
function for reusability across calculators.
```

## Pull Request Process

### Before Submitting

1. ✅ All tests pass (`make test-race`)
2. ✅ Code is formatted (`make fmt`)
3. ✅ Linter passes (`make lint`)
4. ✅ Documentation updated (if applicable)
5. ✅ Commit messages follow convention

### PR Title

Use conventional commit format:

```
feat(ranking): add Apex Legends calculator
fix(auth): resolve JWT expiration bug
docs: update architecture diagram
```

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Refactoring
- [ ] Documentation
- [ ] Other

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] All tests passing
- [ ] Code formatted and linted
- [ ] Follows architecture guidelines

## Related Issues
Closes #issue_number

## Screenshots (if applicable)
```

### Review Process

1. At least one approval required
2. All checks must pass (tests, linter)
3. No merge conflicts
4. Follows architecture and coding standards

## Architecture Guidelines

### Adding a New Game

1. **No code changes needed!** Just add to database:

```sql
INSERT INTO games (name, slug, stat_schema, ranking_weights) VALUES (
    'Fortnite',
    'fortnite',
    '{"eliminations": {...}, "placement": {...}}',
    '{"eliminations": 0.35, "placement": 0.35, "builds": 0.30}'
);
```

2. **Optional: Add custom calculator** (if default is insufficient):

```go
// internal/domain/ranking/fortnite_calculator.go
type FortniteCalculator struct{}

func (fc *FortniteCalculator) Calculate(ctx context.Context, stats *player.PlayerStats, game *game.Game) (float64, error) {
    // Game-specific ranking logic
}

func (fc *FortniteCalculator) SupportsGame(gameSlug string) bool {
    return gameSlug == "fortnite"
}
```

3. **Register calculator** in main.go:

```go
rankingService := ranking.NewService(
    ranking.NewWarzoneCalculator(),
    ranking.NewFortniteCalculator(),
    ranking.NewDefaultCalculator(),
)
```

### Adding a New Domain Entity

1. Create in `internal/domain/entity_name/`
2. Define rich entity with behavior
3. Add validation in constructors
4. Write comprehensive tests
5. Define repository interface (if needed)

### Adding a New Use Case

1. Create in `internal/app/use_case_name/`
2. Define service struct with dependencies
3. Implement business logic
4. Add integration tests
5. Document in godoc

### Adding a New HTTP Endpoint

1. Define handler in `internal/infra/http/handlers/`
2. Add route in router
3. Validate input (use DTOs)
4. Call application service
5. Return appropriate status codes
6. Add API tests

## Questions or Issues?

- **Issues**: [GitHub Issues](https://github.com/yourusername/tourney-rank/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/tourney-rank/discussions)
- **Documentation**: [docs/](docs/)

## Recognition

Contributors will be acknowledged in:
- README.md contributors section
- Release notes
- GitHub contributors page

Thank you for contributing to TourneyRank! 🎮

# TourneyRank

**Multi-Game Tournament Management Platform with Automated Ranking System**

TourneyRank is a comprehensive web application designed to manage competitive gaming tournaments across multiple titles, featuring automated ranking calculations, intelligent team formation, and real-time leaderboards.

## Features

### Multi-Game Support
- **Flexible Architecture**: Designed to support multiple competitive games
- **Initial Support**: Call of Duty: Warzone
- **Planned**: Fortnite, Apex Legends, Valorant, CS:GO
- **Customizable Metrics**: Each game has its own stat schema and ranking weights

### Tournament Modes
- **2vs2 Kill Race**: Elimination-based competition for 2-player teams
- **Customs**: Custom matches with specific rules
- **4vs4+ Multiplayer**: Teams of 4+ players (configurable per game)

### Scoring Formats
- **Matchpoint**: Points per won match
- **Killpoint**: Score based on accumulated eliminations
- **Phases**: Tournaments with groups, playoffs, and finals
- **Direct**: No phases, general table

### Automated Ranking System
The application automatically calculates and updates player rankings based on received statistics:

- K/D Ratio (Kills/Deaths)
- Average kills per match
- Total and average damage
- Performance consistency
- Team winrate participation
- Weighted composite score (configurable per game)

**Skill Tiers**:
- Elite (Top 5%)
- Advanced (Top 20%)
- Intermediate (Top 50%)
- Beginner (Rest)

### Intelligent Team Formation
- **Balanced by Ranking**: Algorithm distributes players by skill level
- **Animated Roulette**: Traditional team selection with visual animation
- **Manual Assisted**: Drag & drop with balance indicators
- **Random**: Simple random assignment for casual matches

### Real-Time Features
- **WebSocket Updates**: Live leaderboard updates
- **Match Timeline**: Real-time match progress tracking
- **Instant Stats**: Automatic ranking recalculation on stat submission

### Public Registration System
- Public inscription forms with validation
- Per-game platform ID validation (Activision ID, Epic, Steam, etc.)
- Optional invitation code system
- Configurable player limits per tournament

### Betting/Predictions System
- Virtual points betting (no real money)
- Bet on match winners, MVP, team stats
- Dynamic multipliers based on probabilities
- Bettor leaderboard

### External Integrations
- **Discord**: Webhooks for notifications via n8n
- **LLM**: Automated match analysis and insights
- **JSON API**: Automated stat input from external sources

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Architecture**: Clean Architecture + Domain-Driven Design
- **Database**: PostgreSQL (with JSONB for flexible schemas)
- **Cache**: Redis (for rankings and leaderboards)
- **Real-time**: WebSockets
- **Patterns**: Strategy Pattern for game-specific logic, Event-Driven

### Frontend
- **Framework**: React 18
- **Build Tool**: Vite
- **Styling**: CSS Modules / Tailwind CSS
- **State**: Context API + React Query
- **Real-time**: WebSocket client

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **Migrations**: golang-migrate

## Project Structure

```
tourney-rank/
├── cmd/
│   └── service/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Domain layer (pure Go, no external deps)
│   │   ├── game/                # Game aggregate
│   │   ├── tournament/          # Tournament aggregate
│   │   ├── player/              # Player aggregate
│   │   ├── team/                # Team aggregate
│   │   ├── match/               # Match aggregate
│   │   └── ranking/             # Ranking calculation strategies
│   ├── app/                     # Application layer (use cases)
│   │   ├── tournament/
│   │   ├── team/
│   │   ├── stats/
│   │   └── ranking/
│   ├── infra/                   # Infrastructure layer (adapters)
│   │   ├── http/                # REST API handlers
│   │   ├── websocket/           # WebSocket handlers
│   │   ├── postgres/            # PostgreSQL repositories
│   │   ├── redis/               # Redis cache
│   │   └── discord/             # Discord webhook client
│   └── config/                  # Configuration management
├── migrations/                  # Database migrations
├── web/                         # Frontend application
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/              # Reusable components
│   │   │   ├── tournament/
│   │   │   ├── leaderboard/
│   │   │   └── teams/
│   │   ├── hooks/
│   │   ├── services/
│   │   └── contexts/
│   └── public/
├── api/                         # API specifications
│   └── openapi/
├── test/                        # Integration tests
├── docker-compose.yml
├── Makefile
└── go.mod
```

## API Endpoints (Preview)

```
# Game Management
GET    /api/games                # List supported games
POST   /api/games                # Add new game (admin)
GET    /api/games/:id            # Get game details

# Tournament Management
POST   /api/tournaments          # Create tournament
GET    /api/tournaments/:id      # Get tournament details
PUT    /api/tournaments/:id      # Update tournament

# Match Stats
POST   /api/matches/stats        # Submit match statistics (JSON)
GET    /api/matches/:id          # Get match details

# Rankings & Leaderboards
GET    /api/rankings/players     # Get player rankings (filterable by game)
GET    /api/leaderboard/:tournamentId  # Get tournament leaderboard
WS     /ws/leaderboard/:tournamentId   # Real-time leaderboard updates

# Team Management
POST   /api/teams/generate       # Generate balanced teams
GET    /api/teams/:id            # Get team details

# Registration
POST   /api/registrations        # Public player registration
GET    /api/registrations/:tournamentId  # Get tournament registrations

# Betting
POST   /api/bets                 # Place bet
GET    /api/bets/user/:userId    # Get user's bets
```

## Roadmap

### Phase 1: Multi-Game Core (Weeks 1-2)
- [x] Project setup and architecture
- [ ] User system and roles
- [ ] CRUD for supported games
- [ ] CRUD for tournaments with game selection
- [ ] Flexible stats input (manual/JSON)

### Phase 2: Intelligence (Weeks 3-4)
- [ ] Automated ranking system per game
- [ ] Intelligent team generation
- [ ] Animated roulette
- [ ] Real-time leaderboards

### Phase 3: Engagement (Weeks 5-6)
- [ ] Public registration
- [ ] Betting system
- [ ] Discord/n8n integrations

### Phase 4: Expansion (Future)
- [ ] Second game implementation (Fortnite/Apex)
- [ ] Cross-game comparisons
- [ ] Stream overlays

## Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/tourney-rank.git
cd tourney-rank

# Start infrastructure
docker-compose up -d

# Run migrations
make migrate-up

# Start backend
make run

# Start frontend (in another terminal)
cd web
npm install
npm run dev
```

### Development

```bash
# Run tests
make test

# Run with race detector
make test-race

# Lint code
make lint

# Format code
make fmt
```

## Configuration

Configuration is managed via environment variables and config files:

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/tourneyrank

# Redis
REDIS_URL=redis://localhost:6379

# Server
HTTP_PORT=8080
WS_PORT=8081

# External Services
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Commit Convention
Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` New feature
- `fix:` Bug fix
- `refactor:` Code refactoring
- `test:` Adding tests
- `docs:` Documentation changes

## License

MIT License - see LICENSE file for details

## Contact

Project Link: [https://github.com/yourusername/tourney-rank](https://github.com/yourusername/tourney-rank)

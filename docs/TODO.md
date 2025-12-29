# TourneyRank Development Roadmap

## Current Status

✅ **Completed**:
- [x] Project structure setup
- [x] Domain entities: Game, Player, PlayerStats
- [x] Ranking system architecture (Strategy Pattern)
- [x] Warzone ranking calculator implementation
- [x] Docker Compose setup (MongoDB, Redis)
- [x] Makefile with development commands
- [x] Architecture documentation
- [x] Configuration management (env loading, validation)
- [x] MongoDB connection manager with retry logic

## Phase 1: Core Multi-Game Foundation (Weeks 1-2)

### Week 1: Infrastructure & Authentication

- [x] **Configuration Management**
  - [x] Create `internal/config/config.go` with env loading
  - [x] Validate required environment variables
  - [x] Support multiple environments (dev, staging, prod)

- [ ] **Database Layer (MongoDB)**
  - [x] Create `internal/infra/mongodb/client.go`
  - [ ] Implement Game Repository
  - [ ] Implement User Repository
  - [ ] Implement Player Repository
  - [x] Add health check endpoint for DB

- [ ] **Authentication System**
  - [ ] User registration endpoint
  - [ ] Login with JWT token generation
  - [ ] Password hashing with bcrypt
  - [ ] Role-based access control (RBAC) middleware
  - [ ] Refresh token mechanism

- [ ] **Testing Setup**
  - [ ] Add testcontainers for integration tests
  - [ ] Create test fixtures and factories
  - [ ] Setup CI pipeline (GitHub Actions)

**Deliverable**: Basic API with user authentication

### Week 2: Game & Tournament Management

- [ ] **Game Management (Admin)**
  - [ ] POST `/api/games` - Create new game
  - [ ] GET `/api/games` - List games
  - [ ] GET `/api/games/:id` - Get game details
  - [ ] PUT `/api/games/:id` - Update game config
  - [ ] PATCH `/api/games/:id/activate` - Activate/deactivate

- [ ] **Tournament CRUD**
  - [ ] Create Tournament Repository
  - [ ] POST `/api/tournaments` - Create tournament
  - [ ] GET `/api/tournaments` - List tournaments (with filters)
  - [ ] GET `/api/tournaments/:id` - Get tournament details
  - [ ] PUT `/api/tournaments/:id` - Update tournament
  - [ ] DELETE `/api/tournaments/:id` - Delete tournament

- [ ] **Frontend Setup**
  - [ ] Initialize React + Vite project
  - [ ] Setup routing (React Router)
  - [ ] Create layout components
  - [ ] Add Tailwind CSS / styling system
  - [ ] Create game selector component

**Deliverable**: Admin can create/manage games and tournaments

## Phase 2: Stats & Intelligent Ranking (Weeks 3-4)

### Week 3: Stats Submission & Processing

- [ ] **Match Stats Endpoint**
  - [ ] Create Match entity and repository
  - [ ] POST `/api/matches/stats` - Submit match stats (JSON)
  - [ ] Validate stats against game schema
  - [ ] Manual stats entry form (team leader)
  - [ ] Stats validation and error handling

- [ ] **Ranking Calculation Engine**
  - [ ] Implement `RankingService` in application layer
  - [ ] Add Fortnite Calculator (second game)
  - [ ] Add Apex Legends Calculator
  - [ ] Add Generic Calculator as fallback
  - [ ] Create ranking calculation tests

- [ ] **Stats Aggregation**
  - [ ] Update player_stats on match recorded
  - [ ] Calculate total kills, deaths, damage
  - [ ] Track matches played count
  - [ ] Update last_match_at timestamp

- [ ] **Background Job System**
  - [ ] Setup async job processor (e.g., Asynq)
  - [ ] Background ranking recalculation
  - [ ] Batch updates for performance

**Deliverable**: Match stats can be submitted and rankings auto-calculated

### Week 4: Leaderboards & Real-Time Updates

- [ ] **Leaderboard API**
  - [ ] GET `/api/leaderboard/:tournamentId` - Tournament leaderboard
  - [ ] GET `/api/rankings/players?game=warzone` - Global rankings
  - [ ] Pagination support
  - [ ] Filter by tier, date range

- [ ] **Redis Caching**
  - [ ] Create `internal/infra/redis/cache.go`
  - [ ] Cache leaderboards in sorted sets
  - [ ] Cache player stats
  - [ ] Cache invalidation on updates

- [ ] **WebSocket Server**
  - [ ] Create `internal/infra/websocket/server.go`
  - [ ] WS `/ws/leaderboard/:tournamentId` - Subscribe to updates
  - [ ] Pub/sub for leaderboard broadcasts
  - [ ] Connection management

- [ ] **Frontend Leaderboards**
  - [ ] Leaderboard component with live updates
  - [ ] Player stats card component
  - [ ] Tier badges (Elite, Advanced, etc.)
  - [ ] Game filter/selector
  - [ ] WebSocket client integration

**Deliverable**: Live leaderboards with real-time updates

## Phase 3: Team Formation & Registration (Weeks 5-6)

### Week 5: Intelligent Team Generation

- [ ] **Team Domain Logic**
  - [ ] Create Team entity and repository
  - [ ] Team member management
  - [ ] Team stats aggregation

- [ ] **Team Formation Algorithms**
  - [ ] Create `internal/app/team/formation.go`
  - [ ] Balanced algorithm (K/D-based distribution)
  - [ ] Random algorithm
  - [ ] Manual algorithm with validation
  - [ ] Serpentine draft for fairness

- [ ] **Team Generation API**
  - [ ] POST `/api/teams/generate` - Generate balanced teams
  - [ ] GET `/api/teams/:id` - Get team details
  - [ ] PUT `/api/teams/:id/members` - Update team members
  - [ ] Balance indicator calculation

- [ ] **Animated Roulette**
  - [ ] Frontend roulette component with animation
  - [ ] Sequential player selection UI
  - [ ] Team assignment visualization
  - [ ] Export team composition

**Deliverable**: Automated balanced team generation

### Week 6: Public Registration & Betting

- [ ] **Registration System**
  - [ ] Create Registration entity and repository
  - [ ] POST `/api/registrations` - Public registration
  - [ ] GET `/api/registrations/:tournamentId` - List registrations
  - [ ] PATCH `/api/registrations/:id/approve` - Approve registration
  - [ ] Email validation
  - [ ] Invitation code support

- [ ] **Betting System**
  - [ ] Create Bet entity and repository
  - [ ] Virtual points economy
  - [ ] POST `/api/bets` - Place bet
  - [ ] GET `/api/bets/user/:userId` - User's bets
  - [ ] GET `/api/bets/leaderboard` - Top bettors
  - [ ] Bet settlement on match completion
  - [ ] Dynamic multipliers

- [ ] **Frontend**
  - [ ] Registration form component
  - [ ] Betting interface
  - [ ] User points display
  - [ ] Bet history

**Deliverable**: Public can register and bet on tournaments

## Phase 4: External Integrations (Week 7)

- [ ] **Discord Integration**
  - [ ] Create `internal/infra/discord/client.go`
  - [ ] Webhook for match start notifications
  - [ ] Leaderboard update announcements
  - [ ] Tournament reminders

- [ ] **n8n Automation**
  - [ ] Setup n8n workflows
  - [ ] Match stats webhook receiver
  - [ ] Automated social media posts
  - [ ] Email notifications

- [ ] **LLM Integration (Optional)**
  - [ ] Match analysis endpoint
  - [ ] Player performance insights
  - [ ] Tournament summaries
  - [ ] OpenAI/Anthropic API integration

**Deliverable**: Automated notifications and insights

## Phase 5: Expansion & Polish (Week 8+)

- [ ] **Second Game Implementation**
  - [ ] Add Fortnite to games table
  - [ ] Verify Fortnite calculator
  - [ ] Test multi-game tournaments

- [ ] **Advanced Features**
  - [ ] Tournament brackets visualization
  - [ ] Stream overlay generation
  - [ ] Player certificates (PDF)
  - [ ] Cross-game comparisons

- [ ] **Performance Optimization**
  - [ ] Database query optimization
  - [ ] Add database indexes
  - [ ] Redis cache tuning
  - [ ] Load testing

- [ ] **Security Hardening**
  - [ ] Rate limiting
  - [ ] Input sanitization
  - [ ] SQL injection prevention
  - [ ] CORS configuration
  - [ ] Security headers

- [ ] **Documentation**
  - [ ] API documentation (OpenAPI/Swagger)
  - [ ] User guide
  - [ ] Admin guide
  - [ ] Deployment guide

**Deliverable**: Production-ready application

## Current Next Steps

**Immediate Priorities** (This Week):

1. **Configuration Layer**
   ```bash
   # Create config package
   touch internal/config/config.go
   ```
   - Load environment variables
   - Validate required configs
   - Support .env file

2. **Database Connection**
   ```bash
   # Create postgres infrastructure
   touch internal/infra/postgres/connection.go
   touch internal/infra/postgres/game_repository.go
   ```
   - PostgreSQL connection pool
   - Repository implementations

3. **HTTP Server Setup**
   ```bash
   # Create HTTP infrastructure
   touch internal/infra/http/server.go
   touch internal/infra/http/middleware/auth.go
   touch internal/infra/http/handlers/game_handler.go
   ```
   - Gorilla Mux router
   - Middleware chain
   - First endpoints

4. **Run Migrations**
   ```bash
   make docker-up
   make migrate-up
   ```

5. **First API Endpoint**
   - GET `/api/games` - List supported games
   - Test with default Warzone game from migrations

## Testing Checklist

- [ ] Unit tests for all domain entities
- [ ] Unit tests for ranking calculators
- [ ] Integration tests for repositories
- [ ] API endpoint tests
- [ ] WebSocket connection tests
- [ ] Load tests for leaderboard queries
- [ ] Security tests (OWASP Top 10)

## Deployment Checklist

- [ ] Dockerfile for backend
- [ ] Dockerfile for frontend
- [ ] docker-compose for production
- [ ] Environment variable documentation
- [ ] Database backup strategy
- [ ] Monitoring setup (Prometheus/Grafana)
- [ ] Logging aggregation
- [ ] SSL/TLS certificates
- [ ] Domain configuration

## Future Enhancements

- [ ] Mobile app (React Native)
- [ ] Machine learning predictions
- [ ] Player behavior analytics
- [ ] Tournament templates
- [ ] Multi-language support (i18n)
- [ ] Dark mode
- [ ] Social features (friend system)
- [ ] Achievement system
- [ ] Season/ladder system

---

**Last Updated**: Initial setup
**Current Phase**: Phase 1 - Week 1

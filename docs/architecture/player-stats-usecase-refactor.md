# Player stats + handlers refactor plan

Date: 2026-02-07
Scope: backend only
Goal: decouple HTTP handlers from Mongo repositories and move player stats logic into usecase layer with explicit domain ports.

## Why
- Reduce coupling between HTTP and MongoDB.
- Centralize player stats and leaderboard logic in usecase layer.
- Keep domain pure and infra focused on persistence + transport.

## Current state
- Player and leaderboard handlers call Mongo repositories directly.
- Player stats logic and leaderboard aggregation live in Mongo repository.
- Domain ranking service is not used by handlers.

## Target architecture
- Usecase: new service(s) to orchestrate player stats, leaderboard reads, and ranking updates.
- Domain: repository interface(s) for player stats and leaderboard queries.
- Infra: Mongo repositories implement new interfaces; handlers depend on usecase services only.

## Proposed steps
1) Define domain ports
   - Create a player stats repository interface in internal/domain/player (or internal/domain/ranking if you prefer).
   - Include methods used by handlers: GetByPlayer, GetByPlayerAndGame, GetLeaderboard, GetLeaderboardByTier, GetPlayerRank, CountByGame, GetTierDistribution, Create, Update, UpdateRanking, IncrementStats, GetOrCreate.

2) Add usecase service
   - Create internal/usecase/leaderboard/service.go (or playerstats/service.go).
   - Implement functions for leaderboard pages, player rank, and player stats summary.
   - Keep DTOs in usecase layer and map domain entities to response DTOs.

3) Rewire handlers
   - Update LeaderboardHandler to depend on usecase service instead of Mongo repo.
   - Update PlayerHandler to use usecase service for stats/leaderboard read paths.
   - Keep auth and request parsing in handlers only.

4) Move Mongo repository behind ports
   - Make mongodb.PlayerStatsRepository implement the new domain interface.
   - Avoid exposing Mongo-specific structs to handlers or usecase.

5) Tests
   - Add usecase unit tests with fake repositories.
   - Keep repo tests (if needed) isolated to infra.

## Non-goals (for now)
- Refactor game endpoints to usecase.
- Remove bson/json tags from domain entities.
- Introduce Redis caching.

## Risks
- Method signature drift between handler and new usecase.
- DTO mapping inconsistencies across endpoints.

## Validation
- Handlers import only usecase + middleware + stdlib.
- Usecase depends only on domain interfaces.
- Mongo repositories implement domain ports and are wired in main.

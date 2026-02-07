# 🚧 Current System Status

> [!NOTE]
> **Updated**: Refactoring Complete. HTTP handlers decoupled from MongoDB repositories via usecase layer. Compilation successful.

This document serves as the source of truth for what currently exists vs. what is planned.

## 📉 Gap Analysis: Promised vs. Implemented

| Feature Area | README Description | Current Code Status | Notes |
| :--- | :--- | :--- | :--- |
| **Project Structure** | legacy docs mentioned an older structure and extra domain packages | ⚠️ **Partial** | Current structure is `internal/{domain,usecase,infra,config}` + `frontend/`. |
| **API Endpoints** | REST API for Games, Leaderboards | ✅ **Implemented** | HTTP handlers for Games and Leaderboard CRUD. |
| **Game Domain** | Flexible Schema, Ranking Weights | ✅ **Implemented** | `internal/domain/game` has robust logic. |
| **Ranking Logic** | Strategy Pattern for generic/specific ranking | ✅ **Implemented** | `internal/domain/ranking` implements the Strategy pattern. |
| **Database** | MongoDB + Redis | ✅ **MongoDB Done** | MongoDB repositories implemented. Redis caching pending. |
| **Frontend** | React + Vite Dashboard | ⚠️ **Partial** | Auth flows/pages + API services exist; feature coverage still limited. |

## ✅ Working Components

The following subsystems are implemented and tested:

### 1. Game Domain (`internal/domain/game`)
*   **Entity**: `Game` struct with support for flexible `StatSchema` and `RankingWeights`.
*   **Logic**: Validation of stats against the schema, weight updates.

### 2. Player Domain (`internal/domain/player`)
*   **Entity**: `Player` and `PlayerStats` structs with tier system.
*   **Logic**: Stats tracking, tier calculation (Bronze to Master).

### 3. Ranking Strategy (`internal/domain/ranking`)
*   **Pattern**: Strategy Pattern (`Calculator` interface).
*   **Implementations**:
    *   `DefaultCalculator`: Basic K/D ratio calculation.
    *   `WarzoneCalculator`: Weighted score based on K/D, Damage, Kills, etc.

### 4. Configuration (`internal/config`)
*   **MongoDB**: Connection URI and database name configuration.
*   **Redis**: Cache layer configuration (not yet connected).
*   **Environment**: Support for dev/staging/prod environments.

### 5. MongoDB Persistence (`internal/infra/mongodb`)
*   **Client**: Connection manager with retry logic and health checks.
*   **GameRepository**: Full CRUD operations with slug lookup and indexes.
*   **PlayerRepository**: CRUD with search and platform ID lookups.
*   **PlayerStatsRepository**: Stats persistence with aggregation pipelines for leaderboards.

### 6. HTTP API Layer (`internal/infra/http`)
*   **Game Endpoints**:
    *   `GET /api/v1/games` - List all games
    *   `POST /api/v1/games` - Create a new game
    *   `GET /api/v1/games/{id}` - Get game by ID or slug
    *   `PATCH /api/v1/games/{id}/status` - Update game status
    *   `DELETE /api/v1/games/{id}` - Delete a game
*   **Leaderboard Endpoints**:
    *   `GET /api/v1/leaderboard/{gameId}` - Get leaderboard with pagination
    *   `GET /api/v1/leaderboard/{gameId}/tier/{tier}` - Leaderboard by tier
    *   `GET /api/v1/leaderboard/{gameId}/player/{playerId}` - Get player rank
    *   `GET /api/v1/leaderboard/{gameId}/tiers` - Tier distribution
*   **Health Endpoints**:
    *   `GET /healthz` - Liveness probe
    *   `GET /readyz` - Readiness probe (checks MongoDB)

## 🔜 Next High-Priority Steps

To move from "MVP" to "Production Ready":

1.  **Add Player Endpoints**: Create REST endpoints for player management.
2.  **Implement Match Submission**: Add endpoint to record matches and update stats.
3.  **Add Redis Cache**: Cache leaderboard results for better performance.
4.  **Add Authentication**: Implement JWT-based auth for admin endpoints.
5.  **Add Tournament Domain**: Create `internal/domain/tournament` for competition logic.
6.  **E2E Tests**: Add integration tests with real MongoDB.

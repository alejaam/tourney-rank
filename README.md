# TourneyRank

Multi-game tournament platform (Go backend + React frontend) with automated ranking calculations.

## Tech Stack

- Backend: Go (Clean Architecture)
- Data: MongoDB
- Cache (planned / optional): Redis
- Frontend: React + Vite

## Quick Start

### Backend (local)

```bash
make infra-up
make run
```

Health endpoints:

- `GET http://localhost:8080/healthz`
- `GET http://localhost:8080/readyz`

### Frontend (dev)

```bash
cd frontend
npm install
npm run dev
```

### Tests

```bash
make test
```

## Documentation

- [docs/QUICKSTART.md](docs/QUICKSTART.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- [docs/CURRENT_STATUS.md](docs/CURRENT_STATUS.md)
- [docs/ROADMAP.md](docs/ROADMAP.md)

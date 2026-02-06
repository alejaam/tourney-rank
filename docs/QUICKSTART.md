# TourneyRank Quick Start Guide

Get TourneyRank running locally (backend + infrastructure, plus optional frontend dev server).

## Prerequisites

- Go 1.21+
- Docker + Docker Compose
- Make (recommended)
- Node.js (for the frontend)

## 1) Clone

```bash
git clone https://github.com/alejaam/tourney-rank.git
cd tourney-rank
```

## 2) Environment

```bash
cp .env.example .env
```

Defaults work for local development, but the important values are:

- `MONGODB_URI`
- `MONGODB_DATABASE`
- `REDIS_URL`
- `JWT_SECRET`

## 3) Start infrastructure (MongoDB + Redis)

```bash
make infra-up
```

Verify:

```bash
docker compose ps
```

## 4) Run the backend

```bash
make run
```

Verify:

```bash
curl -sSf http://localhost:8080/healthz
curl -sSf http://localhost:8080/readyz
```

## 5) Run the frontend (optional)

```bash
cd frontend
npm install
npm run dev
```

## 6) Run tests

```bash
make test
```
# Try again

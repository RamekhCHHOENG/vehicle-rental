# Car Rental Marketplace

A two-sided vehicle rental marketplace: owners list cars/motorbikes, renters book them,
and an admin verifies every listing before it goes live.

## Stack

- **Backend**: Go (chi router + GORM)
- **Database**: PostgreSQL
- **Frontend**: Nuxt + Nuxt UI
- **Dev environment**: Docker Compose

## Getting started

```bash
cp .env.example .env   # then edit values
docker compose up
```

- API: http://localhost:8080
- Web: http://localhost:3000

## Project layout

```
backend/    Go API (cmd/api is the entrypoint)
frontend/   Nuxt app
```

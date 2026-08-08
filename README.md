# Yan — inspected vehicle rental

A two-sided vehicle rental marketplace: owners list cars/motorbikes, renters book
them in 3 steps with transparent pricing, and an **admin verifies every listing
before it goes live** — the platform's core trust feature.

## Stack

| Layer     | Tech                                      |
|-----------|-------------------------------------------|
| Backend   | Go · chi router · GORM                    |
| Database  | PostgreSQL 17                             |
| Frontend  | Nuxt 4 · Nuxt UI · Tailwind CSS · Archivo |
| Auth      | JWT in an httpOnly cookie · role-based    |
| Dev env   | Docker Compose (hot reload on both sides) |
| Deploy    | VPS · Docker Compose · Caddy (auto-HTTPS) |

## Quick start (development)

```bash
cp .env.example .env        # then edit values
docker compose up
```

That builds and starts all three services with hot reload on both apps.

- Web: http://localhost:3000
- API: http://localhost:8090 (`/api/health` to check)
- Postgres: localhost:5434

All three publish on `127.0.0.1` only. Docker's default is `0.0.0.0`, which
would put the database — dev password and all — on the public internet whenever
the host is reachable.

Two addresses point at the API because SSR and the browser reach it differently:
`NUXT_PUBLIC_API_BASE` is what the visitor's browser calls, while
`NUXT_API_INTERNAL` (`http://api:8080`) is what the Nuxt server calls over the
compose network. `useApi` picks between them.

The admin account is seeded from `ADMIN_EMAIL` / `ADMIN_PASSWORD` in `.env`
on first startup.

You can also run the pieces natively (Postgres still via Docker):

```bash
docker compose up -d postgres
cd backend && set -a && source ../.env && set +a && go run ./cmd/api
```

```bash
cd frontend && npm install && npm run dev
```

## Tests

```bash
cd backend && go test ./...
```

Covers the booking rules (price, day counting, date-overlap) and the auth
middleware (valid/expired/forged tokens, role checks) with table-driven tests.

## Project layout

```
backend/
  cmd/api/main.go        entrypoint: config → DB → routes → listen
  internal/
    config/              env var loading
    database/            GORM connect, AutoMigrate, admin seed
    models/              User, Vehicle, VehiclePhoto, Booking, Review
    handlers/            auth, vehicles, uploads, bookings, admin
    middleware/          JWT auth + role-based access control
    httputil/            JSON response helpers
frontend/
  app/
    pages/               file-based routes (public, owner/, admin/, dashboard/)
    components/          VehicleCard, BookingForm, VehicleForm, StatusBadge
    composables/         useApi (cookie-aware fetch), useAuth, usePhotoUrl
    middleware/          auth / owner / admin route guards
deploy/Caddyfile         production reverse proxy + HTTPS
docker-compose.yml       development (hot reload)
docker-compose.prod.yml  production (multi-stage builds + Caddy)
```

## Design system — "Inspected"

The visual layer starts from `@ramekhchhoeng/designkit` — its frosted-glass
surfaces, tight tracking, press-on-tap feedback and system body font are kept.
The radii are deliberately tighter than the kit's: records and forms have crisp
corners, not bubbly ones.

- **Palette** — petrol ink `#0a2422`, paper `#fcfbf8`, and a single saffron
  accent `#f2a93b`. Actions and the verification seal share that one colour, so
  pressing the button and trusting the platform are the same gesture. Saffron is
  light, so anything sitting on it takes ink text, never white.
- **Type** — Archivo (variable, width axis) for display, set wide with tight
  tracking so headings read like signage rather than editorial. The designkit's
  system stack stays for body. `.numeric` applies tabular figures so prices and
  dates line up like a receipt.
- **Signature** — `VerifiedSeal`, a stamped "Inspected" mark used **once per
  surface**, over vehicle photography. Pass `on-photo` when it sits on an image
  so it gets its own dark ground.
- **Radius** — four values, all in `main.css`: `--r-surface` 12px (cards,
  dialogs), `--r-control` 8px (buttons, inputs), `--r-media` 10px (photos),
  `--r-small` 5px (badges, the seal, numbered markers). Nothing hardcodes a
  radius, so the whole system retunes from those four lines.
- **Logo** — `AppLogo`. A licence plate whose checkmark breaks out through the
  top edge; the plate stroke is masked along the check's path so the two read as
  one object. Rename the product in one place: the `name` prop default.
- **Name** — Yan (យាន), Khmer for "vehicle" — it covers motorbikes as well as
  cars, which "CarRental" did not.
- **Themes** — light and dark, toggled in the header and persisted by
  `@nuxtjs/color-mode`. Dark mode grounds on deep petrol `#061715`, not black.

Tokens live in `frontend/app/assets/css/main.css`; component shapes in
`frontend/app/app.config.ts`.

## Roles & flow

1. **Owner** signs up → lists a vehicle with photos → status `pending`.
2. **Admin** reviews the queue → approves (goes public) or rejects with a reason.
3. **Renter** browses approved vehicles → books in 3 steps (dates → review →
   confirm). Price is computed server-side; dates clashing with a confirmed
   booking are refused.
4. **Owner** confirms or rejects the request, and marks the vehicle returned.
5. **Renter** leaves a 1–5★ review shown on the public listing.

## API overview

| Method & path | Who | What |
|---|---|---|
| POST `/api/auth/register` · `/login` · `/logout`, GET `/me` | public / any | session management |
| GET `/api/vehicles`, `/api/vehicles/{id}` | public | browse approved listings |
| POST/GET/PUT/DELETE `/api/owner/vehicles…` + `/photos` | owner | manage own listings |
| GET `/api/owner/bookings`, POST `…/confirm` `…/reject` `…/complete` | owner | handle requests |
| POST/GET `/api/bookings`, POST `…/cancel` `…/review` | renter | book, cancel, review |
| GET `/api/admin/vehicles` `…/users` `…/stats`, POST `…/approve` `…/reject` | admin | verification, users & metrics |

### What the admin can do

- **Verify listings** — review the pending queue, approve (goes public) or reject with a reason the owner sees.
- **Take a live listing down** — the same reject action works on an approved listing, so a vehicle that goes bad after approval can be pulled. The panel warns first if the vehicle has active bookings; existing bookings are left untouched.
- **Restore** — re-approve a rejected listing; the stale rejection reason is cleared.
- **Browse all users** — `GET /api/admin/users` (with `?role=` and `?q=` search) returns every user plus how many vehicles they list and bookings they have made. Columns are selected explicitly so the password hash can never be serialised.
- **See platform metrics** — users, vehicles, pending/approved counts, bookings, completions.

Not built yet: suspending users, changing someone's role, and an admin view of individual bookings.

## Deploying to a VPS (DigitalOcean, Hetzner, …)

1. Provision a small VPS (1–2 GB RAM), install Docker + the compose plugin.
2. Point your domain's DNS A record at the server's IP.
3. Clone the repo, then create `.env` with **production** values:
   strong `POSTGRES_PASSWORD`, long random `JWT_SECRET`, real `ADMIN_EMAIL`
   / `ADMIN_PASSWORD`, and `NUXT_PUBLIC_API_BASE=https://yourdomain.com`.
4. Edit `deploy/Caddyfile`: replace `example.com` with your domain.
5. Start everything:

   ```bash
   docker compose -f docker-compose.prod.yml up -d --build
   ```

   Caddy obtains the HTTPS certificate automatically. Uploaded photos and the
   database live in named Docker volumes and survive restarts.
6. Update after a code change:

   ```bash
   git pull && docker compose -f docker-compose.prod.yml up -d --build
   ```

**Backups**: `docker exec <postgres-container> pg_dump -U carrental carrental > backup.sql`
on a cron job, plus a copy of the `uploads` volume.

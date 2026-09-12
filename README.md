# KabadiConnect — E-Waste Recycler Website

KabadiConnect is an **e-waste formalization and traceability platform** that connects informal e-waste collectors/aggregators with authorized recyclers through transparent price information, material matching, recycler offers and traceable digital handovers.

This repository is **only** the Recycler Website plus the shared Go API/MongoDB contracts needed for integration. The collector/Kabadiwala mobile application is a separate team workstream and is not implemented here.

## Problem

Informal e-waste collectors have strong last-mile collection reach but limited access to authorized recycling channels, transparent prices, traceable handovers and formal transaction histories.

## Solution

KabadiConnect bridges collectors and authorized recyclers through digital e-waste lots, transparent price information, recycler matching, traceable handovers and transaction/payment records.

## Architecture

```text
Collector App (future, separate team)
            |
            v
       Go REST API
            |
            v
         MongoDB
            ^
            |
Recycler React Website (this repository)
```

Both clients share the same backend business entities. Collector mobile UI is deliberately absent from this repository.

## Technology stack

Frontend: React, TypeScript, Vite, React Router, CSS.

Backend: Go REST API, modular monolith, JWT access tokens with rotating HTTP-only refresh sessions.

Database: MongoDB, with Docker Compose for local development.

No microservices, Kafka, Kubernetes, blockchain or fake AI/ML.

## Canonical e-waste scope

Supported initial material categories:

- CRT
- LCD Panel
- PCB
- Cable
- Battery
- Motor
- Magnet-bearing Assembly
- Mixed Plastics from EEE
- Other E-Waste

Ordinary PET bottle waste, newspaper, normal cardboard, regular household glass and municipal plastic waste are outside product scope unless explicitly generated from end-of-life electrical/electronic equipment.

## Current verified-target vertical slice

```text
Recycler registers/logs in
→ dashboard shows e-waste/authorization data
→ opens E-Waste Lots
→ selects a lot such as Mixed PCBs
→ sees lot/material/collector/compatibility details
→ sends an offer
→ backend validates authorization/material/quantity/rate and stores offer
→ My Offers displays it as Pending
```

Collector acceptance remains outside this Recycler Website and is documented as a shared future contract.

## Folder structure

```text
backend/
  cmd/api/                  HTTP server
  cmd/seed/                 tagged development e-waste/price seed data
  internal/
    auth/                   authentication + refresh sessions
    users/                  recycler/collector identities and recycler profile
    listings/               canonical EWasteLot model; package name retained for migration compatibility
    offers/                 recycler offers
    prices/                 priceRecords dataset + board/history
    transactions/           immutable transaction/read model
    handovers/              traceability records + recycler confirmation
    payments/               recycler payment visibility
    safety/                 future shared safety-content model
    matching/               deterministic recycler eligibility/scoring preparation
    middleware/ database/ config/ httpx/ requestctx/
frontend/src/
  components/
  layouts/
  pages/
  services/
  state/
  types/
docs/
  MIGRATION.md
  SHARED_API_CONTRACT.md
  DATA_MODEL.md
  OFFLINE_SYNC.md
  FUTURE_AI_DATA.md
  MATCHING.md
```

## Non-destructive MongoDB migration

The physical V1 `scrapListings` collection is retained to avoid destructive development-data migration. New data uses the canonical `EWasteLot` fields while the Go model normalizes earlier fields. `/api/v1/lots` is canonical; `/api/v1/listings` remains a temporary compatibility alias.

See `docs/MIGRATION.md` before changing shared field names/statuses/routes.

## MongoDB collections

Current/relevant collections:

- `users` — recycler/collector identity, recycler authorization/profile
- `scrapListings` — physical V1 collection containing canonical EWasteLot documents plus compatibility fields
- `offers`
- `priceRecords`
- `transactions`
- `handoverRecords`
- `payments`
- auth refresh-session collection

`safetyGuidance`, notifications and recycler requirements are planned/shared-contract areas, not falsely claimed as complete production features.

## Local prerequisites

- Go 1.23+
- Node.js 20+
- Docker Desktop / MongoDB

## Environment

Backend: copy `backend/.env.example` to `backend/.env`.

```dotenv
PORT=8080
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=kabadiconnect
JWT_ACCESS_SECRET=replace-with-at-least-32-random-characters
JWT_REFRESH_SECRET=replace-with-a-different-at-least-32-random-characters
FRONTEND_ORIGIN=http://localhost:5173
ENVIRONMENT=development
```

Frontend: copy `frontend/.env.example` to `frontend/.env`.

```dotenv
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

Never commit real secrets.

## Start MongoDB

From repository root:

```powershell
docker compose up -d mongo
docker compose ps
```

## Seed development data

The seed command only deletes known KabadiConnect development `seedTag` documents and then inserts fictional e-waste examples/price history. It does not silently delete untagged user data.

```powershell
cd backend
go mod tidy
go run ./cmd/seed
```

Development examples include Mixed PCBs, Copper Cable, Damaged LCD Panels, Lithium-Ion Battery Lot, Electronic Motors, Magnet-Bearing Assemblies, Mixed E-Waste Plastics and CRT Units. Seed authorization/pricing is explicitly demo data.

## Run backend

```powershell
cd backend
go run ./cmd/api
```

Health check:

```text
GET http://localhost:8080/healthz
```

Expected:

```json
{"status":"ok"}
```

## Run frontend

```powershell
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

## API overview

Implemented recycler endpoints:

- auth: register/login/refresh/logout
- `GET/PATCH /api/v1/me`
- `GET /api/v1/lots`
- `GET /api/v1/lots/{id}`
- `POST /api/v1/lots/{id}/offers`
- `GET /api/v1/offers`
- `PATCH /api/v1/offers/{id}/cancel`
- `GET /api/v1/prices`
- `GET /api/v1/prices/history`
- `GET /api/v1/transactions`
- `GET /api/v1/transactions/{id}`
- `GET /api/v1/handovers`
- `GET /api/v1/handovers/{id}`
- `POST /api/v1/handovers/{id}/confirm`
- `GET /api/v1/payments`

Collector-side creation/decision endpoints are **planned contracts only** and are documented in `docs/SHARED_API_CONTRACT.md`.

## Authorization behavior

Recycler participation is not merely decorative. Offer creation requires a valid transaction-capable authorization state and material compatibility.

- production registration → `pending_verification`
- local `ENVIRONMENT=development` registration → `demo_verified` with `isDemo=true`
- `demo_verified` is explicitly not a government authorization
- recycler cannot self-promote authorization status through profile PATCH

## Security foundations

- bcrypt password hashing
- access token kept in application memory, not localStorage
- rotating HTTP-only refresh sessions
- recycler RBAC
- auth rate limiting
- request body size cap + unknown JSON field rejection
- CORS origin configuration
- server-derived recycler/collector ownership
- server-calculated offer totals
- authorization/material/lot availability checks
- recycler-owned offer cancellation
- recycler-owned transaction/payment/handover reads
- recycler-owned handover confirmation
- no password hashes serialized
- no payment credentials stored

## UI implemented

- Recycler login/registration
- authorization-aware Dashboard
- E-Waste Lots search/filter/sort
- e-waste lot cards/details
- hazardous-material formal-handling guidance
- compatibility-aware Send Offer flow
- My Offers
- Price Board using persisted demo price records
- Transactions / Orders read view
- Handover Records + recycler confirmation action when records exist
- Payments read view
- Recycler Profile edit/view
- responsive sidebar/drawer and empty/loading/error/success states

## Mocked/development-only

- recycler `demo_verified` authorization in development
- seed collector/recycler profiles
- demo price history and basic trend calculation
- e-waste seed lots

These are visibly marked as development/demo data.

## Planned / not implemented

- collector mobile app/UI
- collector onboarding, Hindi/Marathi mobile UX and offline client
- collector lot creation endpoint implementation
- collector offer accept/reject implementation
- accepted-offer → transaction creation service
- automatic handover creation
- recycler material requirements implementation
- collector relationship page implementation
- notification backend
- real authorization registry integration
- live market price feeds
- production maps/GPS verification
- payment gateway execution
- trained AI image classification or price prediction

## Tests and verification commands

Backend:

```powershell
cd backend
go mod tidy
gofmt -w .
go test ./...
go vet ./...
```

Frontend:

```powershell
cd frontend
npm install
npm run lint
npm run build
```

Then verify the health endpoint and the core flow above against local MongoDB.

Do not claim production readiness until dependency-backed tests/build and the end-to-end flow pass in the target development environment.

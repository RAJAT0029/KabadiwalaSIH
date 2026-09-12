# Implementation Status — E-Waste Recycler Refactor

## Preserved

- React + TypeScript + Vite frontend
- Go modular-monolith REST API
- MongoDB repository architecture and Docker development setup
- JWT access + rotating HTTP-only refresh sessions
- recycler RBAC/security middleware
- existing authentication → lot discovery → offer vertical slice
- `/marketplace` internal route and `/api/v1/listings` compatibility aliases
- physical `scrapListings` collection to avoid destructive migration

## Implemented/modified

- canonical `EWasteLot` domain model and official e-waste category/status constants
- canonical `lot`/`collector` API fields with V1 response aliases
- legacy lot field normalization and new dual-written seed records
- recycler authorization/profile model, pickup/service area and accepted-material logic
- development-only `demo_verified` status; production registration remains pending verification
- offer authorization/material compatibility enforcement
- canonical offer weight/rate/total/unit fields with V1 dual-write compatibility
- persisted `priceRecords` model/repository/API with basic historical trend
- transaction dataset/recycler-owned read API
- handover dataset/recycler-owned read + confirmation API
- payment dataset/recycler-owned read API
- future safety-guidance Go model
- deterministic recycler-matching scoring package with authorization/material hard gates (API endpoint remains planned)
- e-waste-only navigation/dashboard/cards/detail/profile/price/transaction/handover/payment UI
- realistic tagged development e-waste seed data
- shared contract, migration, dataset lifecycle, offline sync and future AI data documentation

## Generic waste removed from active product UX/demo

No ordinary PET bottle, newspaper, cardboard, regular household glass or municipal plastic examples remain in active seed/UI categories.

## MongoDB decision

`users`, `scrapListings`, `offers`, `priceRecords`, `transactions`, `handoverRecords`, and `payments` are used/prepared. Recycler authorization remains inside the recycler `users` document for this vertical slice to avoid duplicated identity/auth state. See `docs/DATA_MODEL.md`.

## Implemented API status

Implemented: recycler auth/profile, lots read, offers create/list/cancel, price board/history, transactions read, handovers read/confirm, payments read.

Compatibility: `/api/v1/listings...`, `/api/v1/price-board`.

Contract only/planned: collector lot create/update, collector offer decision, accepted-offer → transaction creation, collector handover confirmation.

## Working target flow

Recycler login/register → Dashboard → E-Waste Lots → lot detail → authorization/material check → Send Offer → persisted pending offer → My Offers.

## Mocked/development

- `demo_verified` recycler authorization
- seeded collector/recycler profiles
- tagged demo price records and lots
- basic rule/statistical price trend

## Planned/data-dependent

- real government authorization source/process
- live price data
- recycler matching endpoint/UI beyond the prepared deterministic scoring package
- requirements/collector relationship modules
- notifications
- production payment integration
- collector offline client
- AI/ML models

## Verification performed in artifact environment

- Go formatting: executed
- `go test ./...`: attempted, blocked before compilation because this environment cannot reach `proxy.golang.org` and the artifact did not include a pre-existing `go.sum`
- `go mod tidy`: attempted, blocked by external DNS/network restriction
- frontend `npm install`: attempted, timed out in the restricted artifact environment before dependencies were installed
- frontend `npm run lint`: attempted after install failure; could not run because local ESLint dependency was unavailable
- frontend `npm run build`: attempted after install failure; dependency resolution failed as expected; a pre-existing `tsconfig.node.json` option incompatibility surfaced and was fixed, then syntax parsing/import-resolution checks passed
- TypeScript/TSX syntax parse: 22 files, 0 syntax errors
- frontend relative-import resolution: passed

No dependency-backed pass is claimed. Run the README verification commands on the development PC, where package downloads are available, and verify the first end-to-end flow before production claims.

## Migration risks

- collector app must use documented canonical fields/statuses rather than physical Mongo collection names
- legacy aliases should not be removed until both clients migrate
- production authorization cannot use development `demo_verified`
- accepted-offer transaction creation must be idempotent and snapshot material/price state

## Recommended next step

Finalize and jointly fixture-test the collector offer-acceptance contract, then implement idempotent accepted-offer → immutable transaction creation in the shared backend. Do not start collector UI or AI/ML in this workstream.

# KabadiConnect E-Waste Data Model and Lifecycle

This document explains the data lifecycle required by the application. MongoDB collections alone are not treated as sufficient evidence of a working dataset.

## 1. Material Dataset — `EWasteLot`

**Generated:** primarily by the future collector app from material selection, description, phone photograph references, approximate weight, condition, source type, collection time/location and flexible material-specific attributes. Development seed data is tagged and explicitly fictional.

**Validated:** shared API contracts constrain supported categories, positive quantities, unit values, GeoJSON shape/order, status enums and material-specific attribute types when collector write endpoints are implemented.

**Stored:** physical V1 collection `scrapListings` is retained non-destructively. New data uses the canonical `EWasteLot` shape; the Go model normalizes legacy development fields.

**Updated:** collector-owned updates are planned. Status transitions will be server-controlled; recycler clients cannot rewrite lot ownership or history.

**Queried:** Recycler Website uses `/api/v1/lots` with category, sub-category, condition, city, quantity, indicative rate, search, sorting and pagination.

**Analyzed:** deterministic material-profile compatibility and distance/rate ranking are suitable for V1. No ML classifier is claimed.

**Consumed:** dashboard matched lots, E-Waste Lots page, lot detail, offer validation and future transaction snapshots.

**Privacy:** recycler views receive collector business/operating context only; unnecessary collector personal information is not exposed.

## 2. Price Dataset — `priceRecords`

**Generated:** future records may come from validated recycler/aggregator quotes and completed transactions. Current seed records use `sourceType=demo_reference` and `isDemo=true`.

**Validated:** category/sub-category, location, timestamp, positive buying/quoted prices and unit.

**Stored:** `priceRecords` with category/sub-category/timestamp and location indexes.

**Updated:** append new records rather than rewriting history.

**Queried:** `/api/v1/prices` returns latest values plus observed recent range and simple direction; `/api/v1/prices/history` returns raw recent records.

**Analyzed:** current implementation compares recent historical observations and labels trend `up`, `down` or `stable`; this is statistical/rule-based, not prediction.

**Consumed:** Recycler Price Board and lot valuation reference architecture.

**Limitations:** demo values are not live Indian market prices.

## 3. Recycler Authorization/Profile Dataset — `users` (recycler role)

For the current vertical slice, recycler identity and authorization profile remain together in `users` to preserve authentication cohesion and avoid duplicated identity records. A future separate `recyclers` projection can be added if operational needs justify it.

**Generated:** recycler registration/profile maintenance plus future authoritative verification workflow.

**Validated:** email/phone uniqueness, recycler role, accepted material list, structured authorization details and pickup/service area. Clients cannot self-promote authorization status through `PATCH /me`.

**Stored:** `users.authorization`, `acceptedMaterials`, `pickup`, facility/address and offered rates.

**Updated:** recycler can maintain reference/profile data; trusted verification process must own production authorization state.

**Queried/analyzed:** offer creation checks active account, authorization status/expiry and material compatibility.

**Consumed:** dashboard authorization banner, lot-detail compatibility, offer authorization and profile page.

**Development behavior:** registrations use `demo_verified` only when backend `ENVIRONMENT=development`; this is clearly marked non-government demo data.

## 4. Transaction Dataset — `transactions`

**Generated:** planned accepted-offer service creates one transaction with immutable lot/material/price snapshots.

**Validated:** ownership, accepted offer state, available quantity, final/quoted price values and legal status transition.

**Stored:** unique transaction reference, lot/collector/recycler IDs, material snapshot, approximate/final quantity, quoted/final rates/totals, locations, payment status and append-only status history.

**Updated:** only server-side business actions append status events. Frontends must not replace historical events.

**Queried:** recycler-owned GET endpoints are implemented.

**Consumed:** Transactions / Orders UI, handover creation and payment linkage.

**Current limitation:** accepted-offer → transaction creation is planned because collector offer acceptance belongs to the shared collector integration boundary.

## 5. Traceability Dataset — `handoverRecords`

**Generated:** planned from an accepted transaction when a pickup/handover is arranged.

**Validated:** authenticated recycler ownership and allowed status state.

**Stored:** unique handover reference, lot/transaction IDs, parties, photographs, weights, time/location, collector/recycler confirmations and append-only status history.

**Updated:** recycler confirmation endpoint only updates records assigned to that recycler and appends an audit event; if collector was already confirmed, the handover can complete.

**Queried:** recycler-owned handover list/detail.

**Consumed:** Handover Records page and traceability evidence.

**Design:** no blockchain. Server-generated reference + ownership rules + immutable history are sufficient for V1.

## 6. Payment Dataset — `payments`

**Generated:** future transaction/payment workflow. Supports `cash` and `digital`; digital is not mandatory.

**Validated:** transaction/party ownership, positive amount, allowed status/method.

**Stored:** payment ID, transaction/recycler/collector IDs, amount/currency, method, status and timestamps. No card/UPI credentials are stored.

**Queried/consumed:** recycler-owned `/api/v1/payments` and Payments page; future collector app can derive an earnings ledger from shared transaction/payment records.

## 7. Collector Dataset — `users` (collector role)

Minimal collector compatibility data currently uses shared user identity records: preferred language (`hi`, `mr`, `en`), general operating location, business display name and timestamps. Detailed collector mobile onboarding/profile UX is outside this repository.

## 8. Safety Guidance Dataset — contract prepared

The Go `safety.Guidance` model supports material category, title, pictorial/audio references, language, hazard level and guidance text for future collector UX. No detailed collector safety interface is built here.

## 9. Future AI/ML Dataset

Potential features are data-dependent only. Candidate training/analysis fields: material images, validated category/sub-category labels, weights, prices, locations, transactions, recycler matches and outcomes. See `FUTURE_AI_DATA.md`. No dataset size, model accuracy, image recognition or production price prediction is claimed.

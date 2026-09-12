# KabadiConnect Shared E-Waste API Contract — Recycler Workstream

This is the integration boundary between the Recycler Website in this repository and the independently built collector/Kabadiwala mobile application. No collector mobile UI is implemented here.

## Ownership

Recycler workstream: recycler auth/profile, E-Waste Lots read UX, recycler offers, recycler transaction/handover/payment visibility, recycler handover confirmation, price board, shared Go API contracts and MongoDB business models.

Collector app workstream: collector onboarding and vernacular/offline mobile UX, phone photography, collector lot creation UI, approximate weight entry, offer accept/reject UI, earnings/safety UI and collector-side handover confirmation.

## Canonical e-waste categories

`CRT`, `LCD Panel`, `PCB`, `Cable`, `Battery`, `Motor`, `Magnet-bearing Assembly`, `Mixed Plastics from EEE`, `Other E-Waste`.

Ordinary PET bottles, newspaper, cardboard, regular household glass and municipal plastic waste are outside scope unless explicitly generated from end-of-life electrical/electronic equipment.

## Canonical lot contract

```json
{
  "id": "ObjectId",
  "referenceId": "KC-EW-...",
  "collectorId": "ObjectId",
  "domain": "e-waste",
  "material": {
    "category": "PCB",
    "subCategory": "Mixed PCBs",
    "description": "Mixed printed circuit boards from dismantled consumer electronics",
    "condition": "Dismantled",
    "sourceType": "consumer electronics",
    "attributes": { "boardType": "mixed consumer electronics" }
  },
  "quantity": { "approximateWeight": 42, "availableWeight": 42, "unit": "kg" },
  "valuation": {
    "estimatedValue": 13020,
    "estimatedPricePerUnit": 310,
    "marketRangeMin": 260,
    "marketRangeMax": 360,
    "estimationSource": "demo_historical_reference"
  },
  "collection": {
    "collectedAt": "ISODate",
    "location": { "type": "Point", "coordinates": [76.8783, 29.9695] },
    "label": "Kurukshetra, Haryana"
  },
  "status": "available",
  "createdAt": "ISODate",
  "updatedAt": "ISODate"
}
```

GeoJSON order is longitude, latitude. `material.attributes` is flexible by material category but backend write endpoints must validate known fields/types rather than accepting raw Mongo query structures.

## Implemented recycler endpoints

| Method | Route | Purpose |
|---|---|---|
| POST | `/api/v1/auth/register` | Recycler registration |
| POST | `/api/v1/auth/login` | Recycler login |
| POST | `/api/v1/auth/refresh` | Rotate refresh session |
| POST | `/api/v1/auth/logout` | Revoke refresh session |
| GET/PATCH | `/api/v1/me` | Recycler profile and editable references/preferences |
| GET | `/api/v1/lots` | Paginated/filterable available e-waste lots |
| GET | `/api/v1/lots/{lotId}` | E-waste lot detail |
| POST | `/api/v1/lots/{lotId}/offers` | Send authorization/material-checked recycler offer |
| GET | `/api/v1/offers` | Recycler-owned offers |
| PATCH | `/api/v1/offers/{offerId}/cancel` | Cancel own pending offer |
| GET | `/api/v1/prices` | Demo/historical price board |
| GET | `/api/v1/prices/history` | Price records/history |
| GET | `/api/v1/transactions` | Recycler-owned transaction records |
| GET | `/api/v1/transactions/{id}` | Recycler-owned transaction detail |
| GET | `/api/v1/handovers` | Recycler-owned handover records |
| GET | `/api/v1/handovers/{id}` | Recycler-owned handover detail |
| POST | `/api/v1/handovers/{id}/confirm` | Recycler confirmation when business state permits |
| GET | `/api/v1/payments` | Recycler-owned payment records |

Compatibility aliases `/api/v1/listings...` and `/api/v1/price-board` remain during V1 migration.

## Offer request

Canonical request:

```json
{
  "approximateWeight": 20,
  "offeredPricePerUnit": 295,
  "pickupAvailable": true,
  "message": "Can arrange pickup after inspection."
}
```

The backend derives recycler/collector IDs, validates recycler authorization/material compatibility, validates positive quantity/rate and lot availability, and calculates `offeredTotal`. Legacy quantity/price field names remain accepted temporarily.

Offer statuses: `pending`, `accepted`, `rejected`, `expired`, `cancelled`.

## Recycler authorization

Authorization statuses: `pending_verification`, `demo_verified`, `authorized`, `expired`, `suspended`, `unknown`.

`demo_verified` is development-only and must always be marked as demo data. Production registration starts pending verification. Offer creation requires `authorized` or `demo_verified` plus a matching accepted material category.

## Collector-side contracts — planned, not implemented here

These contracts require team coordination before implementation:

- `POST /api/v1/lots` — collector-authenticated idempotent lot creation; backend derives collector ownership.
- `PATCH /api/v1/lots/{lotId}` — collector-owned draft/available updates where business rules permit.
- `GET /api/v1/collector/lots/{lotId}/offers` — collector-owned offer list.
- `PATCH /api/v1/collector/offers/{offerId}/decision` — accept/reject with idempotent accepted-offer → transaction creation.
- collector-side handover confirmation endpoint to be finalized jointly.

Collector APIs should support optional client operation IDs/idempotency keys and safe retries for future offline-first Android synchronization.

## Contract-change rule

Before changing a shared field, enum, ownership rule or endpoint: document the change, preserve compatibility where practical, communicate it to the collector-app team, update fixtures/tests, and only then remove old behavior.

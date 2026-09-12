# E-Waste Domain Migration Decision

## Preserved architecture

The React + TypeScript/Vite frontend, Go modular-monolith API, JWT/refresh-session authentication, MongoDB repository pattern, Docker MongoDB development setup, shared HTTP response format, recycler RBAC and existing offer vertical slice are retained.

## Non-destructive schema decision

The physical MongoDB collection `scrapListings` is retained during this development migration. It now stores the canonical `EWasteLot` shape for new seed/platform data, while the Go model can still read the legacy fields used by earlier development documents.

Canonical additions include `referenceId`, `collectorId`, `domain`, `material.subCategory`, `material.description`, `material.sourceType`, `quantity.approximateWeight`, `quantity.availableWeight`, `valuation`, and `collection`.

Legacy compatibility fields include `kabadiwalaId`, `material.subType`, `material.grade`, `quantity.total`, `quantity.available`, `pricing`, flat `location`, and `locationLabel`.

The API emits canonical `lot`/`collector` fields and temporarily emits `listing`/`supplier` aliases for V1 clients. `/api/v1/lots` is canonical; `/api/v1/listings` remains a compatibility alias.

No development data is silently dropped. The seed command deletes only documents carrying known KabadiConnect development `seedTag` values.

## Status migration

Canonical lot statuses are `draft`, `available`, `offer_received`, `matched`, `handover_scheduled`, `handed_over`, `completed`, `cancelled`. Legacy `active` lots remain readable as `available` through normalization.

## Offer migration

New offers dual-write canonical fields (`lotId`, `collectorId`, `approximateWeight`, `offeredPricePerUnit`, `offeredTotal`, `unit`) and legacy fields (`listingId`, `kabadiwalaId`, `quantity`, `offeredPricePerKg`, `totalAmount`). Server-side totals and ownership remain authoritative.

## Future cleanup gate

Legacy fields/routes can be removed only after both recycler and collector clients have migrated, contract fixtures have been updated, existing development data has been migrated/backed up, and the removal is communicated to both teams.

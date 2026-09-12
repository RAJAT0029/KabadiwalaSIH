# Deterministic Recycler Matching — V1 Preparation

The initial matching approach is deterministic, not ML.

Hard gates:
1. recycler account active;
2. authorization status is `authorized` or development-only `demo_verified` and not expired;
3. lot material category/sub-category is accepted by recycler.

Ranking signals for eligible recyclers:
- material compatibility baseline;
- pickup availability;
- shorter distance;
- offered rate relative to lot/reference rate.

`backend/internal/matching.Score` implements this scoring primitive and never marks an unauthorized recycler eligible. A future `GET /api/v1/lots/{lotId}/recyclers` endpoint is planned only after collector/recycler contract ownership and authoritative authorization data are finalized.

Future ML recommendation may replace/rerank this only after validated operational data exists; the deterministic eligibility gates remain authoritative.

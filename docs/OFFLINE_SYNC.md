# Future Collector Offline Synchronization Contract

The collector mobile app is a separate workstream. This repository only documents backend properties that should make future low-connectivity synchronization safe.

Recommended write semantics:

- collector client generates an operation ID/UUID per offline mutation;
- POST/PATCH requests include an idempotency/operation ID once the collector contract is finalized;
- backend stores processed operation IDs with collector ownership and result reference for a bounded retention period;
- retries of the same operation return the original result rather than creating duplicate lots/confirmations;
- server timestamps remain authoritative, while client capture timestamps may be stored separately;
- documents retain `createdAt` and `updatedAt` for synchronization cursors;
- conflict rules must be domain-specific: server status/ownership history cannot be overwritten by stale offline payloads;
- photographs use durable object references and upload retry/resume semantics when object storage is integrated.

Do not add Kafka, distributed consensus or complex synchronization infrastructure for V1. A small idempotency store plus explicit state transitions is sufficient initially.

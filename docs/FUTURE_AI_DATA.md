# Future AI/ML Data Preparation — Planned / Data-Dependent

KabadiConnect currently does not ship trained AI/ML models.

Potential future tasks:

1. material image classification;
2. approximate valuation support;
3. recycler matching recommendation;
4. abnormal transaction-value detection.

Required evidence before model work:

- sufficiently large, representative and permissioned operational dataset;
- verified category/sub-category labels and quality review;
- consistent weight/unit normalization;
- price provenance and timestamp/location context;
- removal/minimization of personally identifying collector data;
- train/validation/test separation that prevents leakage across repeated lots/collectors;
- documented geographic/time coverage and class imbalance;
- benchmark against simple rule/statistical baselines;
- documented failure modes and confidence/abstention behavior.

Possible source data comes from actual platform lot images/labels, price records, completed transactions, matching outcomes and handover results. Development seed data must never be represented as training evidence.

Initial implementation remains manual material selection, rule/historical valuation, deterministic matching and no anomaly model.

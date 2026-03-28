# Architecture Decision Records — ComplianceForge GRC Platform

## ADR-001: Multi-Tenant Architecture with PostgreSQL RLS

**Status:** Accepted  
**Date:** 2026-03-28

**Context:** The platform must serve multiple European organisations with strict data isolation requirements per GDPR Article 32 and ISO 27001 A.5.15.

**Decision:** Use PostgreSQL Row-Level Security (RLS) with a shared database, shared schema approach. Each table has an `organization_id` column, and RLS policies restrict data access to the current tenant set via `set_config('app.current_tenant', ...)`.

**Consequences:**
- All queries are automatically filtered by tenant without application code changes
- Single database simplifies operations and reduces costs
- Cross-tenant queries are impossible even with application bugs
- Slight performance overhead from RLS policy evaluation on every query

---

## ADR-002: Golang as Primary Backend Language

**Status:** Accepted  
**Date:** 2026-03-28

**Context:** Need a performant, type-safe, concurrent backend for enterprise GRC processing.

**Decision:** Go 1.22+ with Chi router, pgx for PostgreSQL, zerolog for structured logging.

**Consequences:**
- Excellent concurrency for handling multiple compliance calculations
- Fast compilation and deployment cycles
- Strong type safety reduces runtime errors
- Large pool of Go developers available in European market

---

## ADR-003: Cross-Framework Control Mapping Schema

**Status:** Accepted  
**Date:** 2026-03-28

**Context:** European enterprises adopt multiple frameworks simultaneously (e.g., ISO 27001 + GDPR + Cyber Essentials). They need to understand overlap to avoid duplicate work.

**Decision:** A dedicated `framework_control_mappings` table with mapping_type (equivalent, partial, related, superset, subset) and mapping_strength (0.00-1.00) allows precise cross-referencing between any two framework controls.

**Consequences:**
- Organisations can see that implementing ISO 27001 A.8.7 (malware) also satisfies Cyber Essentials CE-MP-01 and PCI DSS 5.2.1
- Compliance scores can be projected across frameworks
- The mapping table must be curated by compliance experts for accuracy

---

## ADR-004: GDPR Breach 72-Hour Notification Workflow

**Status:** Accepted  
**Date:** 2026-03-28

**Context:** GDPR Article 33 requires notification to the supervisory authority within 72 hours of becoming aware of a personal data breach.

**Decision:** When an incident is created with `is_data_breach = true`, a PostgreSQL trigger automatically sets `notification_deadline = reported_at + 72 hours`. The background worker monitors approaching deadlines and sends escalation alerts. The API provides `/incidents/{id}/notify-dpa` to record when notification has been submitted.

**Consequences:**
- Organisations never miss the 72-hour window
- Full audit trail of when breach was reported, when DPA was notified, and when data subjects were informed
- NIS2 incident reporting (24-hour early warning) follows the same pattern

---

## ADR-005: Partitioned Audit Log

**Status:** Accepted  
**Date:** 2026-03-28

**Context:** The audit log grows continuously and must be immutable and queryable for regulatory inspections. ISO 27001 A.8.15 and GDPR accountability require comprehensive logging.

**Decision:** The `audit_logs` table is partitioned by month using PostgreSQL range partitioning on `created_at`. New partitions are created automatically. Data is append-only with no UPDATE or DELETE operations.

**Consequences:**
- Efficient querying of recent audit data
- Old partitions can be archived to cold storage
- Partition pruning improves query performance
- Immutability maintained at the application layer

---

## ADR-006: Redis for Caching and Job Queue

**Status:** Accepted  
**Date:** 2026-03-28

**Context:** Compliance score calculations and dashboard aggregations are expensive. Report generation and email notifications should not block API responses.

**Decision:** Redis serves dual purpose: caching layer (compliance scores, dashboard data, framework lists) with configurable TTLs, and job queue (BRPOP-based) for async tasks like email, reports, and score recalculation.

**Consequences:**
- Dashboard loads are sub-second after first calculation
- Framework data (rarely changes) cached for 1 hour
- Compliance scores cached for 2-5 minutes
- Cache invalidation triggered on relevant data changes
- Job queue provides retry logic with exponential backoff and dead letter queue

---

## ADR-007: EU Data Residency by Design

**Status:** Accepted  
**Date:** 2026-03-28

**Context:** European enterprises require their data to remain within the EU/EEA. GDPR Chapter V restricts international data transfers.

**Decision:** Infrastructure designed for EU-region deployment: Kubernetes manifests target EU cloud regions (Frankfurt, Amsterdam, Dublin), PostgreSQL data residency enforced at the infrastructure level, file storage uses EU-region buckets, and the `DATA_RESIDENCY_REGION` configuration flag is available for application-level checks.

**Consequences:**
- Compliant with GDPR transfer restrictions
- Cloud provider selection limited to those with EU regions
- Disaster recovery must also use EU regions

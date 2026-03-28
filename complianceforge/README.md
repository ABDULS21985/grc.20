# ComplianceForge — Enterprise GRC Platform

> **Compliance Management Solution for European Enterprises**
> Supporting 9 regulatory frameworks with 300+ sellable capabilities

## Supported Standards & Frameworks

| Framework | Version | Domain |
|-----------|---------|--------|
| ISO/IEC 27001 | 2022 | Information Security Management |
| UK GDPR | 2021 | Data Protection & Privacy |
| NCSC CAF | 3.2 | Cyber Assessment (NIS-aligned) |
| Cyber Essentials | 3.1 | Basic Cyber Hygiene |
| NIST SP 800-53 | Rev 5 | Security & Privacy Controls |
| NIST CSF | 2.0 | Cybersecurity Framework |
| PCI DSS | v4.0 | Payment Card Security |
| ITIL | 4 | IT Service Management |
| COBIT | 2019 | IT Governance |

## Tech Stack

- **Backend:** Go 1.22+ (Chi router, pgx, zerolog)
- **Database:** PostgreSQL 16+ with Row-Level Security
- **Cache:** Redis 7+
- **Auth:** JWT (OAuth 2.0 / OIDC ready)
- **Architecture:** Multi-tenant, REST API, Clean Architecture

## Quick Start

```bash
# 1. Clone and setup
cp .env.example .env

# 2. Start infrastructure
make docker-up

# 3. Run migrations
make migrate-up

# 4. Seed framework data
make seed

# 5. Run the API server
make run
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Public
- `GET /api/v1/health` — System health check
- `GET /api/v1/ready` — Kubernetes readiness probe

### Compliance Frameworks
- `GET /api/v1/frameworks` — List all frameworks
- `GET /api/v1/frameworks/{id}` — Get framework with domains
- `GET /api/v1/frameworks/{id}/controls` — List controls (paginated)
- `GET /api/v1/frameworks/controls/search?q=` — Full-text search
- `GET /api/v1/compliance/scores` — Compliance scores per framework

### Risk Management
- `GET /api/v1/risks` — List risk register
- `POST /api/v1/risks` — Create new risk
- `GET /api/v1/risks/{id}` — Get risk detail
- `GET /api/v1/risks/heatmap` — Risk heatmap data

### Policies, Audits, Incidents, Vendors
All routes registered and ready for implementation.

## Architecture

```
cmd/api/main.go          → HTTP server entry point
internal/
  config/                → Configuration (Viper)
  database/              → PostgreSQL connection pool + RLS
  middleware/             → Auth (JWT), Logging, Tenant, CORS
  models/                → Domain entities & enums
  repository/            → Data access (PostgreSQL queries)
  service/               → Business logic
  handler/               → HTTP handlers
  router/                → Route registration
sql/
  migrations/            → PostgreSQL schema (5 migrations)
  seeds/                 → Framework & reference data
```

## Database Schema

**35+ tables** across 5 migrations:
1. Extensions, enums, RLS helper functions
2. Organizations, Users, RBAC, Audit log (partitioned)
3. Frameworks, Controls, Cross-mappings, Implementations, Evidence
4. Risk Register, Assessments, Treatments, KRIs
5. Policies, Audits, Incidents, Vendors, Assets

All tables include: UUID PKs, soft deletes, timestamps, JSONB metadata, RLS policies, and full-text search indexes.

## Multi-Tenancy

PostgreSQL Row-Level Security (RLS) ensures complete tenant isolation:
- Every table has `organization_id` column
- RLS policies enforce data access per tenant
- Session-level tenant context via `set_config('app.current_tenant', ...)`
- Enforced in every database transaction

## License

Proprietary — Digibit Global Solutions

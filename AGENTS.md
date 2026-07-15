# AGENTS.md — ecommerce-tenant-service

Multi-tenancy service: source of truth for tenants, and the orchestrator for **self-service
tenant registration**. Consumed by every other service (each wires `tenant-service-api`).
Read the workspace-root `CLAUDE.md` first for cross-repo rules (release-then-bump, hexagonal
layout, `fx` DI, `buf`-generated contracts, tenant conventions). This file covers what is
specific to *this* service.

## What this service does

Two distinct responsibilities live here:

1. **Tenant CRUD** (`internal/application/tenant/`) — plain command/query handlers over a Mongo
   `tenant` collection, emitting `TenantUpdatedEvent` / `TenantDeletedEvent` to Kafka via the
   transactional outbox. This is the read/write model other services project from.
2. **Registration saga** (`internal/application/registration/`) — a persistent, resumable saga
   that provisions a whole tenant from a signup: creates a Logto user, a tenant, assigns a role,
   publishes the tenant event, and triggers catalog seeding. This is the non-obvious core of the
   service — see below.

`cmd/main.go` is the composition root (fx modules only). Note the extra outbound adapters beyond
the usual Mongo/Kafka: `logto` (identity provider) and `k8s` (seeder job launcher).

## The registration saga (read this before touching `registration/`)

`Registration` (`registration.go`) is a persisted aggregate tracking saga progress with boolean
step flags (`TenantSet`, `RoleAssigned`, `EventPublished`, `CatalogSeeded`) and a `Status`
(`provisioning` → `completed`, or `compensating` → `rolled_back`).

- **Entry point** `register.go`: creates the Logto user *first* (synchronously), then persists the
  `Registration`, then attempts the saga **inline** (fast path). If any step fails inline, it
  returns with the registration still in-flight — the worker resumes it. Password is used once and
  discarded (never stored).
- **`processor.go`** runs steps forward (`Process`) and reverses them (`Compensate`). Every step is
  **idempotent**: it checks its flag/field and returns early if already done, so resuming a
  half-finished saga is safe. Each successful step is persisted immediately.
- **`worker.go`** polls every 10s (`FindActionable`) for registrations that are `provisioning` (with
  a due `NextRetryAt`) or `compensating`, and drives them forward or rolls them back. Registered as
  an fx `worker.RunWorker` in `application/module.go`.

Error handling rules — **preserve these when editing**:
- **Permanent errors** (`isPermanentError`: user/slug already exists, invalid data) → `MarkCompensating`,
  which triggers rollback. Transient errors → `ScheduleRetry` (exponential backoff, base 30s, cap 15m).
- **Catalog seeding never compensates.** It is non-critical: on failure it only schedules a retry.
  Do not add rollback for the seed step.
- Compensation deletes in reverse order (Logto user, then tenant) and is itself retryable.
- `completedAt` drives a Mongo TTL index (7 days) that reaps terminal registrations — do not repurpose
  that field.

## Outbound adapters specific to this service

- **`logto/`** — implements `tenant.IdentityProvider` against the Logto Management API over HTTP
  (OAuth2 client-credentials via `security.client-credentials`). Caches role IDs. `CreateUser`
  treats HTTP 409 as an existing-user conflict. Config key `logto.*` (`base-url`, `resource`,
  `client-id`, `client-secret`, `token-url`).
- **`k8s/`** — `SeederJobLauncher` implements `registration.CatalogSeeder` by cloning a **CronJob**
  template's job spec into a one-off Job (per-tenant, `GenerateName: seeder-<slug>-`). Uses
  `rest.InClusterConfig()`, so it only works when running inside the cluster. Config key `k8s.*`
  (`namespace`, `cronjob-name`).

## Ports & tests

Domain ports live in `internal/application/tenant/` (`Repository`, `IdentityProvider`,
`TenantEventFactory`) and `internal/application/registration/` (`Repository`, `CatalogSeeder`).
Implement new adapters against these; provide them through the relevant `fx` module, never
construct them in `main.go`.

There are currently no `*_test.go` files in this repo. The Makefile still exposes the standard
`test` / `test-unit` / `test-integration` (build tag `integration`) / `test-e2e` (tag `e2e`)
targets; follow the workspace testing conventions when adding tests.

## Config

YAML per env under `configs/` (`config.standalone.yaml`), overridable by env/`.env` (`APP_ENV`
selects the file). The standalone config does **not** include `logto` or `k8s` sections — those
are supplied by the deployment (Helm values) since the k8s seeder only runs in-cluster. Auth is
JWT via `security.jwks`; M2M calls to Logto use `security.client-credentials`.

## Storage

Single Mongo database (`tenant`) — this service's own data is **not** tenant-scoped, so it uses
plain repositories, not `NewTenantRepository`. Indexes are created via JSON migrations in
`db/migrations/` (run by the commons persistence module with `WithMigrations()`): unique `slug`
on both `tenant` and `registration`, plus the registration status/retry and TTL indexes.

## Commands

Standard Go-service Makefile (see root `CLAUDE.md`). Most-used here:
```bash
make run                 # go run ./cmd/main.go
make test                # -race + coverage
make lint                # golangci-lint
make check-all           # deps + fmt + lint + test + vuln-check (CI pipeline)
```

## API contract

Protobuf lives in the sibling repo `ecommerce-tenant-service-api` (`proto/tenant/v1/` for RPCs,
`proto/tenant/events/v1/` for Kafka event schemas). Edit `.proto` there and `make generate`;
never hand-edit `gen/`. Connect-RPC handlers here are in `internal/infrastructure/inbound/connect/`.

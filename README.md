# Subscription Service

A REST API for managing sport course subscriptions with flexible plans, voucher discounts, and trial periods.

## Tech Stack

- **Go 1.22** — language
- **Gin** — HTTP framework
- **GORM + PostgreSQL** — persistence
- **Swaggo** — OpenAPI docs
- **Clean Architecture** — domain → usecase → repository → delivery

## Quick Start

```bash
docker-compose up -d   # start postgres
make reset             # drop tables, migrate, and seed fresh data
make run               # start server on :8080
```

Or if the DB is already migrated and you just want to add seed data:

```bash
make seed
```

Swagger UI: http://localhost:8080/swagger/index.html

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/products` | List all sport courses (with plans) | No |
| GET | `/api/v1/products/:id` | Get course by ID (with plans) | No |
| GET | `/api/v1/products/:id/plans` | List plans for a course | No |
| POST | `/api/v1/vouchers/validate` | Validate voucher + preview discount | No |
| POST | `/api/v1/subscriptions` | Buy a subscription | Yes |
| GET | `/api/v1/subscriptions/me` | Get current user's active subscriptions | Yes |
| GET | `/api/v1/subscriptions/:id` | Get subscription by ID | Yes |
| POST | `/api/v1/subscriptions/:id/pause` | Pause active subscription | Yes |
| POST | `/api/v1/subscriptions/:id/unpause` | Resume paused subscription | Yes |
| POST | `/api/v1/subscriptions/:id/cancel` | Cancel subscription | Yes |

## Authentication

Pass any string as a Bearer token. The raw token value is used as the `userID`:

```
Authorization: Bearer user-123
```

No signature verification — this is intentional for the scope of this assessment.

## Example Requests

**List all sport courses**
```bash
curl http://localhost:8080/api/v1/products
```

**List plans for a course**
```bash
curl http://localhost:8080/api/v1/products/1/plans
```

**Buy a subscription (select course + plan)**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer user-123" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "plan_id": 1}'
```

**Buy with a voucher**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer user-123" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "plan_id": 3, "voucher_code": "SAVE10"}'
```

**Buy with a trial period**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer user-123" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 2, "plan_id": 6, "with_trial": true}'
```

**Get current user's active subscriptions**
```bash
curl http://localhost:8080/api/v1/subscriptions/me \
  -H "Authorization: Bearer user-123"
```

**Get subscription by ID**
```bash
curl http://localhost:8080/api/v1/subscriptions/1 \
  -H "Authorization: Bearer user-123"
```

**Pause**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions/1/pause \
  -H "Authorization: Bearer user-123"
```

**Unpause** (extends end date by days paused)
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions/1/unpause \
  -H "Authorization: Bearer user-123"
```

**Cancel**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions/1/cancel \
  -H "Authorization: Bearer user-123"
```

**Validate a voucher**
```bash
curl -X POST http://localhost:8080/api/v1/vouchers/validate \
  -H "Content-Type: application/json" \
  -d '{"code": "FLAT5", "product_id": 1, "plan_id": 1}'
```

## Seeded Data

### Sport Courses & Plans

Each course has 4 plans (Monthly / Quarterly / Semi-Annual / Annual), all at 19% tax.

| Course | Monthly | Quarterly | Semi-Annual | Annual |
|--------|---------|-----------|-------------|--------|
| Yoga | €9.99 | €24.99 | €44.99 | €79.99 |
| Swimming | €14.99 | €34.99 | €59.99 | €99.99 |
| CrossFit | €19.99 | €49.99 | €84.99 | €149.99 |
| Cycling | €12.99 | €29.99 | €54.99 | €94.99 |

Plans are assigned IDs in insertion order: Yoga gets plans 1–4, Swimming 5–8, CrossFit 9–12, Cycling 13–16.

### Vouchers

| Code | Type | Value | Max Uses |
|------|------|-------|----------|
| `SAVE10` | Percent | 10% off | 100 |
| `FLAT5` | Fixed | €5 off | 50 |

Both vouchers are product-agnostic (apply to any course).

## Testing

Unit tests cover the subscription usecase layer using **testify/mock** — no database required.

```bash
go test ./internal/usecase/... -v
```

| Test | What it verifies |
|------|-----------------|
| `TestBuySubscription_Success` | Subscription created with correct status, prices, and dates |
| `TestBuySubscription_AlreadyActive` | Returns error if user already has active subscription for the same product; `Create` never called |
| `TestPause_Success` | Status transitions to `paused`, `PausedAt` is stamped |
| `TestPause_WhenAlreadyPaused` | Returns error, `Save` never called |
| `TestPause_WhenTrialing` | Returns error, `Save` never called |
| `TestUnpause_ExtendsEndDate` | Status back to `active`, `PausedDays` accumulated, `EndDate` extended by exact days paused |
| `TestCancel_Success` | Status transitions to `cancelled` |
| `TestCancel_WhenAlreadyCancelled` | Returns error, `Save` never called |
| `TestCancel_NotFound` | Repo error is propagated correctly |

**Approach:** `SubscriptionRepository` is mocked with `testify/mock` — expectations are set per test and verified with `AssertExpectations`. `ProductRepository`, `PlanRepository`, and `VoucherRepository` use lightweight in-package stubs. No real DB, no Gin, no network.

## Concurrency & Data Integrity

### Race Condition on Voucher Usage Count (TOCTOU)

A voucher has a `MaxUses` limit and a `UsedCount` counter. Without protection, two concurrent requests using the same voucher can both read `UsedCount = 49` (below the limit of 50), both pass validation, and both increment to 50 — resulting in 51 total uses. This is a classic **Time-Of-Check / Time-Of-Use (TOCTOU)** race.

**Fix applied — two-phase check with `SELECT FOR UPDATE`:**

1. **Outside the transaction** — `GetByCode` does a plain read to give a fast early rejection (no lock held while doing product/plan lookups).
2. **Inside the transaction** — `GetByCodeForUpdate` issues `SELECT ... FOR UPDATE`, acquiring a **row-level lock** on the voucher row. Any concurrent transaction that also tries to lock the same row blocks until this one commits or rolls back. The voucher is re-validated after the lock is acquired (the count may have changed since step 1), then `UsedCount` is incremented and the subscription is created — all atomically.

```
concurrent req A ──► SELECT FOR UPDATE (locks row) ──► re-validate ──► UsedCount++ ──► CREATE sub ──► COMMIT
concurrent req B ──►                      (blocked) ──────────────────────────────────────────────────► SELECT FOR UPDATE ──► re-validate → 429
```

### Database Transaction (Atomicity)

The critical write path (increment voucher usage + create subscription) runs inside a single GORM transaction via the `Transactor` abstraction (`WithinTransaction`). Both the `SubscriptionRepository` and `VoucherRepository` passed inside the callback share the **same underlying `*gorm.DB` transaction**, so:

- If the subscription `Create` fails, the voucher `UsedCount` increment is rolled back automatically.
- If the voucher lock or re-validation fails, no subscription row is written.
- The two operations are **always consistent** — there is no state where a voucher is consumed but no subscription exists, or vice versa.

The `Transactor` interface lives in the domain layer, keeping the usecase free of any GORM or Postgres import. The concrete implementation (`postgres.transactor`) wraps `db.Transaction(...)`.

### Other Integrity Measures

- **Idempotency guard** — before any DB write, the usecase checks for an existing `active`, `trialing`, or `paused` subscription for the same user + product. Duplicate subscriptions are rejected at the application layer.
- **Plan–product validation** — the usecase verifies `plan.ProductID == req.ProductID` before accepting a purchase, preventing a client from mixing plans across courses.
- **State machine guards** — `CanPause`, `CanUnpause`, `CanCancel` are enforced on the entity before any mutation, making invalid transitions impossible regardless of call order.

## Design Decisions

- **Clean Architecture layers** — `domain` (entities + interfaces) has zero external dependencies. `usecase` depends only on domain interfaces. `repository` and `delivery` implement those interfaces. This means business logic is fully testable without a database or HTTP framework.

- **Products vs Plans** — a *Product* is a sport course (Yoga, Swimming, etc.); a *Plan* is the billing configuration for that course (duration, price, tax rate). One course has many plans. When buying, the user picks both a course and a plan, and the plan drives all pricing. This separates the "what you're subscribing to" from "how you're billed".

- **Subscription as a state machine** — valid transitions (`Active → Paused`, `Paused → Active`, `Active/Paused/Trialing → Cancelled`) are enforced on the entity itself (`CanPause`, `CanUnpause`, `CanCancel`). The usecase calls these guards before mutating state, so invalid transitions are impossible regardless of which handler triggers them.

- **Idempotency via usecase guard** — before creating a subscription, the usecase checks whether the user already has an `active`, `trialing`, or `paused` subscription for the same product (course) and returns an error. A user can only have one active subscription per course, regardless of which plan they pick.

- **Plan validation in Buy** — the usecase fetches the plan by ID and verifies `plan.ProductID == req.ProductID`. This prevents a client from submitting a plan that belongs to a different course.

- **No Redis** — caching was deliberately omitted. Product and voucher lookups hit Postgres directly; at this scale the query cost is negligible and adding a cache layer would introduce consistency complexity (invalidation on voucher expiry, used-count drift) for no meaningful gain.

## Extensibility Points

What would be added before going to production:

- **Redis** — cache product/plan listings and voucher lookups; rate-limit the buy endpoint per user
- **Proper JWT** — replace the raw-token-as-userID stub with a real JWT middleware that validates signatures and extracts claims from a user service
- **Payment webhooks** — a `POST /webhooks/payment` handler to transition subscriptions from `trialing` → `active` upon successful charge, with idempotency keys to handle duplicate delivery

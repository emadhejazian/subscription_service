# Subscription Service

A REST API for managing product subscriptions with voucher discounts and trial periods.

## Tech Stack

- **Go 1.22** — language
- **Gin** — HTTP framework
- **GORM + PostgreSQL** — persistence
- **Swaggo** — OpenAPI docs
- **Clean Architecture** — domain → usecase → repository → delivery

## Quick Start

```bash
docker-compose up -d   # start postgres
make seed              # insert products and vouchers
make run               # start server on :8080
```

Swagger UI: http://localhost:8080/swagger/index.html

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/products` | List all products | No |
| GET | `/api/v1/products/:id` | Get product by ID | No |
| POST | `/api/v1/vouchers/validate` | Validate voucher + preview discount | No |
| POST | `/api/v1/subscriptions` | Buy a subscription | Yes |
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

**List products**
```bash
curl http://localhost:8080/api/v1/products
```

**Buy a subscription (plain)**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer user-123" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1}'
```

**Buy with a voucher**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer user-123" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 2, "voucher_code": "SAVE10"}'
```

**Buy with a trial period**
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer user-123" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "with_trial": true}'
```

**Get subscription**
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
  -d '{"code": "FLAT5", "product_id": 1}'
```

## Seeded Data

| Product | Duration | Price | Tax |
|---------|----------|-------|-----|
| Monthly Plan | 1 month | €9.99 | 19% |
| Quarterly Plan | 3 months | €24.99 | 19% |
| Semi-Annual Plan | 6 months | €44.99 | 19% |
| Annual Plan | 12 months | €79.99 | 19% |

| Voucher | Type | Value | Max Uses |
|---------|------|-------|----------|
| `SAVE10` | Percent | 10% off | 100 |
| `FLAT5` | Fixed | €5 off | 50 |

## Design Decisions

- **Clean Architecture layers** — `domain` (entities + interfaces) has zero external dependencies. `usecase` depends only on domain interfaces. `repository` and `delivery` implement those interfaces. This means business logic is fully testable without a database or HTTP framework.

- **Subscription as a state machine** — valid transitions (`Active → Paused`, `Paused → Active`, `Active/Paused/Trialing → Cancelled`) are enforced on the entity itself (`CanPause`, `CanUnpause`, `CanCancel`). The usecase calls these guards before mutating state, so invalid transitions are impossible regardless of which handler triggers them.

- **Idempotency via usecase guard** — before creating a subscription, the usecase checks whether the user already has an `active`, `trialing`, or `paused` subscription for the same product and returns an error. This prevents duplicate billing without needing a unique DB constraint on `(user_id, product_id)`.

- **No Redis** — caching was deliberately omitted. Product and voucher lookups hit Postgres directly; at this scale the query cost is negligible and adding a cache layer would introduce consistency complexity (invalidation on voucher expiry, used-count drift) for no meaningful gain.

## Extensibility Points

What would be added before going to production:

- **Redis** — cache product listings and voucher lookups; rate-limit the buy endpoint per user
- **Proper JWT** — replace the raw-token-as-userID stub with a real JWT middleware that validates signatures and extracts claims from a user service
- **Payment webhooks** — a `POST /webhooks/payment` handler to transition subscriptions from `trialing` → `active` upon successful charge, with idempotency keys to handle duplicate delivery

# Community Event System (RPG)

Requires `ENABLE_RPG_ENDPOINTS=true`.

## Database setup

```bash
mysql -u ... -p ... your_database < api/sql/types/community_events.sql
```

## Cloudflare caching

The only edge-cached player endpoint is:

`GET /api/{API_VERSION}/community-events/current/public`

Origin already sets `Cache-Control: public, max-age=…` (≈5–30s depending on progress / “no event”).
Contribute / claim / player / developer routes use `no-store`.

**Important:** Cloudflare orange-cloud proxy alone does **not** cache JSON API responses.
You must add a Cache Rule (or equivalent) or every request still hits origin.

### 1. Cache Rule (Cloudflare dashboard)

Caching → Cache Rules → Create rule, roughly:

| Field | Value |
|-------|--------|
| When | Hostname equals your API host (e.g. `rc.gunraidersapi.com`) **AND** URI Path equals `/api/<version>/community-events/current/public` |
| Then | Eligible for cache / Cache eligibility: eligible |
| Edge TTL | Use cache-control header / Respect origin |
| Browser TTL | Respect origin (or same) |

Keep the match **exact** to that one path so auth’d CE routes are never cached.

Optional: under the rule, ignore query string (this GET has none) and do not require cache keys on cookies/Authorization (Unity public fetch sends no auth).

### 2. Env vars on the API process (purge)

When an event completes/fails or admin mutates state, the API calls Cloudflare purge for that public URL. Set:

| Var | Example |
|-----|---------|
| `CLOUDFLARE_ZONE_ID` | Zone ID for `gunraidersapi.com` |
| `CLOUDFLARE_API_TOKEN` | Token with **Zone → Cache Purge → Purge** |
| `CLOUDFLARE_PUBLIC_API_BASE` or `PUBLIC_API_BASE` | `https://rc.gunraidersapi.com` (no trailing path) |
| `API_VERSION` | Must match the path clients use (e.g. `v2`) |

Purge URL built as:

`{PUBLIC_API_BASE}/api/{API_VERSION}/community-events/current/public`

Without zone/token (or base/version), purge is a **no-op**; `Cache-Control` headers are still set.

Admin can also force purge: `POST /admin/community-events/:eventId/purge-cache`.

### 3. Verify

```bash
curl -sI "https://rc.gunraidersapi.com/api/v2/community-events/current/public"
```

Look for:

- `cache-control: public, max-age=…`
- Second request within TTL: `cf-cache-status: HIT` (first may be `MISS` / `DYNAMIC` until the rule is live)

If you always see `DYNAMIC` / `BYPASS`, the Cache Rule is missing or too narrow/wide.

## Player API

| Method | Path | Cache |
|--------|------|-------|
| GET | `/community-events/current/public` | `public, max-age=…` |
| GET | `/community-events/:eventId/player` | `no-store` |
| POST | `/community-events/:eventId/contributions` | `no-store` |
| POST | `/community-events/:eventId/claim` | `no-store` |

## Admin API

`/admin/community-events/*` (auth + admin privilege)

## Developer API

`/developer/community-events/*` (auth + admin privilege), uncached.

## Tests

```bash
cd api/gofiber
go test ./domain/community_event/ -v
```

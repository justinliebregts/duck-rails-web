# Duck Rails gameplay coins

This ports Duckz's wallet/store ledger and admin controls into Duck Rails. Do not run migrations or publish images against Duckz or Classic.

API base: `https://duckrails.gunraidersapi.com/api/v1`. All paths below are relative to that base. There is no `/duckz` prefix: for example, `https://duckrails.gunraidersapi.com/api/v1/wallet` and `https://duckrails.gunraidersapi.com/api/v1/store/buy`. Deploy the matching API, admin and Unity client together. Internal Duckz package/table names are implementation reuse, not routes to the Duckz server.

1. Apply `api/sql/create_duckz_economy.sql` to **duckrails_tencent_global**. It is additive and creates economy receipts, entitlements/loadouts and a one-time welcome-grant receipt. Existing `transactions` and `players` are reused.
2. Build/push/deploy the API and admin following `AGENTS.md`, with unique commit-based tags. Registry push and TKE rollout are separate operations; use the existing Duck Rails workloads.
3. Confirm a fresh player's first authenticated `GET /wallet` returns 100 coins, then a repeat call does not add more. An existing player's first Rails wallet initialization also grants 100 once.
4. Check `/store` contains exactly seven non-durable standard items. Prices and bar value live in `api/gofiber/config/json/duckz_economy.json`; the checked-in initial values match Duckz. No coin packs or cosmetic IAP products are added.
5. Test signed `POST /store/buy` and `/coins/deposit`: repeated request IDs must not charge/credit twice; another account cannot reuse an existing request ID. Prices and bar conversion are calculated on the server. Concurrent balance changes lock the player row.
6. Open an admin player's detail page: inspect Duck Coins, grant/remove a test amount, verify the resulting balance and the transaction list. Non-admin calls must be rejected. Admin credentials remain outside the game build.
7. Verify Unity's shop deposits pay all connected recipients and the safe pays only escaped participants. Wallet API tests alone do not prove Photon eligibility, physical item delivery or headset interaction.

The grant and its receipt commit in one transaction under the same player lock as purchases and admin adjustments. A zero balance, reinstall, new run or soft-deleted transaction does not reset the welcome receipt.

Duckz's existing trust model is retained: game sessions submit signed, bounded, idempotent gold claims. The API does not independently simulate Photon runs. Deterministic receipt IDs prevent accidental duplicate claims; they are not cryptographic proof of an escape.

Local validation and production rollout evidence must be recorded separately. No migration, image push or TKE rollout has been performed by adding these files.

URL follow-up validation: `go test -p 2 ./router ./router/duckz` passes. Route tests cover all ten unprefixed economy endpoints requiring authentication, the removed wallet path returning 404, and public routes retaining their original access. The admin production build and Unity script compilation also pass; these checks do not establish deployment.

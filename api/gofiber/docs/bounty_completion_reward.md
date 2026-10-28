# Bounty completion reward (RPG only)

Requires `ENABLE_RPG_ENDPOINTS=true`. Gun Raiders `/bounty/get` and `/bounty/update` contracts are unchanged.

## Database setup

Run before enabling RPG bounty completion on an environment:

```bash
mysql -u ... -p ... your_database < api/sql/types/player_bounty_completion_reward.sql
```

Table: `player_bounty_completion_reward`

## API (RPG client only)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/rpg/bounty/update` | Same body as `/bounty/update`; response adds `completion_rewards` |
| GET | `/rpg/bounty/completion-rewards` | Optional poll (e.g. on bounty screen open) |
| POST | `/rpg/bounty/completion-reward/claim` | Body: `{ "reward_id": 123 }` → RPG loot |

Standard bounties: existing `GET /bounty/get`, `POST /bounty/update` (unchanged).

## Client flow

1. `GET /bounty/get` for the list.
2. `POST /rpg/bounty/update` (not `/bounty/update`) when syncing progress — read `completion_rewards` from the response.
3. If `daily.claimable` or `weekly.claimable`, show chest UI and `POST /rpg/bounty/completion-reward/claim`.
4. Optionally `GET /rpg/bounty/completion-rewards` when opening the bounty UI without an update.

## Claim by `reward_id` (period rollover)

- **Earn** ties a row to a bounty set (`set_start`); the client receives `reward_id`.
- **Claim** accepts that `reward_id` until `claimed_at` is set — including after UTC midnight (e.g. earn 23:59, claim 00:05).
- **GET** `completion-rewards` only surfaces unclaimed chests for the **current** daily/weekly window (for UI hints). If GET shows `null` after rollover but the client still has `reward_id` from the last update, **claim with that id still works**.

## Tests

```bash
cd api/gofiber
go test ./model/ -run BountyPeriod -v
go test ./service/bounty/ -v
go test ./repository/ -run BountyCompletion -v
```

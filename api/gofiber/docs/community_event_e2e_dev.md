# E2E: Developer-only community event

Manual / API checklist for a DevelopersOnly event (does not affect public players).

## Prerequisites

1. Apply `api/sql/types/community_events.sql`
2. `ENABLE_RPG_ENDPOINTS=true`
3. Admin session with `admin_privileges`
4. Optional CF env vars (purge no-ops if unset)

## Steps

1. **Create draft**
   `POST /admin/community-events/`
   ```json
   {
     "visibility": "DevelopersOnly",
     "title": "Dev Kill Mobs",
     "description": "E2E",
     "objective_type": "MobKilled",
     "target_id": "",
     "target_amount": 20,
     "reward_unique_equipment_ids": [<valid_ueq_id_1>, <valid_ueq_id_2>],
     "start_at": "<now-1m UTC>",
     "end_at": "<now+2h UTC>"
   }
   ```

2. **Publish**
   `PUT /admin/community-events/{id}` with `{ "publish": true }`

3. **Confirm not on public cache endpoint**
   `GET /community-events/current/public` → `data: null` (or a different Public event)

4. **Developer current**
   `GET /developer/community-events/current` → status `Active`

5. **Contribute**
   `POST /developer/community-events/{id}/contributions`
   ```json
   { "batch_id": "e2e-1", "objective_type": "MobKilled", "target_id": "", "amount": 15 }
   ```
   Repeat with new `batch_id` until Completed (or force complete).

6. **Idempotency**
   Resubmit same `batch_id` → `duplicate: true`, same accepted amount.

7. **Claim**
   `POST /community-events/{id}/claim` as contributing developer → `Granted` or `AlreadyOwned`

8. **Unity**
   - Add `CommunityEventBootstrap` + island UI in CAMP
   - Sign in as developer admin account
   - Enter CAMP → island appears for DevelopersOnly event
   - Kill mobs → pending queue flushes → progress updates
   - On complete → claim via chest ButtonPush

## Automated package tests

```bash
cd api/gofiber
go test -mod=mod ./domain/community_event/ -count=1
```

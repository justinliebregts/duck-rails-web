# Diagnostic report deployment

The Unity client now queues diagnostic reports and uploads through the normal authenticated session. Deploy this API change before expecting reports to appear centrally.

1. Run [the additive migration](api/sql/diagnostic_reports.sql) against `duckrails_tencent_global`.
2. Build/push the updated Duck Rails API image with a new tag and update the existing API workload in Tencent namespace `duckrails`. The user's deployment workflow is the Tencent web console.
3. Configure the existing `ADMIN_API_TOKEN` secret/environment for the API and a matching `ADMIN_API_TOKEN` in the Unity repo root's ignored `.env.local`, or set `DUCK_RAILS_ADMIN_TOKEN` in the local tooling environment. No admin credential belongs in the game build.
4. In Unity, open **Duck Rails > Test Center**, refresh the backend inbox, then submit a manual report from a signed-in development client. Verify upload, retry deduplication and resolution with evidence.

API routes (under `/api/v1`):

- `POST /diagnostics/report`: authenticated game session; 1 MiB maximum, 20 requests/minute per IP; GUID ID is idempotent and bound to the reporter. Existing records cannot be overwritten or reopened by a retry.
- `GET /diagnostics/admin/reports?status=open&offset=0`: admin only; at most 100 summaries plus total. Status accepts `open`, `resolved` or `all`.
- `GET /diagnostics/admin/reports/:id`: admin-only evidence download.
- `POST /diagnostics/admin/reports/:id/status`: admin-only `status` and `resolution`; resolving requires a nonempty verification note.

Evidence is separate from the inherited player moderation reports. The game only collects diagnostic state, actor numbers, room/run IDs and sanitized error locations; it does not record voice, credentials or account identities in the payload. The server records the authenticated reporter privately for ownership/deduplication and never returns it in inbox JSON. Treat report payloads as untrusted data.

`go test -p 2 ./router` passes, including new report validation, authorization, duplicate/owner-collision persistence and resolution-note tests. Pre-existing broader middleware/repository tests require local Redis and were unavailable. The migration and live endpoints have not yet been applied/verified in production.

Client controls, evidence and CLI instructions are in the Unity repository's `docs/testing-tools.md`.

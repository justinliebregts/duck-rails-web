# Duck Rails web project guidance

- This checkout is Duck Rails. Do not publish to Duckz or Gun Raiders Classic repositories or namespaces; some inherited filenames still say GRC.
- Read `DUCKRAILS-SETUP.md` for game-specific settings and `DIAGNOSTICS-SETUP.md` for report migration requirements.
- Keep credentials in ignored local configuration or Tencent Secrets. Never print or commit credentials.

## Build and push to Tencent Silicon Valley (TCR)

1. Run relevant checks (API router changes: `go test -p 2 ./router` from `api/gofiber`). Review and commit intended web changes, then push Git. Image tags use the current short commit hash; do not build an unidentified dirty checkout.
2. Docker Desktop's Linux engine must be running. Registry credentials come from ignored `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat`; use its template when configuring a new machine.
3. The root `Push GRC SV.bat.lnk` points to `api/gofiber/cmd/tencent/tcr/pushToSilliconValley.bat`. Run that batch file with its directory as the working directory, e.g. `cmd /c pushToSilliconValley.bat`.
4. The script builds `api/gofiber/Dockerfile` for `linux/amd64` and pushes `<tencentBase>/<tcrNamespace>/duckrails-api-golang:api-<git-short-hash>`. Verify Docker's successful push digest; a started build is not a completed push.
5. This uploads to TCR only. It does not roll out TKE. Justin uses Tencent's web console: select the existing Duck Rails API workload in namespace `duckrails`, record the previous image for rollback, select the new image tag, and apply the rolling update. Verify the actual workload name before changing it and confirm replacement pods are Ready and the API responds.
6. Apply required additive SQL migrations to `duckrails_tencent_global` before the API rollout. For diagnostics, use `api/sql/diagnostic_reports.sql`. Configure `ADMIN_API_TOKEN` in the API's Tencent Secret/environment; the Unity inbox's matching local credential lives in the Unity repo root `.env.local`, not this repo.

- The admin website has a separate script, `admin/cmd/tencent/tcr/pushToSV-GRC.bat`; do not run it for API-only changes.
- Report Git push, TCR image push, database migration, and TKE rollout separately; do not claim a production rollout from a registry push alone.

# Gun Raiders Classic API - TKE Deploy

The click-by-click console order is [DEPLOY.md](../../../../../DEPLOY.md). This file only covers rebuilding an image and rolling a workload that already exists.

## 1. Build and push to TCR
Copy `sv/sharedEnv.bat.template` to `sv/sharedEnv.bat` and fill the Silicon Valley TCR login. Do not commit that file.

Run `pushToSilliconValley.bat` from this folder. It builds the API for `linux/amd64` from the current commit and pushes:

`uswccr.ccs.tencentyun.com/gr-tcr/grc-api-golang:api-<git-short-hash>`

Commit first. The tag is the commit hash, so uncommitted changes are built but not identifiable.

The admin panel is pushed separately by `admin/cmd/tencent/tcr/pushToSV-GRC.bat` to `gr-tcr/grc-api-admin:adm-grc-<git-short-hash>`.

## 2. Update the TKE workload
1. Tencent Cloud Console > **TKE** > Clusters > select the existing cluster.
2. **Workloads** > **Deployments**, namespace `grc`.
3. Edit `v1-api-deployment` and set the container image to the tag pushed in step 1.
4. Apply to trigger a rolling update.

Current images bake public settings from `api/gofiber/config/build.env`. Use one Opaque application Secret, `game-api-env`, for the credentials described in `DEPLOY-CONFIG.md`. Only `DB_NAME` is used; there is no reader connection or Elasticsearch. JWT keys are optional. Use the Classic Meta app ID and secret. Existing workload variables override build defaults, so remove obsolete imports and overrides when migrating.

## 3. Verify
- Pods reach Running/Ready in the `grc` namespace.
- `GET https://grc.gunraidersapi.com/api/v1/info/all` returns 200.
- Photon custom auth URL is `https://<classic-api-host>/api/photon/authenticate`.

## Cluster maintenance
### Node Pool & Version Updates
Every time the TKE version is updated, the node pools and individual nodes must be updated manually.

Update workflow:
1. Create the new node pool in the same availability zone as the existing one. Persistent volumes are zone-locked.
2. Drain the old nodes so pods move to the new node pool.
3. Delete the old node pool after the pods are running on the new nodes.

> Duck Rails: start with [DUCKRAILS-SETUP.md](DUCKRAILS-SETUP.md). Classic names below are inherited examples; use the Duck Rails build settings and namespace.

# Minimal production configuration

This version uses one MySQL account and connection pool, Redis, and Meta/Photon authentication. It has no Elasticsearch dependency. Attestation history endpoints return HTTP 501; live Meta verification, device bans, and remote ban-status checks remain available.

## 1. Set public values once, before building

Edit `api/gofiber/config/build.env`. This file is copied into the image as `/app/.env`; developer `.env` files are excluded from the Docker build context.

- `DB_HOST`: private MySQL address including port, for example `10.0.0.12:3306`.
- `DB_NAME`: this game's database (Classic defaults to `grc_tencent_global`). Import the schema separately using a database administrator.
- `DB_ACCOUNT`: this game's single database account (Classic defaults to `grcuser`).
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_USERNAME`: the Redis endpoint and account. Leave username empty if the instance uses password-only authentication.
- `REDIS_NAMESPACE`: unique game/environment name such as `grc` or `newgame-prod`. Use lowercase letters, digits, hyphens, or underscores, starting with a letter or digit. Keep `grc` for existing Classic sessions.
- `OCULUS_APPID`: this game's Meta application ID, from its developer dashboard.
- `K8S_NAMESPACE`: this game's namespace.
- `APP_VERSION`: the version expected by the game. Other platform version keys can be added if those platforms are used.

The other defaults are already filled in. Docker sets `DEPLOY_ENV=production`. No read-only database values, Elasticsearch values, or JWT keys are required. Existing session authentication uses Redis tokens. Optional legacy JWT support still works if valid keys are explicitly supplied.

Build using the existing image push script. Changing build settings requires rebuilding the image. Environment variables can still override build settings when needed. An explicitly empty environment variable also overrides the build value.

Do not deploy the unedited `REPLACE_*` values: startup reports their key names and exits before connecting to services.

## 2. Obtain the four credentials

Copy `runtime.env.example` to `runtime.env` beside it. The resulting root-level file is ignored by Git and is outside the API Docker build context (`api/gofiber`). Enter unquoted values, one per line; do not paste credentials into chat.

### DB_PASSWORD

In TencentDB for MySQL, open the target instance and its Account Management page. Create the account named by `DB_ACCOUNT`, choose a password, and grant access to this game's database. Use a host restriction that permits the TKE pod network; do not bind the account to one temporary pod IP. Give it the database privileges needed by the application, not instance administrator privileges. Use this password as `DB_PASSWORD`. Do not create a second reader account.

For an existing account, use its current password from your password manager. Creating a random value in the Secret alone does not change the database password. Coordinate any reset with other deployments using that account.

Reference: [TencentDB account management](https://indonesia.tencentcloud.com/document/product/236/11274).

### REDIS_PASSWORD

Use the password for the Redis instance/account named in the build settings. Obtain it from the existing deployment's credential store or password manager. If provisioning a new Redis instance/account, save the password you set there. Only leave this value empty if that Redis instance deliberately requires no authentication. Do not reset a shared Redis password just to configure a new game.

### OCULUS_APPSECRET

Open the Meta Horizon developer dashboard, select this game's application, and open its API section. Copy its App Secret. Its App ID must match `OCULUS_APPID` in the build settings. This is supplied by Meta; do not generate a random replacement and do not reuse another game's app secret.

Reference: [Meta platform setup](https://developers.meta.com/horizon/documentation/android-apps/ps-get-started/).

### PHOTON_AUTH_PHASE

Generate a random phrase for a new game. In PowerShell:

```powershell
$phraseBytes = New-Object byte[] 32
$phraseGenerator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$phraseGenerator.GetBytes($phraseBytes)
$phraseGenerator.Dispose()
[BitConverter]::ToString($phraseBytes).Replace('-', '').ToLowerInvariant()
```

Copy the resulting 64-character value into `PHOTON_AUTH_PHASE` and the Photon Custom Server dashboard query parameter named `secret`. Do not put it in Unity. Preserve it for updates; rotation requires updating both the backend Secret and Photon dashboard. If Photon custom authentication is intentionally unused, set `ENABLE_PHOTON_AUTH=0` in the build settings and omit this Secret key.

## 3. Create one application Secret in TKE

In the correct TKE cluster and namespace, open Configuration Management > Secret, create an **Opaque** Secret named `game-api-env`, and enter the four core key/value pairs above, plus `ADMIN_API_TOKEN` when using the admin website and optionally `META_EDITOR_USER_IDS` for Editor access. In the API Deployment, import all environment variables from this Secret. Do not attach an application ConfigMap.

Alternatively, from the repository root with kubectl already connected to the intended cluster:

```powershell
kubectl config current-context
kubectl -n grc create secret generic game-api-env --from-env-file=runtime.env --dry-run=client -o yaml | kubectl -n grc apply -f -
```

Replace `grc` with the game's namespace. The API container configuration is:

```yaml
envFrom:
  - secretRef:
      name: game-api-env
```

Remove old ConfigMap imports and individual environment overrides during migration, or those values may override the new image defaults. Keep the separate registry image-pull Secret: it authenticates image downloads and is not application configuration. The existing certificate, service, ingress, and load balancer setup still applies.

After changing Secret values, roll out the API Deployment again. For the existing Classic deployment:

```powershell
kubectl -n grc rollout restart deployment/v1-api-deployment
kubectl -n grc rollout status deployment/v1-api-deployment
```

Verify readiness, player login, a database write/read, and Photon authentication before retiring old configuration objects or database accounts. No cloud changes are made by editing these repository files.

Reference: [Kubernetes Secret environment variables and restarts](https://kubernetes.io/docs/tasks/inject-data-application/distribute-credentials-secure/).

## Reusing this backend for another game

Ban status is private: clients can call `GET /api/v1/serviceban/get` with their valid `Authorization: Bearer <access-token>` header. The server obtains the player identity from that session; no player ID parameter is needed. The response's `data` contains `mic`, `game`, `now`, and `end_date`. The public `/ban/all` and `/serviceban/all` routes have been removed, and `HASH_BANALL` is no longer a setting. Administrator-only ban management remains available. The current Classic Unity scripts have no public ban-list request to migrate; any future in-game ban-status display should use the individual endpoint. Client-side checks do not replace backend ban enforcement.

This profile currently describes Classic. Each game/environment sharing Redis must have its own `REDIS_NAMESPACE` in the build settings. Classic keeps `grc`, preserving its existing session, online-player, display-name, and nonce keys. Response-cache entries now also carry the namespace; older unprefixed response-cache entries are ignored. Changing a deployed game's namespace makes its previous sessions/cache inaccessible to the new pods, so keep the namespace stable. This prevents accidental key collisions; it is not Redis ACL-based access control. Do not use FLUSHDB/FLUSHALL on a shared Redis instance. The admin frontend's API URL is also a separate build setting in `admin/.env.production`.

Optional Cloudflare IP-ban automation and internal routing integrations require their own settings if enabled; they are not part of this minimal profile. Do not carry over development authentication bypass phrases into production.

See [ADMIN-SETUP.md](ADMIN-SETUP.md) for the working static-token admin configuration.

## Additional runtime settings

- `ADMIN_API_TOKEN`: a separate admin credential. It must exactly match the static token supplied by `admin/src/stores/auth.js` (or the current browser `acc_token` cookie). It is not a Photon phrase, Meta secret, or player token. For a new game, generate a new value using the generator above, change the frontend fallback token, rebuild the admin image, and put the matching value in the API Secret. The frontend currently stores this value in source/bundled JavaScript; see ADMIN-SETUP.md for the implications and rotation procedure.
- `META_EDITOR_USER_IDS`: optional comma-separated **organization-scoped** Meta IDs. Use the Editor's `[SignIn] Editor Meta organization-scoped user ID` log. These IDs are not secrets, but putting them in `game-api-env` keeps runtime configuration in one place. Empty means Editor login disabled. This grants Editor access, not administrator privileges.
- `API_VERSION`: optionally keep a Custom environment variable such as `v1` on the API workload to override the baked route prefix. Update client/admin URLs when changing it; the version does not create new handlers by itself.

The admin container needs no custom runtime variables and must not import `game-api-env`. Its API URLs and Apache password hashes are built into the image. Environment changes to the API Secret require a rollout; changes to baked settings require a new image.

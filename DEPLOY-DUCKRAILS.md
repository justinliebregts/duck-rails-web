# Deploy Duck Rails on the existing Tencent stack

Follow this guide in order for the first Duck Rails deployment. All game-specific names below are filled in for Duck Rails. This copy includes the Duck Rails build values and credential instructions. [runtime.env.example](runtime.env.example) is the blank credential template; other inherited guides may still show Classic examples. Public settings are baked into the API image, and one application Secret supplies credentials. The registry pull Secret and ingress certificate Secret are separate infrastructure objects.

Duck Rails runs on the same Tencent account as Duckz. It gets its own database name, its own TKE namespace, its own pods, and its own public load balancer. It reuses the Silicon Valley registry, the Redis host, and the TKE cluster `rc-tke`.

Do the sections in order. The workload form can only attach a Secret that already exists in the same namespace. Duckz objects live in namespace `duckz` and will not appear while the namespace selector says `duckrails`.

The cluster from the Duckz secret page is West US (Silicon Valley), cluster `rc-tke` (`cls-1gdfwglk`).

For an existing deployment, use the update checklist at the end. Reuse existing resources instead of creating duplicates. This guide records the intended Duck Rails configuration, not an already-running Duck Rails deployment; verify live resource names and controller settings in Tencent before applying changes.

Health probes: `/live` checks that the API is serving requests; `/ready` checks MySQL and Redis. The old example failure, unrelated public DNS check, and arbitrary goroutine limit have been removed. A dependency outage makes the pod unready without forcing a process restart.


## Duck Rails values — use these throughout

This is the Duck Rails-specific copy of DEPLOY.md. The original remains unchanged. Work from `X:\Work\duck-rails-web`. These values match the copied repository's current build settings and push scripts. Cloud resources have not been verified or provisioned by this work; confirm shared instance addresses and the cluster before creating objects.

```text
TKE region: West US (Silicon Valley)
TKE cluster: rc-tke
TKE cluster ID: cls-1gdfwglk
TKE namespace: duckrails
Ingress name: duckrails-lb
Load balancer name: duckrails-lb
API Deployment and Service: v1-api-deployment
Admin Deployment and Service: admin-deploy
Application Secret: game-api-env
Registry pull Secret: rc-tcr-key
Database: duckrails_tencent_global
Database account: duckrailsuser
Redis namespace: duckrails
DNS record name: duckrails
Public hostname: duckrails.gunraidersapi.com
API base URL: https://duckrails.gunraidersapi.com/api/v1/
Admin URL: https://duckrails.gunraidersapi.com/admin/
Photon callback: https://duckrails.gunraidersapi.com/api/photon/authenticate
Certificate display name: cloudflare-duckrails.gunraidersapi.com
Registry host: uswccr.ccs.tencentyun.com
Shared registry namespace: gr-tcr
API image repository: duckrails-api-golang
Admin image repository: duckrails-api-admin
Meta app ID: 28912202281742343
Photon Realtime app ID: 36bef5d2-63ad-4567-948e-52e966d7f2e9
Photon Voice app ID: b7c80e26-6a61-45ca-97d9-478af736a6d3
```

`duckrails` is the TKE namespace. Keep `gr-tcr` as the existing shared **registry** namespace; these are different objects. Database/account and image names above are the prepared provisioning names. Never select another game's database, namespace or balancer for a Duck Rails write.

Reuse the existing Tencent account, registry, TKE cluster, MySQL instance, Redis instance and Cloudflare zone. Create the separate per-game resources below. The push scripts already target Duck Rails repositories. Their filenames `pushToSV-GRC.bat` and the root `Push GRC` shortcuts are inherited names; the shortcuts were retargeted to this copy.

### Values you must supply

Only these cannot be prefilled from the repository:

- Duck Rails database password; use the same value when creating `duckrailsuser` and setting `DB_PASSWORD`.
- Existing shared Redis password and registry login credentials from your credential store.
- The Meta App Secret belonging specifically to app `28912202281742343`.
- A new Photon shared phrase, used identically in the API Secret and both Photon dashboards.
- A separate admin API token and Apache username/password if deploying the admin panel.
- Your organization-scoped Meta user ID(s) for Editor login.
- The image git hash printed by each push, certificate ID/TLS Secret name returned by Tencent, and the new load balancer address.

Do not paste placeholder text as a credential. The ignored `runtime.env`, local `.env`, and admin password file may have been copied from Classic; they were not replaced by the sign-in changes. Review them before use.

### Public build settings

`api/gofiber/config/build.env` is already configured as follows. Verify the two shared private endpoints against your existing services; retain the remaining Duck Rails values:

```dotenv
API_VERSION=v1
APP_VERSION=0.2.7
DB_PROTOCOL=tcp
DB_HOST=10.199.1.2:3306
DB_NAME=duckrails_tencent_global
DB_ACCOUNT=duckrailsuser
REDIS_HOST=10.199.1.33
REDIS_PORT=6379
REDIS_USERNAME=
OCULUS_APPID=28912202281742343
ENABLE_PHOTON_AUTH=1
ENABLE_OCULUS_ATTESTATION=1
ENABLE_REQUEST_ID=true
ENABLE_BUILDHASH_CHECK=0
K8S_NAMESPACE=duckrails
REDIS_NAMESPACE=duckrails
```

`admin/.env.production` is already configured as follows:

```dotenv
NODE_ENV=production
VUE_APP_API_URL=https://duckrails.gunraidersapi.com/api/v1
VUE_APP_DEFAULT_API_URL=https://duckrails.gunraidersapi.com/api/v1
VUE_APP_PREFAB_ID=duckrails
VUE_APP_ENABLE_ATTESTATION=false
VUE_APP_K8S_NAMESPACE=duckrails
```

The full walkthrough below includes the admin panel. V1 has **no coins or wallet**; direct item IAP is required. Deploying this backend does not finish the remaining Duckz item-IAP/ownership integration. Wallet endpoints are not required and inherited wallet calls still need removal.

The push scripts build and push images; they do **not** update TKE Deployments. Select the new image tags in TKE after each push. No application ConfigMap, read-only database, Elasticsearch or JWT setup is required. The application, registry-pull and certificate-reference Secrets have separate purposes.

## 1. Create the database

1. Console, product **TencentDB for MySQL**. Use the same instance Duckz uses. Do not create a new instance.
2. Open the instance, then **Database Management** (DMC) and log in with the same account you use for Duckz.
3. The API database is `duckrails_tencent_global`. Open `api/sql/structure_classic.sql`; despite its inherited filename, it already creates and selects `duckrails_tencent_global`. Verify both the `CREATE DATABASE` and `USE` statements, then run that file for the new database. Do not change it to `gunraiders_classic` or `grc_tencent_global`, and do not use it as an upgrade script for an existing populated database.
4. In the database list, confirm `duckrails_tencent_global` exists and contains `players`, `transactions`, `items`, `skin_styles`, `platform_session_entries`, and `ip_blacklists`.

Record `DB_NAME=duckrails_tencent_global` and the private MySQL endpoint, including port, for `DB_HOST` in the build settings. If reusing the Duckz instance, its host is the same, but the application account is different.

Create account `duckrailsuser` on this instance. In **Account Management**, allow connections from the TKE pod network. `%` permits any source host at the MySQL account layer, so use it only with the intended private-network/security-group restrictions. A pod address such as `10.199.11.111` changes every time the pod restarts, so an account locked to one address is denied. Choose and save its password as `DB_PASSWORD` in root-level `runtime.env` (copy `runtime.env.example` first). Set `DB_ACCOUNT=duckrailsuser` in the build settings. Grant this account the application permissions for `duckrails_tencent_global`, including reads and writes. Do not create a second reader account. `Access denied for user 'duckrailsuser'@'...' (using password: YES)` can indicate a username, permitted host, password, or database grant mismatch. The process has already started.

## 2. Redis

Do not create a Redis instance. This setup reuses the Duckz Redis host and uses Duck Rails-specific `duckrails/` prefixes for its game/session keys. Set `REDIS_NAMESPACE=duckrails`; this scopes sessions, online-player keys, names, nonce records, scans, and response caching.

Copy `REDIS_HOST`, `REDIS_PORT`, `REDIS_USERNAME`, and `REDIS_PASSWORD` from the Duckz API workload. Put the host, port, and username in `api/gofiber/config/build.env`, and the password in the application Secret.

## 3. Configure the build

1. Open [api/gofiber/config/build.env](api/gofiber/config/build.env).
2. Set `DB_HOST`, `DB_NAME`, and `DB_ACCOUNT` using the database step. The host includes the port, for example `10.0.0.12:3306`. Keep `DB_PROTOCOL=tcp`.
3. Set `REDIS_HOST`, `REDIS_PORT`, and `REDIS_USERNAME`. Set `REDIS_NAMESPACE=duckrails`. Do not put the password here.
4. Set `OCULUS_APPID=28912202281742343` and `K8S_NAMESPACE=duckrails`. Use `API_VERSION=v1` for the API route prefix. `APP_VERSION` is backend configuration; it is separate from the Unity application version and Photon matchmaking version.
5. Keep `ENABLE_PHOTON_AUTH=1`: the Duck Rails client now sends Custom authentication. Generate the matching phrase as described in step 7 and configure BOTH Photon apps in step 12.
6. Remove every `REPLACE_*` placeholder. Keep passwords and private keys out of this file. The image sets `DEPLOY_ENV=production`; do not override that with a development value.
7. Fill root-level `runtime.env` with the credentials described in step 7 below. The application Secret is created after the namespace exists.

There is no application ConfigMap, read-only database configuration, Elasticsearch configuration, or required JWT key. Settings from this file are copied into the image; changing them later requires rebuilding.

## 4. Push the images

Do this from the repo on your PC before you create pods. The pods need an image tag that already exists in the registry.

1. In Tencent Container Registry, West US (Silicon Valley), shared namespace `gr-tcr`, create private repositories `duckrails-api-golang` and `duckrails-api-admin` if absent. Then start Docker Desktop with Linux containers. Copy `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat.template` to `sharedEnv.bat` in that same folder if it does not exist. Set `tencentBase=uswccr.ccs.tencentyun.com`, `tcrNamespace=gr-tcr`, and the registry login values `tcrUsername` and `tcrPass`. Keep this file local.
2. Commit the code and public build settings you want to ship. The image tag is the current git hash, so uncommitted edits are built but the tag will not say so.
3. Run the API script from its own directory (PowerShell, starting at the repository root):

   ```powershell
   Push-Location api/gofiber/cmd/tencent/tcr
   .\pushToSilliconValley.bat
   Pop-Location
   ```

   The root-level `Push GRC SV` shortcut, if present on your PC, is optional and is not tracked in the repo. The script builds the API and pushes `uswccr.ccs.tencentyun.com/gr-tcr/duckrails-api-golang:api-<git-hash>`.
4. For the admin panel, verify the already-configured `admin/.env.production` against the exact block at the top of this guide. Only copy the example if the file is missing, then replace its values with that block. Set both `VUE_APP_API_URL` and `VUE_APP_DEFAULT_API_URL` to `https://duckrails.gunraidersapi.com/api/v1`. That URL is baked into the admin image. Changing it later means building the admin image again.
5. Ensure `admin/misc/passwd` exists. The admin Dockerfile requires this ignored Apache Basic-auth password file. Reuse the existing Duck Rails file only for later updates; for this first deployment, do not assume a copied Classic password file is the intended credential. For a first installation, create it with Apache `htpasswd -cB admin/misc/passwd <username>` (run from the repository root; the tool prompts for the password). Do not use `-c` on an existing file unless intentionally replacing its accounts. Store the chosen login in your password manager. This is the browser Basic-auth gate, separate from the API token; the current setup does not require a second interactive login. Before the first admin build, generate the separate admin token using the generator in step 7, put it in `admin/src/stores/auth.js` as described there, and save the same value for the API Secret. Do this now; create the actual Secret after the namespace exists.
6. Run the admin script from its own directory:

   ```powershell
   Push-Location admin/cmd/tencent/tcr
   .\pushToSV-GRC.bat
   Pop-Location
   ```

   This pushes `uswccr.ccs.tencentyun.com/gr-tcr/duckrails-api-admin:adm-duckrails-<git-hash>`.

The login those scripts use is in `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat` (`tcrUsername` `200027379897`, password in `tcrPass`). Do not commit that file.

Confirm the images in the console:

1. Tencent Cloud Console, product **Tencent Container Registry**.
2. Region **West US (Silicon Valley)**. The registry host is `uswccr.ccs.tencentyun.com`.
3. Open namespace **gr-tcr**. This is the shared registry namespace. Duck Rails does not get its own TCR namespace.
4. Open repository `duckrails-api-golang` and confirm a tag `api-<git-hash>`.
5. Open repository `duckrails-api-admin` and confirm a tag `adm-duckrails-<git-hash>`.

These batch scripts do not reliably stop after every failed build command. Read the output and verify both pushes succeeded; a script reaching its final pause is not proof of success.

Write those two full image names down. You will paste them into the workloads.

## 5. Create the TKE namespace

1. Console, product **Tencent Kubernetes Engine**.
2. Open cluster **rc-tke** in West US (Silicon Valley).
3. Left menu, **Namespace**.
4. **Create**. Name it `duckrails`. Leave the other fields at their defaults.
5. Confirm the new namespace is in the list next to `duckz`.

Every object below must be created while this namespace is selected. If the page header still says `duckz`, you are in the wrong place.

## 6. Create the image pull secret

Pods cannot download a private `gr-tcr` image until this secret exists in `duckrails`. The Duckz secret `rc-tcr-key` is the same kind of object, but it only works inside `duckz`. Create one with the same name in `duckrails`.

1. In cluster `rc-tke`, left menu **Configuration Management**, then **Secret**.
2. Set the namespace filter to `duckrails`. The Duckz secrets should disappear.
3. **Create**.
4. Name: `rc-tcr-key`.
5. Type: Docker registry credential (`kubernetes.io/dockerconfigjson`), not Opaque.
6. Image repository domain: `uswccr.ccs.tencentyun.com`.
7. Username: `200027379897`.
8. Password: the `tcrPass` value in `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat`. Confirm it.
9. Create. The type column for this row should read `kubernetes.io/dockerconfigjson`. The domain is only `uswccr.ccs.tencentyun.com`. A pull failure may be reported as `not found` or unauthorized; check both the image name/tag and registry credentials.

## 7. Create the application Secret

In cluster `rc-tke`, select namespace `duckrails`, then **Configuration Management > Secret > Create**. Name it `game-api-env` and choose **Opaque**. Add these keys and supply their values:

```dotenv
DB_PASSWORD=
REDIS_PASSWORD=
OCULUS_APPSECRET=
PHOTON_AUTH_PHASE=
ADMIN_API_TOKEN=
META_EDITOR_USER_IDS=
```

- `DB_PASSWORD`: the password you set for `duckrailsuser` in step 1.
- `REDIS_PASSWORD`: existing password for the shared Redis instance/account. Do not reset that shared password.
- `OCULUS_APPSECRET`: Meta developer dashboard, Duck Rails app `28912202281742343`, its App Secret. It must match the App ID baked into the API image.
- `PHOTON_AUTH_PHASE`: generate a new random phrase locally using the command below. Save it and use exactly the same value for the `secret` query parameter on BOTH Photon apps in step 12. Never add it to Unity.
- `ADMIN_API_TOKEN`: generate a second, different random value if using the admin panel. Replace the literal passed to `tmpLoginForever(...)` inside `ensureAuthenticated()` in `admin/src/stores/auth.js` with that value before building the admin image. Put the identical raw value here (no `Bearer ` prefix). This existing frontend embeds the token in its bundle; anyone with it has shared API admin access. It is separate from the Apache password and all game credentials. If the image was already built, rebuild it after changing this value.
- `META_EDITOR_USER_IDS`: optional comma-separated organization-scoped IDs from the Unity `[SignIn] Editor Meta organization-scoped user ID` log. Empty disables Editor access. These IDs do not grant administrator privileges.

Generate each random phrase/token separately in PowerShell:

```powershell
$phraseBytes = New-Object byte[] 32
$phraseGenerator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$phraseGenerator.GetBytes($phraseBytes)
$phraseGenerator.Dispose()
[BitConverter]::ToString($phraseBytes).Replace('-', '').ToLowerInvariant()
```

Also save the values in the ignored root `runtime.env` if using it as your configuration record. Do not commit credentials. Create the Secret before the workload; then import it only into the API deployment. Do not attach it to the admin container. Retain the separate `rc-tcr-key` from step 6.

## 8. Upload a Cloudflare origin certificate

Duckz did not buy a certificate from Tencent. Its row in **My Certificates** is `cloudflare-dz.gunraidersapi.com`, source **Upload certificate**, issuer **CloudFlare Origin Certificate**. This Duck Rails setup uses the same kind of certificate. Cloudflare signs it, and you paste it into Tencent. There is no DNS validation step.

`gunraidersapi.com` is already a site in Cloudflare. Do not add the site again. This setup gives Duck Rails one public hostname, `duckrails.gunraidersapi.com`, and one certificate for that name. The admin panel is a path on that same host, `https://duckrails.gunraidersapi.com/admin/`.

Create it in Cloudflare:

1. Cloudflare dashboard, zone **gunraidersapi.com**.
2. **SSL/TLS**, then **Origin Server**.
3. **Create Certificate**.
4. Private key: **Generate private key and CSR with Cloudflare**. Key type **RSA**.
5. Hostnames: remove the defaults (`gunraidersapi.com` and `*.gunraidersapi.com`). Add only `duckrails.gunraidersapi.com`.
6. Certificate validity: **15 years**.
7. **Create**.
8. Key format: **PEM**.
9. Copy the **Origin Certificate** and the **Private Key** into two text files on your machine. Cloudflare shows the private key on this screen only. Do not put either file in the git repo.

Upload it to Tencent:

**Always include both certificates:** paste your generated Origin Certificate first, then append the Cloudflare Origin RSA root certificate shown below in the same Certificate field. The extra root certificate is required for this setup, not an optional recovery step. Complete this upload before creating the pods.

1. Tencent Cloud Console, product **SSL Certificate Service**, page **My Certificates**.
2. **Upload Certificate**.
3. Standard: **International standard**.
4. Certificate name: `cloudflare-duckrails.gunraidersapi.com`.
5. Certificate: paste the Origin Certificate, including the `-----BEGIN CERTIFICATE-----` and `-----END CERTIFICATE-----` lines. On the next line, paste the Cloudflare Origin RSA root below. The certificate box then holds two certificates, one after the other. The private key box still holds only the private key.
6. Private key: paste the Private Key, including its BEGIN and END lines.
7. **Upload**. If Tencent opens **Incomplete certificate chain**, click **Cancel**. The certificate box is missing the root below. Add it after the origin certificate, then upload again. Do not click **OK** on that dialog. Continuing leaves a chain that the load balancer can refuse.

```
-----BEGIN CERTIFICATE-----
MIIEADCCAuigAwIBAgIID+rOSdTGfGcwDQYJKoZIhvcNAQELBQAwgYsxCzAJBgNV
BAYTAlVTMRkwFwYDVQQKExBDbG91ZEZsYXJlLCBJbmMuMTQwMgYDVQQLEytDbG91
ZEZsYXJlIE9yaWdpbiBTU0wgQ2VydGlmaWNhdGUgQXV0aG9yaXR5MRYwFAYDVQQH
Ew1TYW4gRnJhbmNpc2NvMRMwEQYDVQQIEwpDYWxpZm9ybmlhMB4XDTE5MDgyMzIx
MDgwMFoXDTI5MDgxNTE3MDAwMFowgYsxCzAJBgNVBAYTAlVTMRkwFwYDVQQKExBD
bG91ZEZsYXJlLCBJbmMuMTQwMgYDVQQLEytDbG91ZEZsYXJlIE9yaWdpbiBTU0wg
Q2VydGlmaWNhdGUgQXV0aG9yaXR5MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRMw
EQYDVQQIEwpDYWxpZm9ybmlhMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAwEiVZ/UoQpHmFsHvk5isBxRehukP8DG9JhFev3WZtG76WoTthvLJFRKFCHXm
V6Z5/66Z4S09mgsUuFwvJzMnE6Ej6yIsYNCb9r9QORa8BdhrkNn6kdTly3mdnykb
OomnwbUfLlExVgNdlP0XoRoeMwbQ4598foiHblO2B/LKuNfJzAMfS7oZe34b+vLB
yrP/1bgCSLdc1AxQc1AC0EsQQhgcyTJNgnG4va1c7ogPlwKyhbDyZ4e59N5lbYPJ
SmXI/cAe3jXj1FBLJZkwnoDKe0v13xeF+nF32smSH0qB7aJX2tBMW4TWtFPmzs5I
lwrFSySWAdwYdgxw180yKU0dvwIDAQABo2YwZDAOBgNVHQ8BAf8EBAMCAQYwEgYD
VR0TAQH/BAgwBgEB/wIBAjAdBgNVHQ4EFgQUJOhTV118NECHqeuU27rhFnj8KaQw
HwYDVR0jBBgwFoAUJOhTV118NECHqeuU27rhFnj8KaQwDQYJKoZIhvcNAQELBQAD
ggEBAHwOf9Ur1l0Ar5vFE6PNrZWrDfQIMyEfdgSKofCdTckbqXNTiXdgbHs+TWoQ
wAB0pfJDAHJDXOTCWRyTeXOseeOi5Btj5CnEuw3P0oXqdqevM1/+uWp0CM35zgZ8
VD4aITxity0djzE6Qnx3Syzz+ZkoBgTnNum7d9A66/V636x4vTeqbZFBr9erJzgz
hhurjcoacvRNhnjtDRM0dPeiCJ50CP3wEYuvUzDHUaowOsnLCjQIkWbR7Ni6KEIk
MOz2U0OBSif3FTkhCgZWQKOOLo1P42jHC3ssUZAtVNXrCk3fw9/E15k8NPkBazZ6
0iykLhH1trywrKRMVw67F44IE8Y=
-----END CERTIFICATE-----
```

The new row should look like the Duckz row: source **Upload certificate**, status **Issued**, and the domain list shows `duckrails.gunraidersapi.com` and **CloudFlare Origin Certificate**.

This certificate is trusted by Cloudflare, not by a browser that talks to the load balancer directly. The `duckrails` DNS record in step 12 has to stay proxied.

## 9. Create the API deployment

If you already created an API deployment without these objects, edit that deployment instead of creating a second one. Saving it restarts the pods.

Create one API deployment. In this Tencent workload-creation flow, its Service is created with the workload; there is no separate Create Services step. Set the generated `v1-api-deployment` Service port to `3000`, targeting container port `3000`. After creation, confirm that Service exists and selects the API pods.

1. Left menu **Workload**, **Deployment**.
2. Namespace `duckrails`. **Create**.
3. Name: `v1-api-deployment`. Replicas: use the existing Duck Rails count for updates. For a first test deployment, start with 1 and size it for production after verification.
4. On the container. The form splits the image across two fields:
   - Name: `api`.
   - Image: `uswccr.ccs.tencentyun.com/gr-tcr/duckrails-api-golang`.
   - Image tag: `api-<git-hash>` from step 4. The tag includes the `api-` prefix.
   - Image pull policy: **Always**.
   - Image access credential: `rc-tcr-key`. The dropdown only lists docker-registry secrets in this namespace. If it is empty, step 6 was done in `duckz` or was skipped.
   - Container port: `3000`. The process listens on 3000.
5. Environment variables: import all keys from Secret `game-api-env`. You may keep a separate Custom `API_VERSION=v1` override in Tencent for flexibility; it overrides the image default. When migrating, compare old overrides with `build.env`, retain any intentionally needed integration settings, and remove obsolete application ConfigMap imports and duplicate values. Do not remove registry or certificate references.
6. Health checks, both HTTP. Set timeout to 10 seconds to allow both dependency checks to finish:
   - Liveness: port `3000`, path `/live`.
   - Readiness: port `3000`, path `/ready`.
7. Create.

Open the new pod and check its status:

- **Running and Ready**: required before sending traffic. Running alone is insufficient.
- **ImagePullBackOff**: inspect the pod events for image name/tag, registry credentials, and registry reachability. Confirm `rc-tcr-key` is in `duckrails` and selected on the workload.
- **CrashLoopBackOff**: inspect the application log. Missing build settings or credentials are reported by key name. Check MySQL connectivity/grants before rebuilding.
- **Running but not Ready**: inspect `/ready` and its dependency checks. Check the database and Redis errors returned by the readiness endpoint.

## 10. Create the admin deployment

This workload-creation flow also creates the `admin-deploy` Service. Use Service port `80` targeting container port `80`; verify the generated Service after creation instead of creating a duplicate.

1. **Workload**, **Deployment**, namespace `duckrails`, **Create**.
2. Name: `admin-deploy`. Replicas: 1.
3. Container name: `admin`.
4. Image: `uswccr.ccs.tencentyun.com/gr-tcr/duckrails-api-admin`. Image tag: `adm-duckrails-<git-hash>` from step 4.
5. Image pull policy: **Always**. Image access credential: `rc-tcr-key`.
6. Container port: `80`. Apache in that image listens on 80. The panel files are served under `/admin/`.
7. Do not attach `game-api-env` to the admin deployment. The admin image does not read it. Its API URL and Apache password file are baked in at build time. No custom runtime environment variables are required. Remove explicit `PATH` and `HTTPD_*` overrides; their defaults come from the Apache image. Do not remove inherited image defaults.
8. Liveness and readiness: **TCP port `80`**. `/admin/` is protected by Apache Basic authentication and normally returns 401 without credentials, so an unauthenticated HTTP probe there would fail. Verify the authenticated page manually after deployment.
9. Create.

## 11. Create a public load balancer for Duck Rails

Duckz already has a balancer. Leave it alone. This guide uses a separate Duck Rails balancer to avoid changing Duckz routing. Tencent also supports explicitly configured shared-CLB ingress groups; that is a separate design, not something to enable by simply choosing the Duckz balancer. [Tencent shared-CLB documentation](https://intl.cloud.tencent.com/ind/document/product/457/78649).

You do not create the balancer in the Cloud Load Balancer product. Creating the Ingress below creates it.

1. Stay in cluster `rc-tke`. Left menu **Services and Routes**, then **Ingress**.
2. Set the namespace to `duckrails`. **Create**.
3. Ingress name: `duckrails-lb`. Use `duckrails-lb` for the generated load balancer's display name as well; the TKE namespace remains `duckrails`.
4. Namespace: `duckrails`.
5. Ingress type: the CLB type (the Tencent load balancer controller). If the form lists Nginx as well, do not pick Nginx.
6. Network type: **Public network**.
7. Load balancer: **Automatically create**. Do not choose the Duckz balancer.
8. IP version: **IPv4**. VPC: **Current VPC**. Carrier: **BGP**. Review the displayed billing and instance-size options and their recurring price before creating this new paid balancer.
9. Match the Duckz load balancer options on this same form:
   - **Network mode**: **Disable CLB-To-Pod Direct Access**.
   - **Bypassing by Default**: leave **Enable synchronization of Bypass Backend Security Group configurations** unchecked.
   - **Security group**: leave **Enable synchronization of security group configurations** unchecked.
   - **Redirect**: **Custom**. Leave the redirect list empty. Duckz does not send one path to another.
   - **Custom Listening Port**: checked. Add **HTTP** port **80** and **HTTPS** port **443**.
10. **Complete TLS configuration before adding forwarding rules.** On the **HTTPS 443** listener, go to **TLS configuration** and click **create a new one** beside the Secret selector.
    - Select the certificate already uploaded in step 8: `cloudflare-duckrails.gunraidersapi.com`.
    - Create the certificate-reference Secret. TKE stores the certificate ID; do not paste the private key again.
    - Return to the Ingress form, refresh the Secret list, and select that new Secret.
    - Set **Domain** to `duckrails.gunraidersapi.com`.
    - **Check Default certificate.** This is required in this setup. Finish these TLS fields before configuring forwarding.
11. Add these forwarding rules on the **HTTPS 443** listener. Select each Service before selecting its port:
    - Domain `duckrails.gunraidersapi.com`, path `/admin`, Service `admin-deploy`, port `80`.
    - Domain `duckrails.gunraidersapi.com`, path `/`, Service `v1-api-deployment`, port `3000`.
    These are the Services created with the workloads in steps 9 and 10. The API Service and container both use `3000`; admin uses `80`. The external HTTPS listener remains `443`.
12. On the **HTTP 80** listener, add the same two forwarding rules (API backend port `3000`, admin backend port `80`) without a certificate.
13. **Create Ingress**.

Keep the certificate configured on the Ingress. A certificate added only on the load balancer page can be overwritten by the ingress controller.

The new row shows a load balancer ID. Click it to open the new balancer in West US (Silicon Valley). Set/confirm its display name is `duckrails-lb` so it is easy to identify. Copy its **public IP address** for the Cloudflare A record in step 12. Use this new balancer's IP, not Duckz's.

## 12. Cloudflare and Photon

These are not Tencent pages. They are required before the game can call the API.

Cloudflare, zone `gunraidersapi.com`:

1. Add an **A** record. Name: `duckrails`. IPv4 address: the public IP of `duckrails-lb` from step 11. The ingress sends `/` to the API and `/admin` to the admin panel.
2. Proxy status: **Proxied** (orange cloud). The origin certificate from step 8 is only trusted by Cloudflare. A grey-cloud record makes the browser talk to Tencent directly and reject the certificate.
3. SSL/TLS mode **Full (strict)** for this hostname. If changing the zone-wide mode, first confirm the other hosts in the zone support it; use a hostname-specific configuration rule when needed. Cloudflare speaks HTTPS to the balancer and checks the origin certificate. [Cloudflare Origin CA guidance](https://developers.cloudflare.com/ssl/origin-configuration/origin-ca/).

Photon dashboard: repeat these settings on **both** Duck Rails apps:

- Realtime: `36bef5d2-63ad-4567-948e-52e966d7f2e9`.
- Voice: `b7c80e26-6a61-45ca-97d9-478af736a6d3`.

1. Custom authentication URL: `https://duckrails.gunraidersapi.com/api/photon/authenticate`.
2. Add a Custom Server authentication provider. Add a dashboard query parameter named `secret` whose value exactly matches `PHOTON_AUTH_PHASE` in `game-api-env`.
3. Reject authentication when the provider is unavailable, and disable anonymous connections. Allow dashboard changes time to propagate.
4. The Unity client sends `usr`, `pla`, and its current session `tkn`. Do not put the shared phrase into Unity. The dashboard supplies it server-side.
5. The Unity assets already contain the two confirmed app IDs above. `PhotonVoiceChat` uses `UsePunAuthValues=true`, so configure this same Custom provider on Voice as well as Realtime. Keep `ENABLE_PHOTON_AUTH=1`.

## 13. Check

1. Pods belonging to `v1-api-deployment` are Running and Ready in namespace `duckrails`.
2. `https://duckrails.gunraidersapi.com/api/v1/info/all` returns 200 and a JSON body with `data.version`.
3. The admin pod is Running, and `https://duckrails.gunraidersapi.com/admin/` prompts for Basic authentication and then loads the admin frontend. The website sends its existing Authorization token; it must match `ADMIN_API_TOKEN` on the API. No second login screen is needed for this mode. Verify an admin API request succeeds, not just that the website loads.
4. Test Quest sign-in using a build installed from the correct Meta release channel. Confirm the Unity Meta App ID matches the dashboard, the account has access, and User ID/User Profile grants are active in Data Use Checkup. For Editor testing, configure the organization-scoped allowlist below. Confirm a row appears in this game's `players` table.
5. Verify a write followed by a read in the game and a reconnect/session refresh. Test Photon authentication with the matching client.
6. Confirm no old ConfigMap values or `_RO`/`ES_*` settings are needed. Attestation-history endpoints now return 501 by design; remote Meta verification and bans remain available.

## Updating an existing Duck Rails deployment

1. Do not recreate the database, namespace, Services, certificate, DNS record, or ingress. Record the currently deployed image tags and environment references for rollback.
2. Update public build settings before building. Keep the existing `REDIS_NAMESPACE` when updating a game. Preserve the existing game/Photon credentials unless deliberately rotating them.
3. Create or update `game-api-env` in `duckrails`. Updating a Secret alone does not refresh a running container's environment.
4. Build/push a new API image using section 4. Build the admin image only if its code, baked API URL, or Basic-auth file changed.
5. Edit the existing API Deployment: select the new image tag, import `game-api-env`, and remove superseded configuration imports after comparing their values with the new build settings. Saving a pod-template change triggers a rollout; if only the Secret changed, explicitly restart the Deployment.
6. Verify section 13 before removing old configuration objects or retiring a second database account. Keep them through the rollback window; restoring an old image may also require restoring its old configuration references.

Only one application Secret is needed. The image-pull Secret and certificate-reference Secret still serve separate purposes and remain in place.

The supported static-token admin setup is covered in steps 4 and 7 above. No SQL administrator grant is needed for that mode. The alternative player-session mode in ADMIN-SETUP.md is outside this walkthrough; its copied SQL still contains Classic database names.

### Unity Editor sign-in

Quest builds continue to use device attestation. Editor requests can instead authenticate with a Meta user proof, but only for accounts explicitly allowed on the backend.

1. Deploy the backend version containing `meta_editor_login.go`.
2. In Unity, use **Meta > Platform > Edit Settings** to sign in the test user for the Duck Rails Meta app (28912202281742343).
3. Enter Play mode. The Console prints `[SignIn] Editor Meta organization-scoped user ID: ...`.
4. Add `META_EDITOR_USER_IDS` to the existing Tencent `game-api-env` Secret, with that organization-scoped ID as its value. Multiple IDs are comma-separated. Also add the setting to your local `runtime.env` if you maintain it as the source for the Secret.
5. Restart/roll out the API deployment so the new environment value takes effect, then retry Play mode.

This setting is optional and disabled when empty. Use the organization-scoped ID from the log, which is also the player ID. Existing app-scoped allowlist entries must be replaced when deploying this version. No Editor bypass password is required. The backend verifies the proof with Meta and verifies the matching organization-scoped identity before issuing the normal session. Normal player bans still apply; device attestation is not available in the Editor. Remove an ID to prevent future Editor logins; existing sessions retain their normal lifetime.

## Resources to keep after migration

After deployment, the Duck Rails namespace should retain these Secrets:

- `game-api-env`: API runtime settings, imported only by the API workload.
- `rc-tcr-key`: private registry credentials, referenced by API and admin image-pull settings.
- The certificate-reference Secret generated in step 11: record its actual name, select it in ingress TLS configuration, and verify `spec.tls[].secretName`. Do not select Classic's or Duckz's certificate Secret.

Keep the Kubernetes-managed `kube-root-ca.crt` ConfigMap if present or recreated. Old application ConfigMaps (`photon-id`, `photon-session-check`, `remote-api-config`, `shared-cache`, `shared-database`) and old credential Secrets are not required by this profile. Before deleting, check every workload's environment/volume references, image-pull references, service accounts, and ingress TLS references in that namespace. Two Secrets containing `qcloud_cert_id` are not necessarily redundant; retain each referenced one. Verify replacement pods start before retiring old objects.

## Troubleshooting this setup

- `/live` and `/ready` returning HTTP 200 with `{}` is normal. If API endpoints work but health URLs time out externally, check that ingress forwards `/` to the API, not only `/api/v1`. Photon also uses the unversioned `/api/photon/authenticate` path.
- For `v1-api-deployment`, set the Service port and container target port to `3000`, then select port `3000` in both ingress forwarding rules for `/`. For `admin-deploy`, use `80`. The public listener ports remain HTTP `80` and HTTPS `443`. Services are created with the workloads in this Tencent flow; do not create duplicates.
- Editor `missing authorization header`: verify the running image contains Editor login support and the request includes `X-Meta-Editor-Login: 1`. An unapproved organization-scoped ID produces a different error. Secret changes require new pods.
- Admin requests to `classic-api.example.com`: fill both URLs in `admin/.env.production`, rebuild/push the admin image, roll out the admin deployment, then hard-refresh the browser. Workload environment variables do not change bundled JavaScript.
- Admin 401: verify `ADMIN_API_TOKEN` is loaded in the API pod and exactly matches the browser's Authorization value (without `Bearer ` in the Secret). Clear stale `acc_token` cookies if the frontend token changed. A valid player session lacking admin privileges returns 403 instead.
- An empty skin-style table now returns HTTP 200 with `data: []`. The previous `404 / no row found` behavior requires an API image update.
- Player IP storage now keeps one validated, normalized address, not two comma-separated forwarded addresses. This fixes `1406 data too long for column ip` without changing the 45-character schema.
- No public ban-list download remains. Use the authenticated per-player `/api/v1/serviceban/get` endpoint. `HASH_BANALL` and `ENABLE_SWAGGER` are removed.

## Unity client integration notes

The client is `X:\Work\duck-rails`, using Unity 6000.3.24f1. Its backend URL is `backendBaseUrl` in `Assets/Resources/GameConfig.asset` (and the default in `Assets/Scripts/Config/GameConfig.cs`), already set to `https://duckrails.gunraidersapi.com/api/v1/`. There is no Classic GameManager URL field to edit in this client.

`Assets/Resources/OculusPlatformSettings.asset` already has mobile Meta app `28912202281742343`. `Assets/Photon/PhotonUnityNetworking/Resources/PhotonServerSettings.asset` already has the confirmed Realtime and Voice IDs listed above. Verify the Android bundle identifier, release channel and signing configuration in this Unity project before uploading builds; do not follow Classic-specific build-menu or keystore names.

`GameSessionClient` posts multipart data to `/api/v1/player/v2/social-account/oculus`. Quest uses device attestation; Editor sends `X-Meta-Editor-Login: 1` with a real Meta user proof. The backend validates the proof and organization identity before checking the Editor allowlist. `DuckzApiClient` refreshes through `POST /api/v1/session/{session_id}`. `PhotonSession` supplies `usr`, `pla`, and `tkn` for Custom authentication and preserves Photon's issued token during room changes.

Local backend authentication/configuration tests and both backend/C# compilation passed. No live deployment, external sign-in or headset acceptance has been completed. Test login, session refresh, Realtime/Voice, room changes, and reconnect with two authenticated accounts/headsets. The inherited additional-Editor local-session shortcut has no backend token and cannot validate secured Photon authentication.

**V1 scope:** no coins and no wallet. Do not port those endpoints. Direct item IAP, transaction verification/deduplication, ownership and applicable restoration are required; exact product types still need definition. The copied Classic backend does not yet provide the Duckz-specific item IAP/loadout/cosmetics routes used by this Unity client. Those requests are asynchronous after login, so they do not gate authentication, but deploying this guide alone does not make purchases or cosmetic restoration work. Inherited wallet requests still need removal.

### Database feature cleanup

The inherited schema keeps accounts, admin privileges, service/IP bans, build hashes, SQL session history, skins, owned inventory, and Meta purchase records. News and cloud player-prefab endpoints and their admin UI have been removed. Legacy friends, match statistics, leaderboards, battle passes, quests, bounties, and XP tables are not used by this backend. Deploy the updated API and admin before removing retired tables. Keep the existing `reports` table: account migration still updates its reporter and reported player IDs. Its legacy CREATE TABLE definition must be carried into a new game database separately; `structure_classic.sql` does not currently define it. Do not drop it or substitute an invented schema.

Redis stores live sessions and refresh tokens; `platform_session_entries` retains SQL session records. Keep both. Add `ip_blacklists` from the schema if missing: otherwise IP checks log a database error and allow the request through.

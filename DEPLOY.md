# Deploy a game backend on the existing Tencent stack

Use this guide to deploy a new game on the existing service stack, or update an existing game. [DEPLOY-CONFIG.md](DEPLOY-CONFIG.md) explains each build setting and credential; [runtime.env.example](runtime.env.example) is the credential template. Public settings are baked into the API image, and one application Secret supplies credentials. The registry pull Secret and ingress certificate Secret are separate infrastructure objects.

Classic runs on the same Tencent account as Duckz. It gets its own database name, its own TKE namespace, its own pods, and its own public load balancer. It reuses the Silicon Valley registry, the Redis host, and the TKE cluster `rc-tke`.

Do the sections in order. The workload form can only attach a Secret that already exists in the same namespace. Duckz objects live in namespace `duckz` and will not appear while the namespace selector says `grc`.

The cluster from the Duckz secret page is West US (Silicon Valley), cluster `rc-tke` (`cls-1gdfwglk`).

For an existing deployment, use the update checklist at the end. Reuse existing resources instead of creating duplicates. This guide records the intended Classic configuration; verify live resource names and controller settings in Tencent before applying changes.

Health probes: `/live` checks that the API is serving requests; `/ready` checks MySQL and Redis. The old example failure, unrelated public DNS check, and arbitrary goroutine limit have been removed. A dependency outage makes the pod unready without forcing a process restart.


## Start here for a new game

Assume the Tencent account, Silicon Valley registry, TKE cluster, MySQL instance, Redis instance, and Cloudflare zone already exist. Do not provision replacements. The detailed steps below use Classic as a worked example; substitute your new game's values throughout.

Before editing anything, choose and record:

- A game slug and TKE namespace, such as `newgame`.
- A database, such as `newgame_tencent_global`, and one database account with read/write access to it.
- A unique, stable `REDIS_NAMESPACE`, such as `newgame-prod`.
- A hostname, such as `newgame.gunraidersapi.com`, and its certificate/ingress routing.
- Separate API and admin registry repositories, such as `newgame-api-golang` and `newgame-api-admin`.
- The game's Meta App ID/App Secret and Photon Realtime app (and Voice app if used).

For a new game, edit `tcrRepo` in copies of both push scripts before running them: the checked-in scripts publish to Classic's repositories. Adapt the admin tag prefix too, or explicitly document that you retained it. Do not overwrite an existing game's configuration or deploy its image to the wrong namespace. Use a separate checkout or game-specific configuration copy when maintaining multiple games.

Configure locally:

1. `api/gofiber/config/build.env`: non-secret API settings.
2. Root `runtime.env`, copied from `runtime.env.example`: runtime credentials and optional Editor allowlist. This file is not included in the image.
3. `admin/.env.production`, copied from its example: both API URLs, prefab ID, and Kubernetes namespace. The build refuses a missing file or the example hostname.
4. `admin/misc/passwd`: Apache password hashes for the website prompt.
5. `admin/src/stores/auth.js`: the existing static admin token used by `ensureAuthenticated()`; see [ADMIN-SETUP.md](ADMIN-SETUP.md) for matching it to the backend. This is not currently an environment-variable setting in the frontend.
6. Local registry login file `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat`.

In Tencent, each game uses one application Secret, `game-api-env`, plus a registry pull Secret and an ingress certificate-reference Secret. The same Secret names can be reused in different namespaces. No application ConfigMap, read-only database, Elasticsearch, Swagger, or JWT setup is required. The per-game namespace, database/schema/account, image repositories, deployments, Services, and hostname routing still need to be created if they do not exist; the underlying service instances are reused.

The batch files build and push images to TCR; they do **not** update TKE Deployments. Select the new tags and complete the rollout in TKE after each push.

## 1. Create the database

1. Console, product **TencentDB for MySQL**. Use the same instance Duckz uses. Do not create a new instance.
2. Open the instance, then **Database Management** (DMC) and log in with the same account you use for Duckz.
3. The API database is `grc_tencent_global`. `api/sql/structure_classic.sql` starts by creating `gunraiders_classic`. In a copy of the SQL file, change both the `CREATE DATABASE` and `USE` statements to `grc_tencent_global`, verify the target, then run it for a new database. Do not run this as an upgrade/migration procedure for an existing populated database.
4. In the database list, confirm `grc_tencent_global` exists and contains `players`, `transactions`, `items`, `skin_styles`, `platform_session_entries`, and `ip_blacklists`.

Record `DB_NAME=grc_tencent_global` and the private MySQL endpoint, including port, for `DB_HOST` in the build settings. If reusing the Duckz instance, its host is the same, but the application account is different.

Create account `grcuser` on this instance. In **Account Management**, allow connections from the TKE pod network. `%` permits any source host at the MySQL account layer, so use it only with the intended private-network/security-group restrictions. A pod address such as `10.199.11.111` changes every time the pod restarts, so an account locked to one address is denied. Choose and save its password as `DB_PASSWORD` in root-level `runtime.env` (copy `runtime.env.example` first). Set `DB_ACCOUNT=grcuser` in the build settings. Grant this account the application permissions for `grc_tencent_global`, including reads and writes. Do not create a second reader account. `Access denied for user 'grcuser'@'...' (using password: YES)` can indicate a username, permitted host, password, or database grant mismatch. The process has already started.

## 2. Redis

Do not create a Redis instance. Classic uses the Duckz Redis host and uses Classic-specific `grc/` prefixes for its game/session keys. Set `REDIS_NAMESPACE=grc` for Classic. Use a unique namespace for every other game/environment sharing Redis, for example `newgame-prod`; this scopes sessions, online-player keys, names, nonce records, scans, and response caching.

Copy `REDIS_HOST`, `REDIS_PORT`, `REDIS_USERNAME`, and `REDIS_PASSWORD` from the Duckz API workload. Put the host, port, and username in `api/gofiber/config/build.env`, and the password in the application Secret.

## 3. Configure the build

1. Open [api/gofiber/config/build.env](api/gofiber/config/build.env).
2. Set `DB_HOST`, `DB_NAME`, and `DB_ACCOUNT` using the database step. The host includes the port, for example `10.0.0.12:3306`. Keep `DB_PROTOCOL=tcp`.
3. Set `REDIS_HOST`, `REDIS_PORT`, and `REDIS_USERNAME`. Set `REDIS_NAMESPACE=grc` for Classic, or a unique game/environment name for another backend. Do not put the password here.
4. Set `OCULUS_APPID` to the Classic Meta application ID and `K8S_NAMESPACE=grc`. Use `API_VERSION=v1` for the API route prefix. `APP_VERSION` is backend configuration; it is separate from the Unity application version and Photon matchmaking version.
5. Keep `ENABLE_PHOTON_AUTH=1` when using custom authentication and prepare the matching phrase now. If intentionally postponing it, set this flag to `0`; enabling it later requires a rebuild or explicit workload override.
6. Remove every `REPLACE_*` placeholder. Keep passwords and private keys out of this file. The image sets `DEPLOY_ENV=production`; do not override that with a development value.
7. Follow [DEPLOY-CONFIG.md](DEPLOY-CONFIG.md#2-obtain-the-four-credentials) to fill root-level `runtime.env`. The application Secret is created after the namespace exists.

There is no application ConfigMap, read-only database configuration, Elasticsearch configuration, or required JWT key. Settings from this file are copied into the image; changing them later requires rebuilding.

## 4. Push the images

Do this from the repo on your PC before you create pods. The pods need an image tag that already exists in the registry.

1. Start Docker Desktop with Linux containers. Copy `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat.template` to `sharedEnv.bat` in that same folder if it does not exist. Set `tencentBase=uswccr.ccs.tencentyun.com`, `tcrNamespace=gr-tcr`, and the registry login values `tcrUsername` and `tcrPass`. Keep this file local.
2. Commit the code and public build settings you want to ship. The image tag is the current git hash, so uncommitted edits are built but the tag will not say so.
3. Run the API script from its own directory (PowerShell, starting at the repository root):

   ```powershell
   Push-Location api/gofiber/cmd/tencent/tcr
   .\pushToSilliconValley.bat
   Pop-Location
   ```

   The root-level `Push GRC SV` shortcut, if present on your PC, is optional and is not tracked in the repo. The script builds the API and pushes `uswccr.ccs.tencentyun.com/gr-tcr/grc-api-golang:api-<git-hash>`.
4. For the admin panel, first put the real API URL in `admin/.env.production` (copy `admin/.env.production.example`). Set both `VUE_APP_API_URL` and `VUE_APP_DEFAULT_API_URL` to `https://grc.gunraidersapi.com/api/v1`. That URL is baked into the admin image. Changing it later means building the admin image again.
5. Ensure `admin/misc/passwd` exists. The admin Dockerfile requires this ignored Apache Basic-auth password file. Reuse the existing Classic file for updates. For a first installation, create it with Apache `htpasswd -cB admin/misc/passwd <username>` (run from the repository root; the tool prompts for the password). Do not use `-c` on an existing file unless intentionally replacing its accounts. Store the chosen login in your password manager. This is the browser Basic-auth gate, separate from the API token; the current setup does not require a second interactive login. Before the first admin build, complete the token setup in [ADMIN-SETUP.md](ADMIN-SETUP.md).
6. Run the admin script from its own directory:

   ```powershell
   Push-Location admin/cmd/tencent/tcr
   .\pushToSV-GRC.bat
   Pop-Location
   ```

   This pushes `uswccr.ccs.tencentyun.com/gr-tcr/grc-api-admin:adm-grc-<git-hash>`.

The login those scripts use is in `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat` (`tcrUsername` `200027379897`, password in `tcrPass`). Do not commit that file.

Confirm the images in the console:

1. Tencent Cloud Console, product **Tencent Container Registry**.
2. Region **West US (Silicon Valley)**. The registry host is `uswccr.ccs.tencentyun.com`.
3. Open namespace **gr-tcr**. This is the shared registry namespace. Classic does not get its own TCR namespace.
4. Open repository `grc-api-golang` and confirm a tag `api-<git-hash>`.
5. Open repository `grc-api-admin` and confirm a tag `adm-grc-<git-hash>`.

These batch scripts do not reliably stop after every failed build command. Read the output and verify both pushes succeeded; a script reaching its final pause is not proof of success.

Write those two full image names down. You will paste them into the workloads.

## 5. Create the TKE namespace

1. Console, product **Tencent Kubernetes Engine**.
2. Open cluster **rc-tke** in West US (Silicon Valley).
3. Left menu, **Namespace**.
4. **Create**. Name it `grc`. Leave the other fields at their defaults.
5. Confirm the new namespace is in the list next to `duckz`.

Every object below must be created while this namespace is selected. If the page header still says `duckz`, you are in the wrong place.

## 6. Create the image pull secret

Pods cannot download a private `gr-tcr` image until this secret exists in `grc`. The Duckz secret `rc-tcr-key` is the same kind of object, but it only works inside `duckz`. Create one with the same name in `grc`.

1. In cluster `rc-tke`, left menu **Configuration Management**, then **Secret**.
2. Set the namespace filter to `grc`. The Duckz secrets should disappear.
3. **Create**.
4. Name: `rc-tcr-key`.
5. Type: Docker registry credential (`kubernetes.io/dockerconfigjson`), not Opaque.
6. Image repository domain: `uswccr.ccs.tencentyun.com`.
7. Username: `200027379897`.
8. Password: the `tcrPass` value in `api/gofiber/cmd/tencent/tcr/sv/sharedEnv.bat`. Confirm it.
9. Create. The type column for this row should read `kubernetes.io/dockerconfigjson`. The domain is only `uswccr.ccs.tencentyun.com`. A pull failure may be reported as `not found` or unauthorized; check both the image name/tag and registry credentials.

## 7. Create the application Secret

Follow [the credential instructions](DEPLOY-CONFIG.md) to create one Opaque Secret, `game-api-env`, in namespace `grc`. Its core keys are `DB_PASSWORD`, `REDIS_PASSWORD`, `OCULUS_APPSECRET`, and `PHOTON_AUTH_PHASE`. Add `ADMIN_API_TOKEN` for the current token-based admin website and optionally `META_EDITOR_USER_IDS` for approved Editor users (organization-scoped IDs, comma-separated). Do not add Elasticsearch or JWT configuration. Keep the separate image-pull Secret from step 6.

## 8. Create the API deployment

If you already created an API deployment without these objects, edit that deployment instead of creating a second one. Saving it restarts the pods.

Create one API deployment. A second copy is not used.

1. Left menu **Workload**, **Deployment**.
2. Namespace `grc`. **Create**.
3. Name: `v1-api-deployment`. Replicas: use the existing Classic count for updates. For a first test deployment, start with 1 and size it for production after verification.
4. On the container. The form splits the image across two fields:
   - Name: `api`.
   - Image: `uswccr.ccs.tencentyun.com/gr-tcr/grc-api-golang`.
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
- **ImagePullBackOff**: inspect the pod events for image name/tag, registry credentials, and registry reachability. Confirm `rc-tcr-key` is in `grc` and selected on the workload.
- **CrashLoopBackOff**: inspect the application log. Missing build settings or credentials are reported by key name. Check MySQL connectivity/grants before rebuilding.
- **Running but not Ready**: inspect `/ready` and its dependency checks. Check the database and Redis errors returned by the readiness endpoint.

## 9. Create the admin deployment

1. **Workload**, **Deployment**, namespace `grc`, **Create**.
2. Name: `admin-deploy`. Replicas: 1.
3. Container name: `admin`.
4. Image: `uswccr.ccs.tencentyun.com/gr-tcr/grc-api-admin`. Image tag: `adm-grc-<git-hash>` from step 4.
5. Image pull policy: **Always**. Image access credential: `rc-tcr-key`.
6. Container port: `80`. Apache in that image listens on 80. The panel files are served under `/admin/`.
7. Do not attach `game-api-env` to the admin deployment. The admin image does not read it. Its API URL and Apache password file are baked in at build time. No custom runtime environment variables are required. Remove explicit `PATH` and `HTTPD_*` overrides; their defaults come from the Apache image. Do not remove inherited image defaults.
8. Liveness and readiness: **TCP port `80`**. `/admin/` is protected by Apache Basic authentication and normally returns 401 without credentials, so an unauthenticated HTTP probe there would fail. Verify the authenticated page manually after deployment.
9. Create.

## 10. Create the Services

A Deployment is not reachable until a Service points at it. This guide disables CLB-to-pod direct access in the ingress, so use NodePort backend Services for that mode. Do not create LoadBalancer Services as well; the ingress provisions the public balancer. [Tencent ingress troubleshooting](https://cloud.tencent.com/document/product/457/90686) documents the ClusterIP backend limitation.

1. Left menu **Services and Routes**, **Service**.
2. Namespace `grc`. **Create**.
3. API service. Name: `v1-api-deployment`. Type: **NodePort**. Select deployment `v1-api-deployment`. Map service port `80` to container port `3000`.
4. Create a second Service named `admin-deploy`. Type: **NodePort**. Select deployment `admin-deploy`. Map service port `80` to container port `80`. The admin container listens on 80, so the ingress rule for `/admin` uses service port 80.

Leave allocated node ports automatic. Confirm each Service selector finds the intended Ready pods, and that the node security groups allow the CLB traffic. Ingress rules still reference Service port `80`, not the allocated node port. If the live cluster uses a different/direct-access mode, verify its supported Service type before changing an existing working Service.

## 11. Upload a Cloudflare origin certificate

Duckz did not buy a certificate from Tencent. Its row in **My Certificates** is `cloudflare-dz.gunraidersapi.com`, source **Upload certificate**, issuer **CloudFlare Origin Certificate**. Classic uses the same kind of certificate. Cloudflare signs it, and you paste it into Tencent. There is no DNS validation step.

`gunraidersapi.com` is already a site in Cloudflare. Do not add the site again. Classic has one public hostname, `grc.gunraidersapi.com`, and one certificate for that name. The admin panel is a path on that same host, `https://grc.gunraidersapi.com/admin/`.

Create it in Cloudflare:

1. Cloudflare dashboard, zone **gunraidersapi.com**.
2. **SSL/TLS**, then **Origin Server**.
3. **Create Certificate**.
4. Private key: **Generate private key and CSR with Cloudflare**. Key type **RSA**.
5. Hostnames: remove the defaults (`gunraidersapi.com` and `*.gunraidersapi.com`). Add only `grc.gunraidersapi.com`.
6. Certificate validity: **15 years**.
7. **Create**.
8. Key format: **PEM**.
9. Copy the **Origin Certificate** and the **Private Key** into two text files on your machine. Cloudflare shows the private key on this screen only. Do not put either file in the git repo.

Upload it to Tencent:

1. Tencent Cloud Console, product **SSL Certificate Service**, page **My Certificates**.
2. **Upload Certificate**.
3. Standard: **International standard**.
4. Certificate name: `cloudflare-grc.gunraidersapi.com`.
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

The new row should look like the Duckz row: source **Upload certificate**, status **Issued**, and the domain list shows `grc.gunraidersapi.com` and **CloudFlare Origin Certificate**.

This certificate is trusted by Cloudflare, not by a browser that talks to the load balancer directly. The `grc` DNS record in step 13 has to stay proxied.

## 12. Create a public load balancer for Classic

Duckz already has a balancer. Leave it alone. This guide uses a separate Classic balancer to avoid changing Duckz routing. Tencent also supports explicitly configured shared-CLB ingress groups; that is a separate design, not something to enable by simply choosing the Duckz balancer. [Tencent shared-CLB documentation](https://intl.cloud.tencent.com/ind/document/product/457/78649).

You do not create the balancer in the Cloud Load Balancer product. Creating the Ingress below creates it.

1. Stay in cluster `rc-tke`. Left menu **Services and Routes**, then **Ingress**.
2. Set the namespace to `grc`. **Create**.
3. Ingress name: `grc`.
4. Namespace: `grc`.
5. Ingress type: the CLB type (the Tencent load balancer controller). If the form lists Nginx as well, do not pick Nginx.
6. Network type: **Public network**.
7. Load balancer: **Automatically create**. Do not choose the Duckz balancer.
8. IP version: **IPv4**. VPC: **Current VPC**. Carrier: **BGP**. Leave billing and instance size on the form defaults.
9. Match the Duckz load balancer options on this same form:
   - **Network mode**: **Disable CLB-To-Pod Direct Access**.
   - **Bypassing by Default**: leave **Enable synchronization of Bypass Backend Security Group configurations** unchecked.
   - **Security group**: leave **Enable synchronization of security group configurations** unchecked.
   - **Redirect**: **Custom**. Leave the redirect list empty. Duckz does not send one path to another.
   - **Custom Listening Port**: checked. Add **HTTP** port **80** and **HTTPS** port **443**.
10. Forwarding configuration. On the **HTTPS** **443** listener, the orange certificate warning is pointing at the **TLS configuration** block under the rules. The certificate is not in the forwarding-rule dropdowns. Those dropdowns are the backend Service and its port, and the port stays empty until a Service is selected.
   In **TLS configuration**, the Secret list says **No data yet** until you create one. Click **create a new one**. On that page, select the uploaded certificate `cloudflare-grc.gunraidersapi.com` and create the Secret. TKE stores the certificate id in that Secret. It does not ask you to paste the key again. Return to the Ingress form, click the refresh icon beside **Secret**, and choose the Secret you just created. Set **Domain** to `grc.gunraidersapi.com` and check **Default certificate**. Configure the certificate here. A certificate added later on the load balancer page is overwritten by this ingress. Add two rules on that same domain:
   - Domain `grc.gunraidersapi.com`, path `/admin`, backend the `admin-deploy` service, service port `80`.
   - Domain `grc.gunraidersapi.com`, path `/`, backend the `v1-api-deployment` service, service port `80`.
   The port in each rule is the Service port. The admin container listens on 80, so the `/admin` rule uses service port 80. The `/admin` rule is the more specific path, so the panel is not sent to the API.
11. On the **HTTP** **80** listener, add the same two rules and no certificate, so a plain `http://` request still reaches the services.
12. **Create Ingress**.

The new row shows a load balancer id. Click it. That opens the new Classic balancer in the load balancer console, in West US (Silicon Valley). Copy the address shown there. A domain-name balancer shows a domain and no IP. An older balancer shows a fixed IP. Write down whichever one is on the page. That address belongs to Classic. It will not match the Duckz balancer.

## 13. Cloudflare and Photon

These are not Tencent pages. They are required before the game can call the API.

Cloudflare, zone `gunraidersapi.com`:

1. Add one record for host `grc`. If step 12 gave you a domain, the type is **CNAME** and the target is that domain. If step 12 gave you an IP, the type is **A** and the target is that IP. The ingress sends `/` to the API and `/admin` to the admin panel.
2. Proxy status: **Proxied** (orange cloud). The origin certificate from step 11 is only trusted by Cloudflare. A grey-cloud record makes the browser talk to Tencent directly and reject the certificate.
3. SSL/TLS mode **Full (strict)** for this hostname. If changing the zone-wide mode, first confirm the other hosts in the zone support it; use a hostname-specific configuration rule when needed. Cloudflare speaks HTTPS to the balancer and checks the origin certificate. [Cloudflare Origin CA guidance](https://developers.cloudflare.com/ssl/origin-configuration/origin-ca/).

Photon dashboard, for the Classic app:

1. Custom authentication URL: `https://grc.gunraidersapi.com/api/photon/authenticate`.
2. Add a Custom Server authentication provider. Add a dashboard query parameter named `secret` whose value exactly matches `PHOTON_AUTH_PHASE` in `game-api-env`.
3. Reject authentication when the provider is unavailable, and disable anonymous connections. Allow dashboard changes time to propagate.
4. The Unity client sends `usr`, `pla`, and its current session `tkn`. Do not put the shared phrase into Unity: the old `photonAuthSecret` field was removed. The dashboard supplies it server-side.
5. Configure the client with this game's Photon Realtime App ID. Voice is a separate Photon app; check its authentication requirements separately if enabling authentication for Voice. Keep `ENABLE_PHOTON_AUTH=1` for this setup.

## 14. Check

1. Pods belonging to `v1-api-deployment` are Running and Ready in namespace `grc`.
2. `https://grc.gunraidersapi.com/api/v1/info/all` returns 200 and a JSON body with `data.version`.
3. The admin pod is Running, and `https://grc.gunraidersapi.com/admin/` prompts for Basic authentication and then loads the admin frontend. The website sends its existing Authorization token; it must match `ADMIN_API_TOKEN` on the API. No second login screen is needed for this mode. Verify an admin API request succeeds, not just that the website loads.
4. Test Quest sign-in using a build installed from the correct Meta release channel. Confirm the Unity Meta App ID matches the dashboard, the account has access, and User ID/User Profile grants are active in Data Use Checkup. For Editor testing, configure the organization-scoped allowlist below. Confirm a row appears in this game's `players` table.
5. Verify a write followed by a read in the game and a reconnect/session refresh. Test Photon authentication with the matching client.
6. Confirm no old ConfigMap values or `_RO`/`ES_*` settings are needed. Attestation-history endpoints now return 501 by design; remote Meta verification and bans remain available.

## Updating an existing Classic deployment

1. Do not recreate the database, namespace, Services, certificate, DNS record, or ingress. Record the currently deployed image tags and environment references for rollback.
2. Update public build settings before building. Keep the existing `REDIS_NAMESPACE` when updating a game. Preserve the existing game/Photon credentials unless deliberately rotating them.
3. Create or update `game-api-env` in `grc`. Updating a Secret alone does not refresh a running container's environment.
4. Build/push a new API image using section 4. Build the admin image only if its code, baked API URL, or Basic-auth file changed.
5. Edit the existing API Deployment: select the new image tag, import `game-api-env`, and remove superseded configuration imports after comparing their values with the new build settings. Saving a pod-template change triggers a rollout; if only the Secret changed, explicitly restart the Deployment.
6. Verify section 14 before removing old configuration objects or retiring a second database account. Keep them through the rollback window; restoring an old image may also require restoring its old configuration references.

Only one application Secret is needed. The image-pull Secret and certificate-reference Secret still serve separate purposes and remain in place.

For the current Apache-password plus static-token admin setup, and the alternative player-session permissions, see [ADMIN-SETUP.md](ADMIN-SETUP.md).

### Unity Editor sign-in

Quest builds continue to use device attestation. Editor requests can instead authenticate with a Meta user proof, but only for accounts explicitly allowed on the backend.

1. Deploy the backend version containing `meta_editor_login.go`.
2. In Unity, use **Meta > Platform > Edit Settings** to sign in the test user for the GRC Meta app.
3. Enter Play mode. The Console prints `[SignIn] Editor Meta organization-scoped user ID: ...`.
4. Add `META_EDITOR_USER_IDS` to the existing Tencent `game-api-env` Secret, with that organization-scoped ID as its value. Multiple IDs are comma-separated. Also add the setting to your local `runtime.env` if you maintain it as the source for the Secret.
5. Restart/roll out the API deployment so the new environment value takes effect, then retry Play mode.

This setting is optional and disabled when empty. Use the organization-scoped ID from the log, which is also the player ID. Existing app-scoped allowlist entries must be replaced when deploying this version. No Editor bypass password is required. The backend verifies the proof with Meta and verifies the matching organization-scoped identity before issuing the normal session. Normal player bans still apply; device attestation is not available in the Editor. Remove an ID to prevent future Editor logins; existing sessions retain their normal lifetime.

## Resources to keep after migration

Classic's working namespace retains these Secrets:

- `game-api-env`: API runtime settings, imported only by the API workload.
- `rc-tcr-key`: private registry credentials, referenced by API and admin image-pull settings.
- `grc-buszzwdc`: Classic's current ingress certificate reference; contains `qcloud_cert_id`. A new game's generated name will differ. Select it in ingress TLS configuration and verify `spec.tls[].secretName`.

Keep the Kubernetes-managed `kube-root-ca.crt` ConfigMap if present or recreated. Old application ConfigMaps (`photon-id`, `photon-session-check`, `remote-api-config`, `shared-cache`, `shared-database`) and old credential Secrets are not required by this profile. Before deleting, check every workload's environment/volume references, image-pull references, service accounts, and ingress TLS references in that namespace. Two Secrets containing `qcloud_cert_id` are not necessarily redundant; retain each referenced one. Verify replacement pods start before retiring old objects.

## Troubleshooting this setup

- `/live` and `/ready` returning HTTP 200 with `{}` is normal. If API endpoints work but health URLs time out externally, check that ingress forwards `/` to the API, not only `/api/v1`. Photon also uses the unversioned `/api/photon/authenticate` path.
- Ingress uses the Service's exposed port, not necessarily the container port. Existing Classic may expose API service port 3000; the example above uses service port 80 mapped to container port 3000. Inspect the actual Service and match it rather than changing a working mapping blindly.
- Editor `missing authorization header`: verify the running image contains Editor login support and the request includes `X-GRC-Editor-Login: 1`. An unapproved organization-scoped ID produces a different error. Secret changes require new pods.
- Admin requests to `classic-api.example.com`: fill both URLs in `admin/.env.production`, rebuild/push the admin image, roll out the admin deployment, then hard-refresh the browser. Workload environment variables do not change bundled JavaScript.
- Admin 401: verify `ADMIN_API_TOKEN` is loaded in the API pod and exactly matches the browser's Authorization value (without `Bearer ` in the Secret). Clear stale `acc_token` cookies if the frontend token changed. A valid player session lacking admin privileges returns 403 instead.
- An empty skin-style table now returns HTTP 200 with `data: []`. The previous `404 / no row found` behavior requires an API image update.
- Player IP storage now keeps one validated, normalized address, not two comma-separated forwarded addresses. This fixes `1406 data too long for column ip` without changing the 45-character schema.
- No public ban-list download remains. Use the authenticated per-player `/api/v1/serviceban/get` endpoint. `HASH_BANALL` and `ENABLE_SWAGGER` are removed.

## Unity client integration notes

Set GameManager's `apiBaseUrl` to the game's versioned API URL; configure Meta and Photon App IDs for that game. Editor sign-in uses Meta test-user settings and a fresh user proof; the backend verifies the app-scoped proof and resolves the organization-scoped identity before checking `META_EDITOR_USER_IDS`. The app-scoped ID is still used internally, but the configured allowlist and player identity use organization-scoped IDs.

The Classic client logs `[SignIn]` and `[PhotonAuth]`, times out stalled sign-in requests, refreshes its session before Photon connection, and stops connecting on refresh failure. The Photon disconnect guard checks the network peer state; it no longer waits indefinitely on a fresh client's high-level Disconnecting state. Verify `Signed in`, `Session refreshed`, `Connected to Master`, then `Joined lobby`.

Classic's Unity build window is under **Tools > GRC > Build > Open Build Window**. It uses one password for its keystore and key alias. Its local `grc-release.keystore` and `grc-signing-password.local.txt` are excluded from Git and must be securely backed up. A new game needs its own reviewed bundle identifier/signing setup; do not replace the key of an already-published app. `Patch-GammaLighting.ps1` in the Unity repo reapplies the intentional Meta/OpenXR Gamma validation overrides after package-cache refreshes. These client build details do not require extra API Secrets.

### Database feature cleanup

The Classic schema keeps accounts, admin privileges, service/IP bans, build hashes, SQL session history, skins, owned inventory, and Meta purchase records. News and cloud player-prefab endpoints and their admin UI have been removed. Legacy friends, match statistics, leaderboards, battle passes, quests, bounties, and XP tables are not used by this backend. Deploy the updated API and admin before removing retired tables. Keep the existing `reports` table: account migration still updates its reporter and reported player IDs. Its legacy CREATE TABLE definition must be carried into a new game database separately; `structure_classic.sql` does not currently define it. Do not drop it or substitute an invented schema.

Redis stores live sessions and refresh tokens; `platform_session_entries` retains SQL session records. Keep both. Add `ip_blacklists` from the schema if missing: otherwise IP checks log a database error and allow the request through.

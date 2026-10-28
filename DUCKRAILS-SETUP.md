# Duck Rails authentication deployment

This repository was copied from Gun Raiders Classic. Use this file for Duck Rails-specific values; DEPLOY-CONFIG.md describes the shared deployment mechanism. No cloud resources or credentials have been provisioned by the compatibility changes.

## Configured public values

- Client API base: https://duckrails.gunraidersapi.com/api/v1/
- Meta app: 28912202281742343 (confirmed by Justin).
- Photon Realtime app: 36bef5d2-63ad-4567-948e-52e966d7f2e9 (confirmed by Justin).
- Photon Voice app: b7c80e26-6a61-45ca-97d9-478af736a6d3 (confirmed by Justin).
- Proposed provisioning names, now used by config/build.env: database duckrails_tencent_global, database account duckrailsuser, TKE namespace duckrails, Redis namespace duckrails.
- Image repositories: duckrails-api-golang and duckrails-api-admin. Provision them before using the push scripts.
- APP_VERSION is 0.2.7, matching the current Unity client. Build-hash checking remains disabled as in the template.

## Before first deployment

1. Provision the separate database/account and import api/sql/structure_classic.sql ONLY into Duck Rails. Despite its inherited filename, the script now creates/selects duckrails_tencent_global. No SQL has been executed. Retain the single shared database instance and existing Redis instance; verify their copied private addresses in config/build.env.
2. Supply this game's DB_PASSWORD, OCULUS_APPSECRET, PHOTON_AUTH_PHASE and the shared Redis account's REDIS_PASSWORD through the duckrails namespace's game-api-env Secret. The copied ignored runtime.env and local .env files have not been read, reused, or changed; replace any copied credentials with the correct values before use. Do not run the local API with Classic's copied .env.
3. For Editor login, add organization-scoped IDs to META_EDITOR_USER_IDS. The Unity Editor logs `[SignIn] Editor Meta organization-scoped user ID: ...`. Configure the Meta Platform Editor test account for the confirmed app. Empty allowlist disables Editor access; no bypass phrase is needed.
4. Configure Custom authentication on BOTH Photon apps with URL https://duckrails.gunraidersapi.com/api/photon/authenticate and a dashboard query parameter named secret equal to PHOTON_AUTH_PHASE. Disable anonymous clients for a secured deployment. The Unity client supplies usr, pla and tkn; the server-only secret must never be put into Unity. The Voice client copies PUN authentication, so both apps need the same backend provider configuration.
5. Set up the duckrails.gunraidersapi.com DNS/TLS/load balancer and deploy the API. Configure health checks using the existing deployment guide. Preserve /api/v1 routes and the unversioned /api/photon/authenticate callback.
6. The copied push shortcuts now target this repository, and push scripts use Duck Rails image repositories. Their inherited filenames still mention GRC. Admin API URLs now point to Duck Rails. Rotate/configure admin credentials separately before publishing the admin site; this work did not change authentication secrets.

## Protocol and evidence

Quest login posts multipart identity/proof fields to /api/v1/player/v2/social-account/oculus and sends device attestation as Authorization: Bearer <token>. Editor login instead uses X-Meta-Editor-Login: 1; the backend validates the Meta nonce, organization ID, and allowlist. The old Remote-Authorization-Skip header is no longer sent by Unity.

The existing registration response matches GameSessionClient.SessionData. POST /api/v1/session/{session_id} accepts refresh_token; the returned session_id is -1 and the backend intentionally does not use the path ID. Photon checks the access token in Redis and returns the composed player identity, matching gr_id.

Verified locally: middleware/config/handler tests pass, including valid/invalid Editor proofs and multipart Unity request/response compatibility; go build succeeds; the Unity-generated Assembly-CSharp.csproj builds with warnings and zero errors. Meta is stubbed in the protocol test. No live Meta, Redis, MySQL, Photon, headset sign-in or deployment was exercised.

After deployment, verify a fresh account and repeat login, allowed/disallowed Editor login, device attestation, token refresh, Realtime connection, Voice connection, room changes and disconnect/reconnect. Use two real accounts/headsets. The inherited additional-Editor local-session shortcut has no backend token and is not a supported client for secured Photon authentication.

## Separate integration still needed

Classic does not provide Duckz's /duckz wallet/store/IAP/loadout/admin-check routes or /player/cosmetics routes. The client starts those requests after login without awaiting them, so authentication can complete, but those features will fail until the existing Duckz implementations are integrated. This change does not claim commerce, cosmetic restoration, or the expedition session milestone is complete.

# Initial administrator setup

## Existing token-based admin access

The backend supports an optional `ADMIN_API_TOKEN` environment variable for the existing admin website. Add it to the API's `game-api-env` Secret with exactly the token the website sends in its Authorization header, then restart the API deployment. Do not use the Photon phrase or database password. No admin frontend rebuild or extra login page is required for this mode.

This token grants admin access only on registered API routes containing an `admin` path segment. It does not create a player session and does not work as a game/Photon token. Empty configuration disables this mode. The browser possesses this credential; anyone with a copy can call these admin API routes without the Apache password prompt. Rotate it by changing both the configured backend value and the website's token.

The session-based setup below is an alternative to static-token mode.

There are two separate access checks. Passing the Apache password prompt does not grant API administrator privileges.

## 1. Browser Basic authentication

`admin/misc/passwd` contains Apache password hashes and is copied into the admin image. Create it as described in DEPLOY.md, or preserve the existing file when updating Classic. Its username/password unlocks access to the admin website files. It is not a player account, database account, or API token.

## 2. API administrator permissions

1. Sign in to this game's backend normally using the intended administrator's game/platform account. This creates the player record.
2. In Tencent's database console, select this game's database. Identify and verify the exact `players.id` for that person. Use the stored ID, including any platform prefix; do not assume the raw Meta user ID or display name is the right value.
3. Use a database administrator account to run the following, replacing the database name and player ID. This grants permissions only if the player exists. It reactivates an existing privilege row instead of creating duplicates.

```sql
USE `grc_tencent_global`;
SET @admin_player_id = 'REPLACE_WITH_VERIFIED_PLAYERS_ID';

SELECT id, name, platform FROM players WHERE id = @admin_player_id;
-- Stop unless this returns exactly the intended person.

START TRANSACTION;
UPDATE admin_privileges AS a
JOIN players AS p ON p.id = a.player_id
SET a.deleted_at = NULL, a.updated_at = UTC_TIMESTAMP()
WHERE p.id = @admin_player_id;

INSERT INTO admin_privileges (player_id, created_at, updated_at, deleted_at)
SELECT p.id, UTC_TIMESTAMP(), UTC_TIMESTAMP(), NULL
FROM players AS p
WHERE p.id = @admin_player_id
  AND NOT EXISTS (
    SELECT 1 FROM admin_privileges AS a WHERE a.player_id = p.id
  );
COMMIT;

SELECT player_id, deleted_at FROM admin_privileges
WHERE player_id = @admin_player_id;
```

4. Authenticate with a valid session for that player. The backend checks the session's composed player ID against `admin_privileges` on protected admin routes. Verify an administrator route succeeds and a non-admin player's session is rejected. Do not enable development mode to create an administrator.

The legacy `api/sql/types/admin_privileges.sql` file contains specific historical identities, not a reusable new-game bootstrap. Do not import it blindly.

## Current frontend behavior and new-game token setup

The supported setup here uses the Apache password prompt as the only interactive login. The placeholder Vue login form and unfinished session-refresh code are not used by static-token mode. The backend does not need to create a development session or enable development mode. A static-token administrator does not need an `admin_privileges` row; that table is for the alternative player-session mode above.

For a new game:

1. Generate a random admin token using the PowerShell generator in DEPLOY-CONFIG.md. Use a different value from all other credentials and other games.
2. In `admin/src/stores/auth.js`, replace the literal passed to `tmpLoginForever(...)` inside `ensureAuthenticated()` with this game's token. This frontend currently has no environment setting for that value. Treat the repository/build as containing the admin credential; never use the token for anything else.
3. Set the same raw value as `ADMIN_API_TOKEN` in the API's `game-api-env` Secret and local `runtime.env`. Do not include a Bearer prefix or quotation marks.
4. Fill `admin/.env.production` with the real versioned API URL in both URL fields, plus the game's namespace/prefab ID. The placeholder example hostname is not deployable.
5. Create/preserve `admin/misc/passwd`, build/push the admin image, and select its tag in the admin Deployment. Roll out the API Deployment after setting its Secret. Do not attach the API Secret to the admin pod.
6. Sign in through the Apache prompt. Verify requests such as `/api/v1/item/admin/all` succeed. The browser sends the token in Authorization; Apache credentials alone do not authorize these direct API requests.

When rotating the static token, update both ends, rebuild the frontend, and clear the old `acc_token` cookie on this hostname. Refreshing the page alone may keep a stale cookie because the auth store prefers it over the bundled fallback. The new token is accepted only on registered admin paths, including nested route groups; it is not a game or Photon session. Everyone using the shared token has the same admin authority, rather than a separate player identity.

Keep the Basic-auth file and the API Secret distinct. Never copy database credentials, Meta App Secret, or Photon shared phrase into the admin frontend.

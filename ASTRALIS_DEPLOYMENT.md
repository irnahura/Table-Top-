# Astralis deployment status

Requested destination: `https://vercel.com/irnahuras-projects`.

## Verified on 25 November 2026

- The local Vercel CLI is authenticated as `dexlabsaieducation-2232`.
- `vercel teams ls` lists only `dexlabsaieducation-2232s-projects`.
- `vercel project ls --scope irnahuras-projects --filter astralis --format=json` returned `The specified scope does not exist`. This login cannot access the requested destination.
- No Vercel project was created or linked and no deployment was started. Do not deploy into the other account as a fallback.
- The latest application work passed Go tests/vet and eleven browser journeys before this deployment request. Deployment/account verification is separate from those local checks.

## Required before publishing the complete application

1. Authenticate to Vercel with an account that can access `irnahuras-projects`, or grant the current account membership there. The Vercel integration is an alternative after its account connection is confirmed.
2. Choose the backend hosting/persistence arrangement. Astralis currently runs one Go HTTP/WebSocket server and saves its JSON database to disk. Uploading static files alone does not deploy authentication, campaigns or multiplayer.
3. Keep `data/astralis.json`, credentials, runtime logs, browser artifacts and local Vercel metadata out of Git and deployment uploads. Local saved accounts/campaigns are private and must not be bundled as production seed data.

Vercel Functions provide read-only deployed files and temporary scratch storage, so the current JSON file is not durable storage there. Vercel supports WebSockets as of its October 2026 announcement; the remaining architecture requirement is durable shared state and a compatible running backend, not a blanket absence of WebSocket support. Vercel Daemons currently provide background workers with ephemeral storage; Drives are mounted in Sandboxes rather than ordinary application Functions.

References: [Vercel runtime filesystem](https://vercel.com/docs/functions/runtimes), [WebSocket support](https://vercel.com/changelog/websocket-support-is-now-generally-available), [Daemons](https://vercel.com/docs/daemons), [Drives](https://vercel.com/docs/drives).

No production deployment is claimed in this record. Recheck account access and hosting configuration before continuing.

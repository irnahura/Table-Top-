# Astralis deployment status

Requested destination: `https://vercel.com/irnahuras-projects`.

## Account verified on 25 November 2026

- The initial CLI login was `dexlabsaieducation-2232`, which could not access the requested team. No project was created in that other account.
- Completed Vercel sign-in as **irnahura**. `vercel whoami`, team listing and project lookup confirmed access to **irnahuras-projects**.
- Created and linked **irnahuras-projects/astralis-tabletop**, with the existing GitHub repository connected. Local `.vercel/` metadata stays ignored.
- The latest application work passed Go tests/vet and eleven browser journeys before this deployment request. Deployment/account verification is separate from those local checks.

## Required before publishing the complete application

1. Account access is now confirmed. Continue targeting `irnahuras-projects` explicitly.
2. Choose the backend hosting/persistence arrangement. Astralis currently runs one Go HTTP/WebSocket server and saves its JSON database to disk. Uploading static files alone does not deploy authentication, campaigns or multiplayer.
3. Keep `data/astralis.json`, credentials, runtime logs, browser artifacts and local Vercel metadata out of Git and deployment uploads. Local saved accounts/campaigns are private and must not be bundled as production seed data.

Vercel Functions provide read-only deployed files and temporary scratch storage, so the current JSON file is not durable storage there. Vercel supports WebSockets as of its October 2026 announcement; the remaining architecture requirement is durable shared state and a compatible running backend, not a blanket absence of WebSocket support. Vercel Daemons currently provide background workers with ephemeral storage; Drives are mounted in Sandboxes rather than ordinary application Functions.

References: [Vercel runtime filesystem](https://vercel.com/docs/functions/runtimes), [WebSocket support](https://vercel.com/changelog/websocket-support-is-now-generally-available), [Daemons](https://vercel.com/docs/daemons), [Drives](https://vercel.com/docs/drives).

## Publication configuration

`vercel.json` builds the current framework-free frontend using `node scripts/build-vercel.cjs` and publishes only `dist/`. The installer is intentionally disabled; the legacy Python requirements/backend and design archive are excluded by `.vercelignore`, along with saved data, credentials and runtime output. `scripts/check-vercel.cjs` passed desktop/mobile checks of the built preview, course and online-play availability notice, with no accidental backend API calls or saved data included.

This deployment is a **public website/interactive preview and guides**, not the persistent game backend. The generated `deployment-config.js` enables an explicit online-play availability page for sign-in/campaign/table routes. Running the Go app locally keeps full existing functionality because that generated configuration is only injected into the Vercel build. Do not remove the availability guard until a persistent backend is configured and tested.

The first application commit is `b319aa9` (`feat: add adventure course and linked character sheets`). Vercel publication configuration is committed separately. Production deployment URL/status will be verified after upload.

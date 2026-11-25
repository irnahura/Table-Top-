# Astralis

Astralis is a browser-based virtual tabletop for RPG campaigns. A Go HTTP/WebSocket server validates gameplay and saves accounts, campaigns and sessions. The frontend uses plain HTML, CSS and JavaScript. The legacy Python backend is optional historical code, not the current game's authoritative runtime.

This README is the project's **single tracked Markdown reference**. Older plans, handoffs, audits and progress notes remain locally but are no longer published in the current GitHub tree. Historical commits are not rewritten.

## Current hosting

- Website: **https://astralis-tabletop.vercel.app**.
- Vercel project: **irnahuras-projects/astralis-tabletop**.
- The public site includes the landing page, interactive preview, handbook and course. **The persistent Go backend is not hosted yet; online sign-in and saved campaigns remain unavailable there.**
- Local full application: **http://localhost:8080/** when the Go server is running.
- Repository: https://github.com/irnahura/Table-Top- .

The Vercel build intentionally shows an availability notice instead of submitting account/campaign requests to a nonexistent backend. Local Go operation retains the full current feature set.

## Run locally

Requires Go 1.26 or a Go installation with automatic toolchain download enabled. Run from the project root:

```powershell
go run ./cmd/astralis
```

Open http://localhost:8080/ . `/health` returns `astralis-go / ok`. `PORT` defaults to `8080`; `ASTRALIS_DATA_FILE` defaults to `data/astralis.json`. The server listens on all interfaces at its configured port and serves the frontend itself.

For practice, open Sign in → **Try a demo table**. Each entry creates a separate saved GM sandbox. Keep the browser signed in to return; demo accounts have no password recovery. For continuing play, register an account. Players register separately and redeem the GM's invitation code from My campaigns.

## Deploy the Go backend on Render

The current architecture needs **one Go web service and a persistent disk**. Render supports Go services and WebSockets, and its disks preserve files across restarts/deploys. Persistent disks require a paid service. Review current charges before creating resources; this repository setup has not created a paid service.

`render.yaml` is a ready-to-review Blueprint: Go runtime, one `0.5c-512mb` instance, 1 GB disk at `/var/data`, `/health`, and the settings below. It provisions nothing until applied in Render.

1. Sign in to https://dashboard.render.com/ and connect this GitHub repository. Create a Blueprint from `render.yaml`, or manually create a Web Service.
2. Select branch `main`, repository root as the working directory, and language/runtime **Go**.
3. Configure:

   | Setting | Value |
   | --- | --- |
   | Build command | `go build -tags netgo -ldflags '-s -w' -o astralis-server ./cmd/astralis` |
   | Start command | `./astralis-server` |
   | Health check | `/health` |
   | `GOTOOLCHAIN` | `auto` |
   | Disk mount path | `/var/data` |
   | `ASTRALIS_DATA_FILE` | `/var/data/astralis.json` |
   | Instances | `1` |

4. Let Render supply `PORT`. Choose a paid service with its persistent disk attached and confirm costs before creating it.
5. Deploy and verify `https://<service>.onrender.com/health`. Open the root URL: **the Go service already serves the full frontend**, so accounts, campaigns and WebSockets can operate on that same Render origin without cross-origin integration.
6. Test with disposable accounts: create a campaign, invite another account, assign a token, exchange live messages, restart the service and verify persistence. Begin production with an empty database; do not upload personal local accounts as seed data.
7. After the backend is healthy, connect the Vercel experience. That integration is **not configured yet**. Initially, gameplay can use the Render URL. To run the table under the Vercel origin, implement/test an explicit API origin or proxy and exact allowed WebSocket/CORS origins before removing the availability guard. Setting an environment variable alone in the current build does not connect it.

The JSON store and in-process live hub require one instance. Do not scale it horizontally with independent files or multiple writers. Back up `/var/data/astralis.json`. PostgreSQL/object storage and transactional save-error handling remain future work. Disk-backed rollouts can briefly interrupt service.

Official references: [Go deployment](https://render.com/docs/deploy-go-gin), [persistent disks](https://render.com/docs/disks), [WebSockets](https://render.com/docs/websocket), [Blueprint specification](https://render.com/docs/blueprint-spec).

## Vercel website publication

`vercel.json` uses the **Other** framework preset, skips dependency installation, runs `node scripts/build-vercel.cjs` and publishes only `dist/`. The build injects preview-only `deployment-config.js`; `deployment.json` reports that the persistent backend is not connected.

```powershell
node scripts/build-vercel.cjs
vercel --prod --yes --scope irnahuras-projects
```

Use an account with access to `irnahuras-projects`. `.vercelignore` excludes local data, environment files, generated output, legacy backend files and the design archive. Local saved users/tokens must never be bundled or published.

The verified production release returned HTTP 200 for the site/open rules PDF, matched the source guide asset, and returned HTTP 404 for `/data/astralis.json`. Live browser checks covered the course, mobile layout and availability notice. Application/configuration commits: `b319aa9`, `fcd9b08`; publication record: `e7ecfa5`. The remote branch history remains available for older implementation checkpoints.

## Guides and preparation

- `/#/help`: 32 searchable guide entries, player/GM filters and direct topic links.
- `/#/help/adventure`: eleven lessons, a desktop chapter list, mobile picker and local progress.
- `/#/help/adventure/resources`: seven original templates for outlines, locations, combat, challenges, rewards, briefings and session-zero agreements. Copy/download or stage a private Journal draft in a GM table and explicitly save after review. Existing drafts are preserved.
- `/#/getting-started` and `/#/guide`: first-session introduction and RPG glossary.

At a working table, **Help & tools**, **?**, **Ctrl/Command + K**, contextual help and **Open this tool** locate controls. The six-step tour changes local panels without issuing saved commands. Course/checklist completion is manually marked local browser state, not automatic campaign readiness.

## Features and boundaries

| Area | Implemented | Boundary |
| --- | --- | --- |
| Accounts/campaigns | Registration, sign-in/out, separate demos, library, launch/resume and expiring/limited/revocable invitations. | No recovery, email verification, public matchmaking or membership removal. |
| Multiplayer | Authenticated WebSockets, presence, replay/reconnect, deduplicated commands, version conflicts and polling fallback. | One Go process. |
| Control/combat | GM assignment, server-checked player actions, terrain/path/occupancy and movement budgets, basic attacks/damage, party turns and rounds. | Astralis basic rules; hostile NPCs stay GM-controlled; no hostile AI or full published-system combat. |
| Maps/scenes | Image import, backgrounds, duplicate scenes, whole-table switching, grid scale/metric, crop and quarter-turn rotation. | No interactive image/grid alignment, hex grids or independent GM/player scenes. |
| Editor | Pan/zoom/fit/grid, grouped tools, freehand/shapes/text/pins, layers, gesture previews, dragging and movement animation. | No drawing selection/resizing or ruler waypoints. |
| Tokens/initiative | HP, markers, rotation, size, locks, artwork, duplication/deletion, party initiative and tracker sorting/reordering. | No custom/NPC initiative entries; markers do not apply rule effects. |
| Sheets/resources | Assigned private/shared sheets, inventory/spells/resources, tracked quantities and validated preparation/use. | Manual notes and effects; no automatic combat-stat conversion. |
| Sheet/token links | GM links/unlinks actors or places tokens from sheets; token menus open authorized linked sheets. Links survive scenes/restarts; deletion clears links. | New party tokens inherit control once. Existing HP/control remain separate; sheet reassignment does not reassign tokens. |
| Visibility | Assigned living-token sight, walls/doors/windows, darkness, lights, darkvision and masked custom maps. | Shared fog planning is not concealment. No manual reveal mask, explored-map memory or blended light. |
| Story tools | Private/shared text handouts, notes, quests/objectives and weighted random tables. | No folders/attachments, automatic traps/puzzles or loot distribution. |
| Communication/dice | Character/account chat, whispers/emotes, public/GM/self-only dice and private macro definitions. | Simple formulas; sheet and macro rolls are public. |
| Music/calling | Shared audio intent with listener opt-in/volume; direct WebRTC voice/video. | No TURN relay; internet-wide/real-device reliability is not established. |
| Game references | Basic, D&D 2014/2024, Warhammer, Pathfinder, Fate and percentile profiles, check helpers, templates and book links. | Not a full system simulator or structured compendium. |
| Ruler/FX | Line/square/circle/cone/beam measurement, snapping/free points, sharing and transient visual effects. | No automatic costs, damage, spells or programmable effects. |

The written [Roll20 adventure-creation series](https://help.roll20.net/hc/en-us/articles/29257395844247-Creating-your-adventure-Overview), eleven chapters and resources informed the course. Astralis supports manual adventure preparation, not complete competitor parity. The Riverside story/worksheets are original; embedded-video playback was not verified.

## Architecture and source map

The authoritative path is `frontend → authenticated Go API → validated command → saved state/event → filtered live updates`. The server owns positions, dice, HP/damage, turns, permissions and durable changes. The client owns view transforms and temporary previews. Future AI narration must remain non-authoritative and respect information access.

Passwords use salted PBKDF2-SHA256; seven-day bearer tokens survive restarts. Common account-specific projection filters snapshots, replay, live sync, listings and conflict/duplicate responses. GM objects, private content, self-only rolls and inaccessible sheet references are filtered before delivery; hidden-token link/placement history remains GM-only.

JSON persistence uses a temporary file and rename under a process mutex. Save failures are logged rather than transactionally returned through every endpoint. Inline images/retained snapshots limit scale. Production rate limits, account lifecycle, role management, storage migration and observability remain open.

| Location | Responsibility |
| --- | --- |
| `cmd/astralis/main.go`, `tools.go` | Server, persistence, HTTP API and core gameplay/tools. |
| `live.go`, `ownership.go` | Live sessions, receipts/replay and actor authority. |
| `table.go`, `sight.go`, `environment.go` | Workspace, common privacy projection, maps and visibility. |
| `adventure.go`, `initiative.go`, `systems.go` | Sheets/links, quests/tables, tracker and game references. |
| `frontend/app.js` | Routes, authentication/API handling and main UI. |
| `tabletop.js`, `canvas.js`, `editor.js`, `systems.js` | Frontend workspace/canvas/specialist tools. |
| `learning.js`, `adventure-guide.js` | Handbook, contextual help, course and original templates. |
| `scripts/` | Contracts, disposable browser journeys and Vercel build. |
| `frontend/rules/` | Open PDFs, attribution and source/checksum manifest. |
| `Astralis-Game-Table/` | Historical prototype archive, excluded from Vercel output. |

The accepted current design is map-focused: dark neutrals, Manrope, warm gold actions, cyan movement/system feedback and coral combat. Older light/isometric/cinematic concepts are historical. Preserve keyboard access, reduced-motion behavior and the usable 390px layout; do not reintroduce the superseded portrait/WebGL backdrop without a new request.

## Verification

```powershell
go test ./...
go vet ./...
node scripts/check-frontend.cjs
go build -o output/astralis-test.exe ./cmd/astralis
```

Browser journeys require Playwright. Set `PLAYWRIGHT_MODULE` if an existing installation is not resolvable as `playwright`; verify the package path before using it. Journeys launch isolated browsers and temporary databases, not saved `data/astralis.json`.

| Script | Coverage / port |
| --- | --- |
| `check-table-tools.cjs` | Tools, invitations, permissions and persistence; 18081. |
| `check-live.cjs` | Live updates, retry, reconnect and polling; 18082. |
| `check-character-control.cjs` | Assignment, forged actions and persistence; 18083. |
| `check-table-workspace.cjs` | Scenes, privacy, art, sight and synthetic media; 18084. |
| `check-adventure.cjs` | Sheets/resources, environment, quests, tables and invites; 18085. |
| `check-canvas.cjs` | Gestures, freehand, pan/zoom and turns; 18086. |
| `check-editor.cjs` | Context menus, tools, scenes and tracker; 18087. |
| `check-systems.cjs` | Profiles, private dice, rulers/FX, books and lighting; 18088. |
| `check-learning.cjs` | Handbook, tool targets, checklists, tour and roles; 18089. |
| `check-adventure-course.cjs` | Course, templates, links, draft safety and persistence; 18090. |
| `check-vercel.cjs` | Built preview/availability and no accidental API calls; 18091. |
| `check-cinematic.cjs` | Landing/auth/campaign/map/mobile regression with API fixtures; defaults to localhost:8080. |

Go tests/vet and all eleven gameplay/help journeys passed for the adventure-course continuation. Later Vercel build/production checks also passed. Earlier suites were not rerun solely for documentation consolidation. No successful race-detector run is claimed; CGO was disabled in the recorded environment. Keep screenshots and test output under ignored `output/`.

## Content provenance

The unchanged publisher D&D SRD 5.1/5.2.1 PDFs are open references under CC BY 4.0. Legal pages, required attribution and checksums remain in `frontend/rules/ATTRIBUTION.txt` and `manifest.json`. Other rulebooks link to publisher or publisher-endorsed sources; commercial books/codexes are not bundled as supported compendiums. Icon attribution/license files remain in `frontend/assets/`.

Historical covers in `Astralis-Game-Table/assets/alchemy-covers/` were localized from the [Alchemy RPG catalog](https://alchemyrpg.com/) for prototype reference: `wilderfeast.webp` (Wilderfeast), `dungeons-and-dragons.webp` (Dungeons & Dragons), `call-of-cthulhu.webp` (Call of Cthulhu), `vaesen.webp` (Vaesen), `candela-obscura.webp` (Candela Obscura), and `vampire-the-masquerade.webp` (Vampire: The Masquerade). Confirm production distribution rights before publishing these assets. The archive remains excluded from Vercel deployment.

## Next priorities

Deploy/verify the persistent backend and connect Vercel. Further work includes structured system fields/explicit stat synchronization, grid alignment, open-license compendium imports, NPC/custom initiative, deeper rules automation, independent GM scene preparation and stronger persistence/auth foundations. Preserve unrelated edits and local saved campaigns.

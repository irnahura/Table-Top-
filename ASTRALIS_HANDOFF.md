# Astralis project handoff

Latest continuation: [25 November 2026 guide and usability checkpoint](ASTRALIS_PROGRESS_2026-11-25.md) and [Roll20 adventure-workflow audit](ASTRALIS_ROLL20_ADVENTURE_AUDIT_2026-11-25.md). The public handbook, contextual help, adventure course, original preparation templates and session sheet/token links extend the feature inventory below. Read those documents for the latest changes and validation; game-rule and infrastructure limits below still apply.

Updated: 25 November 2026. Read the linked 25 November checkpoint for the newest guide/usability work; the final continuation below describes the preceding game-tool implementation. Read this file and `README.md` before changing the project. Earlier design/progress documents are historical references and may describe superseded implementations.

## Project and delivery state

- Workspace: `C:\Users\Mayank\OneDrive\Documents\GitHub\Table top`.
- Product: Astralis, a browser-based virtual tabletop for RPG campaigns.
- Runtime: Go 1.26 HTTP server with plain HTML, CSS and JavaScript. No frontend framework or bundler is required.
- Python code under `backend/` remains an older/optional implementation; it is not the current frontend's authoritative backend.
- Main page: <http://localhost:8080/>. The Go server was running and `/health` returned `astralis-go / ok` when this handoff was written. Recheck runtime health in a future session.
- No commit, push or deployment was made for this work.
- Working tree is already dirty. `README.md` is modified; `frontend/`, `cmd/`, `scripts/`, `data/`, `output/`, `go.mod`, the archive and other files are untracked. Numerous files under `design/` were already deleted before the latest feature work. Preserve these changes; do not reset, restore, delete or broadly stage unrelated files.
- User asked to save the work in Markdown for future context. This file documents the current implementation; older Markdown files have been retained.

## User direction and accepted design

The user rejected generic landing pages, abstract WebGL/particle backgrounds, generic AI copy, and the large cinematic female portrait. An earlier Detroit-inspired treatment was superseded by concrete virtual-tabletop product references. Do not reintroduce the portrait or WebGL backdrop without a new request.

The final direction is a map-focused tabletop workspace with a product-led landing page, inspired by useful patterns from Roll20, Alchemy, Foundry and Owlbear Rodeo. The user also supplied a Roll20 homepage screenshot. Keep Astralis's own assets and copy; do not claim competitor customer numbers, licensed game libraries or unsupported systems.

The `landing-page-design` skill was applied. Current styling uses Manrope, dark neutral surfaces, warm gold selection/actions, cyan movement/system indicators, and coral hostile/combat indicators. Keep responsive layouts, accessible names/states, reduced-motion support and a usable 390px mobile view. The map should remain the primary workspace.

OpenDesign Cloud MCP was attempted during the design work but returned `Transport closed`. No fresh successful design run should be claimed. The connected Browser runtime also reported no available browsers; isolated installed Playwright was used for testing, without personal browser profiles.

## Start and explore

```powershell
go run ./cmd/astralis
```

Default port is 8080. `PORT` overrides it. `ASTRALIS_DATA_FILE` overrides the database path; default is `data/astralis.json`. The Go service also serves `/design/` from `Astralis-Game-Table/` as an archive.

### Demo sign-in

1. Open <http://localhost:8080/#/login>.
2. Choose **Try a demo table**.
3. A separate saved GM sandbox opens directly in a playable Riverside encounter.

No dummy email/password is needed. Every demo entry creates a new account, campaign, map and session; visitors do not share one demo account. Reloads and backend restarts retain the current sandbox while its browser sign-in remains valid. Signing out and entering the demo again creates a new sandbox. There is no password-based recovery flow for a signed-out demo account.

For a second player, register a normal account in another isolated browser context, redeem the GM's invitation code in the campaign library, and resume the shared session.

## Completed interface work

### Landing and learning

- Product-led hero with an interactive miniature battle map and an overlapping read-only character card on desktop.
- Clear campaign CTA and actual feature benefits. No portrait or canvas backdrop.
- Accessible feature tabs for Battle maps, Characters, and Combat & log.
- Included Riverside starter encounter section with truthful initial counts: two party characters, two hostile NPCs, 20 × 14 grid.
- Learning resource cards, FAQ, final CTA, footer and local prototype privacy/terms notices.
- Learn menu with RPG glossary, getting-started guide, and preview walkthrough.
- Three-step public walkthrough: select Mayank, enable the tactical grid, select Bandit Scout. It can start from another page and return to the homepage preview.
- Seven-step getting-started guide with keyboard and mobile instructions.
- Public preview is illustrative and does not mutate a saved campaign.

### Campaign workspace

- Registration, sign-in, sign-out and demo entry.
- API-backed campaign creation and listing, search, name/date sorting, empty/loading/error states.
- Campaign lobby, launch and saved-session resume.
- Server-side session discovery in the library, rather than relying only on a device's saved session reference.
- Invitation creation by the GM and redemption from another user's library.

### Tabletop workspace

- Main battle map, scroll/drag panning, zoom, fit, grid toggle, sync, active turn and round display.
- Actions, Characters, Tools and Log inspector tabs with keyboard tab navigation.
- Selected character inspector, HP and combat stats, directional movement, hostile target selection, attacks and bridge investigation.
- Server-resolved attack roll cards and encounter history.
- Mobile inspector opens as a dismissible side panel; tool rail scrolls to expose all controls.
- Map view survives selection/actions. Form drafts survive renders and polling; polling avoids replacing an actively edited form.
- Shared snapshots arrive over authenticated WebSockets with automatic reconnect and sequence replay. Five-second polling is the fallback; focused forms defer incoming renders and retain drafts.

## Map tools and saved features

| Tool / feature | Current behavior |
| --- | --- |
| Select | Select party tokens or hostile targets; regular movement uses Actions |
| Pan | Drag the map viewport |
| Ruler | Two cell clicks; orthogonal grid-space distance; local transient overlay |
| Line / Area | Two clicks save a line or rectangular area annotation |
| Undo | Removes the last saved annotation |
| Block | Toggles blocked terrain; server movement checks it |
| Fog | Toggles a shared fog planning cell; does not hide information from players |
| Place | GM places the token selected in Tools, ignoring regular turn movement |
| Ping | Saves a shared map ping; the latest ping is shown on the map |
| Grid / Fit / Sync | Local grid/view controls and a server snapshot refresh |
| Scene background | Riverside artwork or blank grid, preserving tokens and edits |
| Map import | PNG/JPEG up to 5 MB and 20 megapixels; fitted to a 20 × 14 grid |
| Add token | Name, HP, coordinates and hostile flag; default basic combat stats |
| Edit token | Name and current HP, validated against maximum HP |
| Dice tray | Server rolls dice and bonus; result card and saved log entry |
| Session notes | GM edits shared plain-text session notes |
| Chat | Campaign members send saved messages |
| Invitations | GM creates a code; another account joins from its library |

Map annotations, terrain, fog plan, scene/image, tokens, HP, notes, chat and dice events persist through reloads and server restarts. The ruler and view transforms are UI state, not shared authoritative gameplay state.

Keyboard: `1` Select, `2` Pan, `G` grid, `+`/`-` zoom. Escape clears measurement/pending annotation selection, returns to Select, and dismisses mobile panels as appropriate. Inputs retain their normal typing behavior.

## Backend architecture and fixes

The current path is `frontend → Go HTTP API → validated game/tool command → persisted snapshot + event`.

- Persistence uses a local JSON database with maps for users, password hashes, campaigns, characters, maps, sessions, invites, tokens and token expirations.
- A database mutex protects access. Saving writes a temporary file and renames it over the database.
- Passwords use salted PBKDF2-SHA256, 600,000 iterations, 32-byte derived keys. Hashes are stored separately from public user JSON.
- Seven-day bearer tokens persist and survive server restarts; logout revokes the token.
- Fixed earlier password persistence: old `PasswordHash` was excluded from JSON. Accounts whose password was already lost cannot recover it from that old file; use the demo or a new account with another email. No password reset was implemented.
- Registration validates fields and rejects duplicate email addresses.
- Session reads/events/commands enforce campaign membership. Map, character and session creation check GM ownership. GM tool editing is restricted in the API as well as the UI. A joined player can chat, roll dice and ping.
- Rejected tool permissions return 403. Invalid gameplay/tool commands generally return 409 with a useful error.
- Fixed the `Point` JSON tags so both `x` and `y` serialize correctly.
- Movement validates integer map coordinates, living/active actor, blocked terrain, occupied cells, an available traversable route and remaining turn allowance. Accepted movement spends the path distance.
- One attack per turn; attacks use a server roll, range/turn checks, hit threshold and server damage.
- Turns cycle through living party tokens, reset movement/attack allowances, and increment rounds on wrap. Hostile NPCs remain GM controlled.
- Commands copy mutable state before validation so a rejection does not mutate the live actors/map slices.
- Dice input limits: 1–20 dice, 2–100 sides, bonus -100 to 100.
- Scene limits: 50 tokens, 200 annotations; notes 12,000 characters and chat messages 1,000 characters. Map uploads validate image data and size.

### API reference

All paths below have prefix `/api/v1`; authenticated requests send `Authorization: Bearer <token>`.

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/auth/register` | Register; returns user, then frontend separately logs in |
| POST | `/auth/login` | Returns token and user |
| POST | `/auth/demo` | Creates isolated demo and returns token, user, campaign, snapshot |
| GET / POST | `/auth/me` / `/auth/logout` | Current user / revoke token |
| GET / POST | `/campaigns` | List accessible campaigns / create campaign |
| POST | `/campaigns/{id}/invitations` | Create invitation |
| POST | `/invitations/redeem` | Join with `{ "token": "..." }` |
| POST | `/characters` | Create character record; token tools edit active scene actors |
| GET / POST | `/maps` | List accessible map records / create map |
| POST | `/campaigns/{id}/sessions` | Launch using `{ "map_id": "..." }` |
| GET | `/sessions` | Accessible sessions, newest first |
| GET | `/sessions/{id}/snapshot` | Direct session object with `state` and `events` |
| GET | `/sessions/{id}/events?after=0` | Events after a sequence number |
| POST | `/sessions/{id}/commands` | Submit `{ "payload": { "type": "...", ... } }` |

Session creation returns `{ session_id, snapshot }`; accepted commands return `{ snapshot, event_type, payload, sequence_number, state_version, status }`. Preserve these shapes; frontend tests cover them.

Gameplay command types: `character.move`, `combat.attack`, `object.interact`, `turn.end`.

Tool command types: `map.block`, `map.fog`, `map.ping`, `map.draw`, `map.undo`, `map.scene`, `map.upload`, `token.add`, `token.update`, `token.place`, `dice.roll`, `notes.save`, `chat.send`.

Example command payloads:

```json
{"payload":{"type":"map.draw","x":1,"y":8,"to_x":4,"to_y":10,"shape":"rectangle"}}
{"payload":{"type":"token.place","token_id":"lira","x":5,"y":8}}
{"payload":{"type":"dice.roll","count":3,"sides":6,"bonus":2}}
```

## Important files

| Path | Responsibility |
| --- | --- |
| `frontend/index.html` | Shell, font, CSS and script loading |
| `frontend/app.js` | Routes, markup, API/auth, map geometry, event handlers, tools and polling |
| `frontend/styles.css` | Base design, typography, forms and shared layouts |
| `frontend/vtt.css` | Product landing, campaign and tabletop workspace styles |
| `frontend/landing.css` | Hero layering, feature tabs, learning guide and walkthrough |
| `frontend/tools.css` | Tools panel, forms, map overlays, demo entry and responsive rules |
| `frontend/assets/riverside-battlemap.png` | Included village map artwork |
| `frontend/assets/ATTRIBUTION.txt`, `PHOSPHOR-LICENSE.txt` | Icon attribution/license |
| `cmd/astralis/main.go` | Server, DB, HTTP routes and gameplay rules |
| `cmd/astralis/tools.go` | Demo auth, password helpers, supplemental APIs, movement pathfinding, tool rules |
| `cmd/astralis/main_test.go`, `tools_test.go` | Real API, persistence, permissions and game rule tests |
| `scripts/check-frontend.cjs` | Dependency-free JS contract tests |
| `scripts/check-cinematic.cjs` | Historical filename; current desktop/mobile UI regression tests with API fixtures |
| `scripts/check-table-tools.cjs` | Real backend + isolated browser integration journey |
| `data/astralis.json` | Actual local saved campaigns/accounts; do not use for disposable tests |
| `output/screenshots/` | Current screenshots, including desktop/mobile tools and map |
| `output/references/` | Earlier design reference captures |
| `README.md` | Runtime and feature documentation |

There is no active `frontend/ambient.js` WebGL implementation. The old portrait asset may remain unused; retaining the file is not permission to put it back in the UI. Temporary editing scripts used during implementation were removed.

## Routes

`/#/` landing; `/#/login`; `/#/register`; `/#/library`; `/#/campaign/new`; `/#/campaign/{id}` lobby; `/#/table/{campaignId}` game; `/#/guide`; `/#/getting-started`.

Browser storage keys include `astralis-token`, `astralis-user`, and `astralis-session-{campaignId}`. Do not copy real saved tokens or database contents into reports.

## Validation and evidence

The following passed at completion of the feature work:

```powershell
go test ./...
go vet ./...
node scripts/check-frontend.cjs
go build -o output/astralis-test.exe ./cmd/astralis
```

Browser tests use installed Playwright. In this environment:

```powershell
$env:PLAYWRIGHT_MODULE = 'C:\Users\Mayank\AppData\Roaming\Python\Python314\site-packages\playwright\driver\package'
node scripts/check-cinematic.cjs
node scripts/check-table-tools.cjs
```

- `check-cinematic.cjs` passed for the existing landing/preview, registration, campaigns, search/sort, tabletop pan/zoom/fit, tabs, movement/combat/turn controls, and desktop/390px layouts. It calls the running local service for frontend assets but intercepts API calls with fixtures; it does not change saved game data.
- `check-table-tools.cjs` passed against a real server and temporary JSON database: demo entry, ruler, line/area, terrain, fog plan, ping, dice, token creation/edit/placement, map upload, scene switching, undo, notes, chat, invitations, player restrictions, shared polling updates, server restart persistence, turn progression and mobile tools.
- That test starts `output/astralis-test.exe` on port 18081 with disposable storage, uses separate browser contexts, restarts the test server, and cleans up its database. Build the executable first after Go changes.
- Go tests cover point serialization, account/campaign/session journey, permission rejection, saved tool state, password and token persistence, dice ranges, invalid commands, movement budget, turn rotation and rejected-command mutation.
- `go test -race ./...` could not run because CGO was disabled. Do not claim a successful race-detector run. Normal Go tests and vet passed.
- Desktop and mobile tools screenshots were visually inspected. Current examples: `astralis-tools-desktop.png`, `astralis-tools-mobile.png`, `astralis-map-tools-mobile.png`.
- This documentation-only turn rechecked `/health`; the feature tests above were completed in the preceding implementation turn rather than rerun just to write this handoff.

## Known boundaries and future work

- The implemented backend supports the current local app; it is not a claim that every commercial VTT feature is complete.
- Fog remains a shared planning overlay. Token line of sight, optional scene darkness, light sources, doors/windows and token darkvision are implemented. Explored-map memory and physically blended colored illumination remain absent.
- Flexible character sheets, tracked inventory, prepared spells/uses and resources are implemented. Licensed system sheets, automated spell effects, published spell libraries, compendiums and a separate handout-file manager remain absent.
- The basic ruleset is custom. Hostile AI and automatic NPC turns are not implemented; the GM can place/edit those tokens.
- Shared state uses WebSockets with a five-second polling fallback. Presence and broadcast currently operate within one Go server process. Direct local voice/video is implemented; hosted discovery and TURN/relay service are absent. Other devices must reach the local server address; media capture requires localhost or HTTPS.
- Map uploads fit the current grid. Grid dimensions can now be changed from 2–100 cells; image crop/alignment/calibration and hex grids remain future work.
- The JSON database and image data are local storage, not PostgreSQL/object storage. Save errors are currently logged rather than propagated transactionally through every API response; production persistence/error handling remains future work.
- No password reset/email verification, production auth hardening or deployment has been completed. Keep real account data private and use isolated test storage.
- Some older CSS selectors and archive files remain. Avoid cleanup that could disturb unrelated work.

## Resume checklist

1. Read this handoff and `README.md`; inspect current files and `git status` because runtime and dirty state may have changed.
2. Verify `/health` and the actual server process before testing; start the Go runtime if needed.
3. Use **Try a demo table** to reproduce the current experience, and inspect the Tools tab plus mobile inspector.
4. Preserve the accepted map-focused design, existing data and unrelated working-tree edits.
5. For further changes, test the real API and browser behavior, not only rendered markup or fixture responses. Update this document with new confirmed results when requested.

## Continuation — 12 November 2026: live multiplayer and command safeguards

The October progress report was reviewed against this newer handoff and the actual source. Continued the unfinished live-transport and command-safety items while preserving the accepted design, saved campaigns, and existing working-tree changes.

### Implementation

- Added `cmd/astralis/live.go` using `github.com/gorilla/websocket` v1.5.3, recorded in `go.mod` / `go.sum`.
- `GET /api/v1/sessions/{id}/live?after={sequence}` upgrades to a WebSocket. The first frame carries `{"token":"..."}` within five seconds; tokens never go into the URL. The default upgrader rejects other browser origins. Campaign membership and unexpired credentials gate the initial snapshot, subsequent delivery and periodic heartbeat.
- Initial `session.sync` contains events newer than the supplied sequence plus the authoritative snapshot. Subscription and replay capture share the database lock so there is no gap. Accepted commands broadcast `session.updated`; `presence` counts unique member accounts in the current session. Presence clears when the connection closes; heartbeat detects abandoned connections.
- Socket writes run outside the DB lock, with bounded per-connection queues and write deadlines. Slow readers disconnect and recover through replay. Client reconnect backs off to ten seconds. Leaving the table closes its connection.
- Guarded POST commands supply `command_id` (1–128 characters) and integer `expected_state_version`. Both are required if either is present. Legacy clients may omit both; their commands remain unguarded for compatibility.
- Accepted receipts are scoped by session, authenticated user and command ID, persisted in `command_receipts`, and fingerprint the payload. A duplicate returns the original event result and current snapshot without applying state or broadcasting again. Reusing an ID for a different payload returns 409. Membership is checked before receipt lookup.
- A new command with a stale version returns 409 `STATE_VERSION_CONFLICT` and a current snapshot. The browser refreshes that snapshot and asks the user to review/retry through its status message. It automatically retries a network failure once with the same command ID; it does not automatically resubmit a conflicting action.
- The browser displays live connection status and online account count. Incoming state waits while a form is focused or a local command is pending. It keeps five-second polling for unavailable live transport. Server snapshots remain authoritative.

### Validation

- Added `cmd/astralis/live_test.go`: persisted original dice results on duplicate retry, reused-ID rejection, invalid/stale versions, concurrent commands, membership, origin rejection, live broadcast, presence, replay, and token revocation before delivery.
- Added `scripts/check-live.cjs`: real backend/disposable DB on port 18082; separate member browser contexts; live chat/presence, focused draft preservation, accepted response intentionally lost then safely retried, reconnect without reload after server restart, mobile overflow, table-route cleanup and polling fallback.
- Existing `scripts/check-table-tools.cjs` passed against the updated backend. It still uses isolated data on port 18081.
- `scripts/check-live.cjs` passed, including the disabled-WebSocket polling fallback. `scripts/check-cinematic.cjs` also passed for desktop and 390px layouts and existing account/campaign/game flows. The live desktop screenshot is `output/screenshots/astralis-live-desktop.png`.
- `go test ./...`, `go vet ./...`, and `node scripts/check-frontend.cjs` passed after the implementation. Race detection has not been established; the earlier CGO limitation still applies.
- The updated runtime was built as `output/astralis-local.exe` and started from this project directory on port 8080. `/health` returned `astralis-go / ok`. Recheck the listener on future continuation; the root `astralis.exe` remains an older build.

### Next unfinished work

1. Per-character control for current session actors is now complete; see the next continuation section. Global character-record CRUD/assignment remains separate future work.
2. Invitation expiry, revocation, usage limits and membership role management.
3. Transactional persistence with propagated save failures, PostgreSQL migrations, and durable/object asset storage. Current JSON save errors are logged; successful disk persistence is necessary for deduplication across restarts. Receipt/event retention has no pruning policy yet.
4. Password/auth hardening (including the original Argon2id/bcrypt target), token rotation, rate limits and account recovery. Current PBKDF2 password storage remains unchanged.
5. Grid calibration, arbitrary scene dimensions, fuller character/inventory/quest workflows, Docker Compose, readiness and observability.

Live delivery currently sends full snapshots, including retained event history and inline map image data; delta delivery, bounded history and multi-instance broadcast remain scaling work. No commit, push or deployment was requested for this continuation.

## Continuation — 12 November 2026: character assignment and permissions

Added saved GM-managed ownership of the current session's party actors. Existing tokens with no assignment remain GM-controlled; the JSON database needs no destructive migration. `Actor.controller_id` and `controller_name` persist with the snapshot. The display name comes from the server's user record; the client cannot grant authority through a supplied name, user ID or role.

### Current behavior

- In **Characters**, the GM assigns a party token to a campaign member or chooses **Game master only** to clear control. The member list loads when the panel is opened; **Refresh players** reloads it after another player joins. The same controls work in the scrollable mobile inspector.
- GM ownership of the campaign grants control of every party actor and retains all existing scene tools. Players can inspect every token but can move, attack, investigate or end a turn only for their assigned party actors. Existing turn, health, range and action-budget checks still apply.
- Hostile NPCs stay GM-controlled. Assignment to an unknown/non-member account, hostile/missing token, or invalid controller field is rejected without changing state. Players cannot self-assign.
- Actor authorization runs before command receipt lookup, so a former assignee cannot bypass revocation by retrying a previously accepted command. Campaign membership remains a prerequisite. GM authority uses the persisted campaign owner ID, not a submitted user/role field.
- Fixed an actor-field mismatch: combat resolution now uses `attacker_id`, matching the authorization check even if a malicious payload also supplies a conflicting `character_id`.
- Assignment, reassignment and clearing use the standard command/version/event/persistence/live-broadcast pipeline. The client shows the controller in the action inspector and roster and disables unauthorized action controls.

### API and files

- `GET /api/v1/campaigns/{id}/members`: signed-in campaign members receive only member `id`, `display_name` and `role`. Unrelated users receive 403; emails, credentials and unrelated profiles are omitted.
- POST a normal session command with `{"payload":{"type":"token.assign","token_id":"mayank","controller_id":"member-id"},"command_id":"unique-id","expected_state_version":1}`. Use an empty controller ID to clear. The event type is `token.assigned`; normal deduplication and state-version checks apply. Use the current snapshot version, rather than the example's fixed number.
- Backend: `cmd/astralis/ownership.go`, actor fields/authorization dispatch in `main.go`, member route dispatch in `tools.go`.
- Frontend: `frontend/app.js`; the accepted visual design and existing map/scene tooling were preserved.
- Tests: `cmd/astralis/ownership_test.go`, `scripts/check-character-control.cjs`; existing frontend/browser fixtures now include the authoritative campaign owner ID.

### Confirmed validation

- `go test ./...`, `go vet ./...`, and `node scripts/check-frontend.cjs` passed.
- The real-backend character-control browser journey passed with a GM and two registered players: assignment, allowed movement/attack/interaction/end-turn, GM override, other-player restrictions, forged request rejection, transfer/clearing, restart persistence and desktop/mobile layouts.
- Existing `scripts/check-table-tools.cjs`, `scripts/check-live.cjs`, and `scripts/check-cinematic.cjs` passed after the ownership change.
- Browser tests use separate temporary databases and ports; they do not modify saved campaigns. Screenshots are `output/screenshots/astralis-character-control-desktop.png` and `astralis-character-control-mobile.png`.
- The local runtime was rebuilt as `output/astralis-current.exe` and replaced only the verified earlier Astralis development process on port 8080. The root `astralis.exe` is still an old build.

Assignments currently belong to actors within a session. They do not connect the older `/characters` records to session spawning or implement full character sheets, inventories or campaign-wide character CRUD. These remain future work. The next unfinished security item is invitation expiry, revocation and usage limits, followed by membership role management and the persistence/auth/asset work listed above. No commit, push or deployment was made.

## Continuation — 12 November 2026: reference-driven tabletop expansion

The user supplied a Roll20 editor screenshot and asked to add missing tabletop functionality. The screenshot and official toolbar/layer/measurement/token/chat documentation informed this implementation. Browser skill setup reported no browser available; discovery returned an empty list. The authenticated editor was not inspected live. Isolated installed Playwright was used for local verification, with synthetic assets/devices and temporary databases.

### Implemented surfaces

- Added the **Table** inspector tab, retaining the existing four panels. Its sections are Scenes, Drawings, Tokens, Art, Lighting, Turns, Chat, Journal, Music, Media, Macros and Settings. Added text/freehand/pin/FX and chat/settings rail controls. All displayed controls invoke implemented local or authoritative actions.
- Scenes: blank/duplicate creation, active-name editing, switching and archived deletion; maximum ten scenes. Active state is stored on the World; archived `Page` entries hold scene-specific World snapshots. Notes, handouts, macros, audio and art stay global across scene switches. Personal macros, handouts and unused library art are omitted from archived payload copies. Players receive only the active scene.
- Drawing types: line, rectangle, ellipse, cone, freehand (2–250 grid points), text and pin, with validated colors, widths and layers. Four layers: map, tokens, GM-private and foreground. Selected-layer undo and individual removal are available. Legacy line/area tools remain supported.
- Tokens: rotation, size (1–4 cells), six status markers, movement lock, initiative modifier, layer, duplication and deletion. Active tokens cannot be deleted/moved off the token layer until their turn ends. Imported portraits can be applied. Larger footprints participate in terrain/occupancy/pathfinding checks; GM/map decoration layers do not secretly block player paths. GM dragging places; an assigned player dragging submits normal movement with path/budget checks.
- Art: validated PNG/JPEG library, maximum ten assets at 2 MB decoded size each. Unused originals are GM-only; placed map/token copies are shared according to their layer/sight projection. Library deletion leaves placed copies intact.
- Initiative: server rolls 1d20 plus each living party actor's initiative modifier, stable sorting on ties, start in rolled order, and subsequent turns follow that order. Hostiles remain GM-controlled. Added movement pause/resume; locks pause movement rather than chat, dice or ending a turn. Updated the RPG glossary to match implemented initiative behavior.
- Chat: account/authorized-character identity, emotes, Enter-to-send, simple server-resolved dice formulas and GM-only dice results. Quoted names or account IDs can target whispers. Legacy `chat.send` also uses the command parser so `/w` is not accidentally broadcast as public literal text. Macros are account-private definitions; running one posts a public server roll.
- Journal: plain-text handout creation/editing/deletion with explicit GM-private/shared status. Input is rendered as escaped text.
- Music: shared HTTP(S) audio URL/title/start time and play/stop intent; explicit local audio enable, volume and join-position seeking. Sources must permit browser playback. No server-side URL fetch is performed.
- Media: authenticated WebSocket signaling plus direct WebRTC voice/video. Explicit Join triggers browser permission/capture, camera is optional, microphone can be muted, Leave/disconnect/navigation stops tracks and closes peers. Signaling is not persisted. Synthetic two-peer audio/video was used in tests. No STUN/TURN service or internet-wide connectivity guarantee is configured.
- Settings: grid width/height 2–100, scale, units and orthogonal/Euclidean/Chebyshev ruler metric. Ruler distance, labels and shared measurements use these settings; movement remains the custom orthogonal ruleset. Added zoom slider and generic unassigned-player preview, which is read-only and uses server filtering. GM preview closes live media and live updates until exited.
- FX currently provides a shared map pulse, rather than a library of programmable particle effects.

### Privacy and line of sight

- `sessionView` is the common projection for GET snapshot, session listings, event replay, live initial sync/broadcast, stale-version responses and duplicate-command responses. It omits GM-layer actors/drawings, private handouts, archived scenes and unused art originals from players. Personal macro definitions are visible only to their owner, including against GM access. Whisper/GM-roll audiences are persisted on events and enforced for every read path; the GM does not automatically receive whispers between other players.
- Optional token sight uses assigned living party actors, Euclidean radius 1–30 cells and Bresenham grid rays. Blocked cells stop rays; a facing wall cell remains visible and closed diagonal corners do not leak sight. Unassigned/dead-only players see no actors/visible cells. GM sees the entire map.
- Custom map images are pixel-masked on the server before delivery. A bounded in-process cache holds up to sixteen masked projections. Client masks also cover overlays. Default Riverside artwork is public elsewhere as sample artwork, so it is not secret content.
- Tokens outside sight, outlying drawing geometry, unseen blocked/fog cells and hidden initiative entries are removed from player snapshots. Token-specific events capture the accounts who could see them when they happened; revealing a token later does not disclose its previously hidden event history. Attacks against targets outside sight are rejected on the server.
- The original Fog tool remains explicitly a shared planning overlay. This sight model does not implement light-source simulation, doors/windows, darkvision, explored-map memory or published game-system vision rules.

### Important files and protocol

- Backend: `cmd/astralis/table.go` (tool commands, extensions and common projection), `sight.go` (rays, event audiences and custom-image masking), extended actor/drawing/event/World fields and dispatch in `main.go`, collision fixes in `tools.go`, authorization in `ownership.go`, and media signaling in `live.go`.
- Frontend: `frontend/tabletop.js` extension factory, `tabletop.css`, integration hooks in `app.js`, and script/style ordering in `index.html`. Frontend contract tests can run without the optional extension factory; browser tests load the actual module.
- New commands: `scene.create/switch/rename/delete`, `table.settings/pause`, `draw.add/delete/undo`, `token.configure/duplicate/delete`, `asset.save/delete/map/token`, `sight.settings`, `initiative.roll/start`, `handout.save/delete`, `macro.save/delete/run`, `audio.set`, `chat.post`, `measure.share`, `fx.burst`. All use the existing session command ID/version pipeline.
- Media frames: initial `session.sync` includes a per-connection ID; presence includes connection profiles/media participation. `rtc.join`, `rtc.leave`, and targeted `rtc.signal` work only between authorized, joined peers in the same session. Sender identity is supplied by the server, never trusted from the frame body. Per-frame limit is 64 KiB, queues remain bounded, and slow readers disconnect.
- Extended browser test: `scripts/check-table-workspace.cjs`, temporary data/port 18084. Existing tests keep their previous isolated ports. The rail now includes four additional drawing tools; existing tool-count checks were adjusted accordingly.

### Validation and runtime

- `go test ./...`, `go vet ./...`, `node --check frontend/tabletop.js`, and `node scripts/check-frontend.cjs` passed during implementation.
- Added `table_test.go`, `sight_test.go` and `media_test.go` for all-path privacy, whisper/macro account scoping, scene/grid persistence, initiative ordering, rejected-command immutability, larger footprints, movement locks, masked custom pixels, occlusion/hidden history, and authenticated same-session media relay.
- The complete expanded browser journey passed: scaled/shared ruler, drawing layers/text/ellipse/freehand, image import/token art, private token configuration, handouts, macros, character/emote/private chat, synthetic audio playback and two-peer video, initiative, scene switching, player preview, sight, restart persistence and desktop/390px layout.
- Existing `check-live.cjs`, `check-table-tools.cjs`, `check-character-control.cjs` and `check-cinematic.cjs` all passed against the expansion. The expanded journey also verified that an assigned player's drag spends the server movement allowance and that the player's browser displays the sight mask and omits unseen hostile tokens.
- Screenshots: `output/screenshots/astralis-table-workspace-desktop.png` and `astralis-table-workspace-mobile.png`. A real microphone/camera and internet-wide peer connectivity have not been verified.
- Player-side sight screenshot: `output/screenshots/astralis-table-workspace-player.png`. `/health` returned `astralis-go / ok` for the updated local runtime. Leaving a table automatically clears generic player-preview mode.
- The tested local runtime is `output/astralis-expanded.exe`, started from the project root on port 8080 after verifying/replacing only the previous Astralis process. Saved `data/astralis.json` was not used by disposable tests. No commit, push or deployment was made.

### Remaining reference gaps

This implements the core map/editor toolkit, not full Roll20 feature parity. Remaining larger work includes full character sheets/inventory/spells and campaign-wide character records; custom rule/compendium import and published modules; cards/rollable tables and advanced dice expressions; hex grids and image crop/calibration; colored lights, doors/windows, darkvision and explored-map memory; programmable FX; hosted lobby and robust STUN/TURN calling. Existing infrastructure work (invitation expiry/revocation/limits, membership management, transactional save errors, PostgreSQL/object storage, auth hardening, observability) remains open. Full snapshots, inline images and JSON persistence are still prototype scaling limits.

## Continuation — 14 November 2026: character, environment and GM toolkit

The user requested important features from comparable VTTs and explicitly asked to implement all three proposed areas one after another: sheets/inventory/spells; lighting/doors/map setup; journals/quests/random tables. Reviewed official Foundry actor/journal/rollable-table documentation, Quest Portal sheets/features, Owlbear fog documentation, Roll20 lighting/doors documentation and Alchemy GM orientation. Source links are in the README. Preserved the accepted map-focused design and unrelated working-tree changes.

### Implemented

- **Sheets** in Table: GM creates/deletes system-neutral sheets and assigns a campaign member. GM/assignee can edit; private sheets are visible to GM/assignee and shared sheets to other members. Includes background, attributes, inventory/spell/resource notes and structured tracked entries. Inventory/spell/resource entries have quantity/capacity, preparation/equipment state, editing/removal and server-authoritative consumption. Unprepared spells and depleted entries reject use. Note edits preserve structured entries. Sheet dice rolls are explicitly public. Values do not automatically change map token combat stats or resolve spell effects.
- **Quests**: GM creation/editing/deletion, title/description/objectives, planned/active/completed/failed status and private/shared visibility. Objectives are plain text using `[ ]`/`[x]`; players read shared quests, while the GM updates progress. Existing private/shared journal handouts remain available.
- **Tables**: GM creates/edits/deletes weighted outcome lists (up to 100 entries, weight 1–1000), draws with/without replacement, resets depleted tables, and chooses private/shared definitions and outcomes. Server resolves draws. Shared results appear in the panel/history; private definitions/results are filtered from player reads.
- **Environment**: cell doors/windows with open/closed, locked and secret flags. Closed doors block paths and sight; windows allow sight while blocking movement. The GM manages them. Players can toggle a visible unlocked door when an assigned living character is adjacent. Occupied doorway closures are rejected. Secret doors are absent from player projections.
- **Lighting**: optional require-light visibility (enable token sight first), wall-aware lights with radius 1–30 and color markers, and token darkvision 0–30. Light visibility is binary; colors identify sources rather than blending physically colored light. Opaque walls/closed doors stop light and darkvision. Windows transmit sight/light. The token's cell remains visible in darkness.
- **Map setup**: server-side percentage crop and 0/90/180/270-degree rotation of uploaded custom images. Replaces the current image with a PNG limited to 2048 pixels per side and 5 MB. Preserve the original file to undo. Existing grid controls set dimensions/scale; shrinking rejects outlying doors/lights. Artwork fitting remains to the scene grid; this is not interactive image-to-grid calibration or hex support.
- **Invitations** in Tools: expiry hours and maximum joins (1–720), owner-only list/revoke API and UI. New defaults: 168 hours and ten joins. Existing string-token database records remain compatible; old codes without rules retain their previous unlimited/no-expiry behavior until revoked. Existing members do not consume uses or have roles overwritten. Expiry, revocation and exhaustion prevent redemption. Revoke prevents new joins; it does not remove current members. Usage checks and increments run under the DB mutex.

### State, authorization and protocol

- `cmd/astralis/adventure.go`: sheet/entry, quest and roll-table records/commands/projection. Embedded `AdventureFeatures` is global to a saved session and preserved across scene switches. Archived scene copies omit these global records. These are not yet campaign-wide `/characters` records or token-spawning templates.
- `cmd/astralis/environment.go`: per-scene doors/lights/darkness, map raster cropping, common movement/sight helpers and filtered environment projection. `Actor.darkvision` persists with the token. Blank new scenes start without environment settings; duplicate scenes copy them.
- `cmd/astralis/invitations.go`: persisted `invitation_rules` metadata alongside the compatible `invites` map. No destructive data migration. Owner-only `GET/POST /api/v1/campaigns/{id}/invitations`, `DELETE /api/v1/campaigns/{id}/invitations/{token}`, existing POST redemption path. Creation accepts numeric `expires_in_hours` and `max_uses` and returns token, expiry, use count and limits.
- New commands: `sheet.save/delete/roll`, `sheet.entry.save/use/delete`, `quest.save/delete`, `rolltable.save/delete/draw/reset`, `door.save/delete/toggle`, `light.save/delete`, `lighting.settings`, `token.vision`, `map.calibrate`. Existing command IDs, expected versions, persistence and live snapshots apply. Assignment authorization precedes receipt lookup for sheet commands.
- New records use the common `sessionView` privacy path, covering snapshot/list/replay/live/conflict/duplicate responses. Content is kept out of public configuration event payloads. Commands copy state before validation, preserving rejected-command immutability.
- Frontend integration remains in `frontend/tabletop.js` and `app.js`; no framework or new dependency. Section controls reuse existing form styling and mobile scrolling. Updated outdated landing FAQ copy about sheets, WebSockets and calling.

### Confirmed validation and runtime

- Passed `go test ./...`, `go vet ./...`, `node --check frontend/app.js`, `node --check frontend/tabletop.js`, and `node scripts/check-frontend.cjs`.
- Added `adventure_test.go` and `environment_test.go`: invitation expiry/revocation/limits/persistence/owner-role preservation/legacy codes; sheet ownership/private data; tracked entry preparation/depletion; note-edit retention; quest/table privacy; scene/restart persistence; closed/open/window/secret/locked doors; darkness, illumination and darkvision; grid shrink rejection; and crop pixel correctness/permissions/rejected-state immutability.
- Added `scripts/check-adventure.cjs` on port **18085**, temporary DB and isolated browsers: GM sheet creation/assignment, player spell consumption and resource edit, shared quests, weighted no-replacement draw, door/light/darkvision/crop forms, invitation revocation, restart persistence and desktop/390px layouts. Passed after final overlay fix; asserts that the doorway marker actually renders.
- Existing `check-table-workspace.cjs`, `check-table-tools.cjs`, `check-character-control.cjs`, `check-live.cjs` and `check-cinematic.cjs` all passed during this continuation. Existing expanded journey additionally covers sight, custom-map masking, private content and synthetic voice/video. No real media hardware or internet relay verification was added.
- Screenshots `output/screenshots/astralis-adventure-desktop.png` and `astralis-adventure-mobile.png` were inspected. Fixed an overlay cleanup ordering issue found during review; reran the new journey to regenerate screenshots and verify the marker.
- Browser skill connection setup reported `No browser is available`, and discovery returned `[]`; installed isolated Playwright provided verification. No personal browser profile was accessed.
- Built and started `output/astralis-current.exe` from the project root on port **8080**, with stdout/stderr logs under `output/`. No prior 8080 listener was present. `/health` returned `astralis-go / ok`. Root `astralis.exe` and `output/astralis-expanded.exe` are older builds.
- Tests used disposable storage; saved `data/astralis.json` was not used for test accounts. No commit, push or deployment was made.

### Remaining scope

The three requested workflow areas now have working baseline implementations, with the boundaries described above. Further depth includes campaign-wide sheet/token linkage, system templates and automatic rule effects; licensed/published libraries and custom compendium import; proper interactive grid alignment and hex grids; explored-map memory and blended colored lighting; cards and advanced dice; journal file/folder management; hosted discovery and TURN service. Membership removal/roles, transactional save errors, PostgreSQL/object storage, stronger account lifecycle/rate limits and production observability remain open. JSON persistence, inline image payloads and full retained snapshots still limit scale. Do not claim complete feature parity with commercial VTTs.

## Continuation — 14 November 2026: tabletop interaction polish

The user supplied a screenshot of the saved Riverside demo with drawings and asked to smooth the workspace using two Roll20 video references:

- <https://www.youtube.com/watch?v=l0dBSXUdMRA> — public YouTube metadata identifies Roll20's **Learn How to GM on Roll20 in JUST 10 MINUTES**.
- <https://www.youtube.com/watch?v=hFOxs45TRgc> — referenced as **How to use the Turn Tracker** by the Roll20 wiki.

Direct YouTube fetching failed, and Browser setup/discovery again reported no available browser (`[]`). Video playback/frames/transcripts were not inspected; do not claim the videos were watched. Public video metadata, the supplied screenshot and Roll20's official [Turn Tracker](https://help.roll20.net/hc/en-us/articles/360039178634-Turn-Tracker) and [toolbar documentation](https://help.roll20.net/hc/en-us/sections/360008414373-Toolbar) informed the implementation. Isolated installed Playwright verified Astralis.

### Changes

- Added `frontend/canvas.js` and `canvas.css`, loaded before the existing workbench. Kept all existing tools and API actions. Compact rail/header/hotbar/inspector controls, lower-opacity grid and legacy annotation styling, horizontally scrollable section navigation, and a hide/show inspector control give the map more room. Sidebar scroll survives ordinary rerenders.
- Added an on-map drawing flyout with freehand, line, rectangle, ellipse, cone, text and pin choices; quick stroke color/width/layer controls and selected-layer undo. Existing sidebar and legacy two-click workflows remain available.
- Freehand strokes now retain sub-cell coordinates and render as quadratic curves with round joins/caps. Existing cell-based freehand strokes also render as smoothed curves without altering saved data. Old line/area overlays use lighter outlines/fills.
- New freehand/shape gestures show a local preview while dragging and commit once on release. Escape/pointer cancellation discards unfinished work. Legacy click-click measurements and annotations still work.
- Token dragging previews the destination locally. The pending-command render retains that destination while awaiting authoritative acceptance; rejected moves restore the server position. Movement/placement permissions and budgets still come from the server. Other accepted token movements animate over 180 ms when the view size is stable; reduced-motion users receive immediate positions.
- Space-drag and middle-button drag temporarily pan without changing the selected tool. Pan mode also preserves a simple token click for selection/targeting; this was caught by the combat regression and fixed. Ctrl/Command + wheel zooms around the pointer, while ordinary scrolling still explores the map. Existing zoom buttons/fit remain available.
- Live snapshots defer rendering while a map gesture is active, as they already do for focused form edits, so remote updates do not replace a drag mid-gesture.
- Added an accessible floating **Turns** panel: party initiative, current turn/round, score editing, earlier/later controls, roll, sort, start and next turn. GM show/hide is persisted/shared; the player receives only its existing authorized token projection. Reordering changes initiative order without resetting the current actor's movement/attack allowances. Existing custom-party rules remain in force: this is not full Roll20 turn-tracker parity, hostile initiative/custom rounds/cards are not added.

### Protocol and files

- `Drawing.path` is an optional array of finite floating-point `{x,y}` positions for freehand, validated inside the current map and limited to 2–1000 points. Existing integer endpoints/`points` remain present for compatibility. Projection checks smooth-path cells as well; grid shrink validation includes them. `cmd/astralis/tools.go`, `table.go`, and `frontend/tabletop.js` contain these changes.
- New `initiative.manage` command in `cmd/astralis/initiative.go`: GM-only actions `visibility` (`visible` boolean), `score` (`token_id`, score -100..100), `up`, `down`, `remove`, `clear`, and `sort`. Only living eligible party tokens can be added. Uses the normal command ID/version pipeline. `tracker_visible` belongs to the scene's `TableFeatures`.
- `frontend/app.js` supplies pointer-anchored zoom, a before-render workbench hook, and the map-gesture deferral flag; existing routes and persisted campaign data are unchanged.

### Confirmed validation

- `go test ./...`, `go vet ./...`, JS syntax checks and `node scripts/check-frontend.cjs` passed.
- Added `cmd/astralis/canvas_test.go` for float-path persistence, invalid-path rejection, private drawing projection, tracker permissions/visibility/order, and preservation of active action budgets.
- Added `scripts/check-canvas.cjs` using port **18086**, a disposable database and isolated browsers. Passed: live smooth/shape/token previews, precision and persistence, single drawing commits, score editing, pointer-anchored zoom, Space pan, sidebar collapse, animated incoming token placement with motion enabled, desktop/390px layouts, and restart persistence.
- Existing `check-table-workspace.cjs`, `check-table-tools.cjs`, `check-adventure.cjs`, `check-live.cjs`, `check-character-control.cjs`, and `check-cinematic.cjs` passed during this work. The cinematic journey passed after restoring click-selection in Pan mode. No test accounts were added to `data/astralis.json`.
- Screenshots `output/screenshots/astralis-canvas-desktop.png` and `astralis-canvas-mobile.png` were visually reviewed; corrected the mobile banner layout and ensured screenshots await tracker close completion.
- Built **`output/astralis-polished.exe`**. Verified the existing listener belonged to project `output/astralis-current.exe` (PID 14592) before replacing only that process. The current development server is `output/astralis-polished.exe` (PID 24988 at verification), started hidden from the project root on **8080**. `/health` returned `astralis-go / ok`; logs are `output/astralis-polished.stdout.log` and `.stderr.log`. Recheck PIDs/listeners in future sessions. Older executables remain on disk.
- Reload the browser to load the new scripts/styles. No commit, push or deployment was made.

Remaining limits from the preceding continuation still apply. Pinch-to-zoom, editable drawing selection/resizing, touch/stylus-specific pressure handling, custom/NPC turn entries, and full video-reference fidelity were not implemented or verified in this polish pass.

## Continuation — 14 November 2026: contextual editor workflow

The user asked to make Astralis better like Roll20 after the interaction-polish pass. Consulted official Roll20 toolbar, layer and token documentation for grouped tools, contextual token properties and map-layer workflow:

- <https://help.roll20.net/hc/en-us/articles/360039674753-Toolbar-Overview>
- <https://help.roll20.net/hc/en-us/articles/360039675053-Layers>
- <https://help.roll20.net/hc/en-us/articles/360039674573-Token-Features>

### Implemented

- Added `frontend/editor.js` / `editor.css`, integrated through the existing canvas factory. The main rail keeps Select, Pan, Draw, Ruler and Ping; adds quick dice, Turns, Scenes, Layers and More; and retains Grid/Fit/Sync/Guide. Secondary legacy tools move into a More flyout, preserving their original controls and all map modes. The rail can still scroll on short screens. Active drawing/secondary-mode indications stay visible on the grouped controls.
- Primary Draw opens the existing shape flyout. Selecting a shape programmatically does not reopen it; this interaction was caught by the canvas journey and corrected.
- Right-click or double-click a token to open an anchored actions panel. The GM can set HP or apply signed adjustments (for example `-5`), rotate by 45 degrees, lock movement, toggle six existing status markers, add an eligible party token to initiative, and change its layer. The panel also opens the normal inspector or ends an authorized active turn. HP bounds, token occupancy/layer restrictions, ownership and combat rules are still server-validated.
- Player token panels are read-only except for previously authorized end-turn actions. They do not include GM HP, marker, rotation, lock or layer controls. Hidden tokens still do not enter player snapshots.
- Selected/targeted/hovered tokens show small HP bars. Existing markers and artwork remain supported.
- Quick dice tray provides d4/d6/d8/d10/d12/d20/d100 with count/bonus, latest result and public server-resolved history. Count and bonus are local form preferences; rolls persist through the existing API.
- GM drawing-layer flyout changes the layer for new annotations. It is explicitly a drawing layer picker, not a claim that all token selection is filtered by an active editor layer. Token layers change through their contextual menu.
- GM scene tray displays the current scene, archived-scene thumbnails, the live-scene indicator and one-click switching. Prepare/New opens the existing Scenes panel. Switching remains an authoritative command that moves the whole table.
- The Table inspector includes a labeled section selector for direct access to all panels, alongside the existing scrollable section buttons.
- Turn tracker entries can now be dragged before another entry. Keyboard earlier/later buttons remain. Added GM-only `initiative.manage` action `before` with `token_id` and `before_id`; validates both entries and preserves the active actor's movement/attack budgets. This does not add hostile/custom turn entries or alter the party ruleset.

### Validation

- Passed `go test ./...`, `go vet ./...`, JS syntax checks and `node scripts/check-frontend.cjs`.
- Extended `canvas_test.go` to cover drag-order authorization, invalid targets, ordering and action-budget preservation.
- Added `scripts/check-editor.cjs`, port **18087**, disposable DB and separate GM/player browsers. Passed grouped-tool access, quick dice count/bonus, drawing-layer selection, contextual HP adjustments/rotation/markers/locking, double-click, player restrictions, scene switching, drag turn order, restart persistence and desktop/390px layout.
- Existing `check-canvas.cjs`, `check-table-tools.cjs`, `check-table-workspace.cjs`, `check-adventure.cjs`, `check-live.cjs`, `check-character-control.cjs` and `check-cinematic.cjs` passed. The legacy-tools journey now opens More before selecting its secondary line/area/block/fog tools, matching the new UI.
- Visually reviewed `output/screenshots/astralis-editor-desktop.png` and `astralis-editor-mobile.png`. Existing data was preserved; test accounts and interactions used only disposable storage.
- Browser runtime still reported no available browser; isolated installed Playwright supplied browser verification. No personal profile was accessed.
- Current runtime is `output/astralis-editor.exe` on port **8080**, started from the project root after verifying the prior listener was `output/astralis-polished.exe`. Logs are `output/astralis-editor.stdout.log` / `.stderr.log`. `/health` returned `astralis-go / ok`. Reload the browser for the new assets. No commit, push or deployment was made.

Remaining infrastructure, game-system, advanced lighting and editing limits from preceding sections still apply. The new scene tray and token menu are Astralis workflows inspired by the official references; no full Roll20 parity is claimed.

## Continuation — 15 November 2026: game profiles, rules references, ruler and FX

The user supplied four Roll20 screenshots (dynamic lighting, effects, dice and measurement) and the editor URL, and requested missing tools, D&D/Warhammer/other tabletop rules and a game selector in the header. Read the existing handoff against source before extending the Go application. Preserved the existing uncommitted work and saved campaigns.

### Implemented behavior

- **Game to play** in the tabletop header: GM-only persisted selection, live-shared with players, for Astralis Basic, D&D 5e 2014, D&D 5e 2024, Warhammer 40,000, Pathfinder 2e, Fate Core and Generic percentile. The profile is global to the session across scene switches. Old saves with no `game_system` use Astralis Basic. Choosing a game preserves tokens, combat values, encounter state and scene scale. **Apply scale** is a separate explicit action.
- **Rules** opens original quick references, a searchable book directory, tailored checks and GM-created manual sheet templates. Templates fill the existing sheet's `stats`/`details` fields and are private by default; they do not link sheet values to token stats. Check form preferences survive rerenders separately per profile.
- Server-resolved helpers: D&D d20 difficulty comparison with normal/advantage/disadvantage; Pathfinder four degrees with natural-face adjustment; Warhammer up to 100 d6 with a chosen 2–6 threshold and success count; Fate 4dF plus modifier; generic percentile roll-under. D&D helper is an ordinary ability-check comparison, not automatic attacks/saves/death saves. Warhammer rerolls/modifiers/wound/save chains/damage allocation and all other system exceptions remain manual. System results are stored as `dice.roll` with system/outcome metadata.
- Existing quick dice tray now exposes All players / You and GM / Only you. Self-only results are omitted from other accounts, including the GM. System checks support the same audiences. Existing public `dice.roll` clients retain their original behavior.
- **Ruler** flyout with line, square, circle, cone and one-cell-wide beam; snap-to-center/free coordinates; drag and click-click; share-on-release, explicit sharing and local clear. SVG previews show the scene's authoritative metric/scale. Circle means radius, cone means depth; square uses opposite corners and the label reports distance between endpoints. Geometry can extend beyond the map; endpoints must remain inside. No movement or combat effects. Existing side-panel sharing continues to work for line measurements.
- **FX** primary rail shortcut and flyout: burst, aimed breath, aimed beam and nova, eleven element colors, transient CSS particles and reduced-motion rendering. Click for point effects or drag to aim. Existing `fx.burst` clients default to magic burst. FX events retain their parameters; the renderer only plays recent events, so old history does not replay on reload. No damage automation or custom particle scripting.
- **Light** shortcut opens a combined token-sight and require-illumination flyout, with access to the existing environment editor for doors, source lights and token darkvision. Reuses the already implemented server visibility/masked-map behavior. Binary illumination, colored source markers and square-cell occluders remain the model; no blended colored lighting, arbitrary wall segments, or explored-map memory was added.

### Books, sources and licensing

- Bundled the unchanged publisher PDFs for **SRD 5.1** (403 pages) and **SRD 5.2.1** (364 pages) under `frontend/rules/`, retaining their legal pages. `ATTRIBUTION.txt` contains both required attribution statements and the CC BY 4.0 license link. `manifest.json` records publisher URLs, byte counts and SHA-256 checksums. `scripts/fetch-open-rules.cjs` verifies existing files/writes the manifest; `--refresh` downloads again. Node's initial direct fetch timed out; Windows `Invoke-WebRequest` successfully obtained the files. The included PDF link was verified over HTTP in the browser journey.
- Other books use external publisher/publisher-endorsed links; commercial books and Warhammer codexes were not copied. Warhammer profile is edition-neutral and points to the publisher's downloads page to avoid pinning an unverified edition. A comprehensive licensed game catalog, book ingestion/search within PDFs, full compendium import and all-system automation remain future work.
- Official references consulted: <https://www.dndbeyond.com/srd>, <https://www.warhammer-community.com/en-gb/downloads/warhammer-40000/>, <https://paizo.com/pathfinder/getstarted>, <https://fate-srd.com/official-licensing-fate>, and Roll20's FX API documentation <https://help.roll20.net/hc/en-us/articles/360037772833-Mod-Scripts-Function-Documentation>.
- The Roll20 editor URL was inaccessible through web retrieval; its signed-in game was not inspected. The Browser skill runtime reported **No browser is available**, and recovery discovery returned `[]`. Existing isolated Playwright supplied local UI verification; no personal browser profile was accessed.

### Protocol and privacy

- `system.select` (`system`), `system.check` (`bonus`, `target`, optional `count`/`mode`, `visibility`), and `dice.quick` (`count`, `sides`, `bonus`, `visibility`) run through the existing authoritative command/version/deduplication/event/persistence/live pipeline. Selection is GM-only; checks and dice are campaign-member tools. Invalid input rejects without changing state.
- Extended `fx.burst` carries `effect`, `element`, optional `to_x/to_y`; extended `measure.share` carries optional `shape` and finite fractional endpoints. Default old commands remain supported. Shared events contain server-computed measurement distances and scene IDs. The renderer filters these effects to their originating scene and clears local ruler geometry on scene changes. The new browser journey verifies scene isolation.
- To avoid exposing hidden spatial information, when token sight is active GM FX and GM measurements are GM-only. A player's measurement requires visible endpoints and is delivered only to that player and the GM. The event audience is fixed at creation and enforced by the existing common session projection on all read transports.
- `GameSystem` is in `TableFeatures`; scene switching restores it from the global live state. UI integration is through `AstralisSystems` within the existing canvas factory; script/style entries are in `index.html`. Legacy FX/ruler renderers remain as fallbacks when the new module is absent. The new module prevents duplicate line labels/pulses.

### Validation

- `systems_test.go`: profile owner authorization, invalid-command immutability, preserved token/scale state, scene and disk persistence; private dice projection; D&D advantage, D6 pools, Fate and percentile bounds; Pathfinder boundary/natural-face cases; free ruler endpoints/metric computation; FX validation, authority and sight privacy.
- `scripts/check-systems.cjs`, port **18088**, disposable DB and separate GM/player browsers: saved/live header selection, D&D advantage, Warhammer pools, Fate, sheet template population, book search and included PDF serving, private dice, ruler shape/free coordinates/sharing/clear, shared aimed FX, sight/darkness controls, restart persistence and 1600px/390px screenshots. Screenshots are `output/screenshots/astralis-systems-desktop.png` and `astralis-systems-mobile.png`.
- Fixed a form `name` omission discovered by the new journey. Screenshot review found and fixed mobile header/panel layout, duplicate legacy ruler labels/pulses and a stale shared ruler after local clear. The existing tools regression caught click suppression after rapidly switching from measuring to drawing; suppression now applies only to the gesture's own release click. Existing tools/workspace journeys now close the new ruler options panel before clicking cells behind it.
- Existing editor, canvas, table-tools, table-workspace, adventure, live, character-control and cinematic journeys passed during this continuation. Go tests, Go vet, JS syntax and frontend contracts passed. All browser accounts and test interactions used temporary storage, not saved `data/astralis.json`.

Remaining scope includes full published-system combat/movement, campaign-wide sheet/token linkage, licensed commercial compendiums, deeper Warhammer army/phase/mission support, advanced lighting/FX, and infrastructure limits recorded above. No full Roll20 parity, universal rulebook catalog or production readiness is claimed. No commit, push or deployment was requested.

### Current local runtime

Built and started `output/astralis-systems.exe` hidden from this project root on **8080** (PID **28908** at verification). The previous listener was a Go-run Astralis executable; verified its `astralis-go / ok` health response and exact serving of this project's current `systems.js` before stopping only that process. The new process explicitly uses `data/astralis.json`; its file checksum was unchanged across the refresh. New `/health` returned `astralis-go / ok`, and `/rules/dnd-srd-5.2.1.pdf` returned HTTP 200. Logs are `output/astralis-systems.stdout.log` / `.stderr.log`. Earlier executables remain on disk. Recheck listener/PIDs on continuation, and reload the browser for current scripts/styles. No commit, push or deployment was made.

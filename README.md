# Astralis Game Table

Deployment destination and current prerequisites: [ASTRALIS_DEPLOYMENT.md](ASTRALIS_DEPLOYMENT.md). The requested Vercel account is `irnahuras-projects`; publishing the complete app requires access to that account and durable hosting for the Go backend.

## Help and tool guides — 25 November 2026

Open `/#/help` for **32 searchable guides** covering campaigns and every current tabletop tool, with player/GM filters and shareable guide URLs (`/#/help/sheets`, for example). Learn, the campaign sidebar, the footer and the existing getting-started/glossary pages link to the handbook.

**Build an adventure** at `/#/help/adventure` connects eleven preparation lessons, local course progress, current-table content counts and seven original worksheets. Copy/download resources or prepare a private Journal draft and save it after review. Existing drafts are preserved. Table → Sheets now provides **Map tokens**: link an existing token, unlink it, or place a new token from a sheet. Open a visible linked sheet from the token menu. New party tokens inherit the sheet controller once; existing stats/control remain independent and sheet reassignment does not reassign placed tokens. See the [Roll20 workflow audit](ASTRALIS_ROLL20_ADVENTURE_AUDIT_2026-11-25.md) for verified coverage and gaps. `node scripts/check-adventure-course.cjs` verifies the course and linked-token flow on disposable port 18090.

Inside a table, choose **Help & tools** or press **?**. **Ctrl/Command + K** opens the tool finder. Guides can open authorized tools directly, including hidden secondary tools and collapsed/mobile workspace panels. **How to use this** explains the current inspector section. A six-step table tour introduces the workspace without issuing game commands. GM/player preparation checklists are manually checked and saved locally per account and campaign. Help stays open while the table receives updates; Escape closes it and focus returns to the table.

The FAQ and glossary now describe the implemented sheets, lighting, journals and game profiles accurately. Existing gameplay and permissions are preserved. `frontend/learning.js` is the shared content/UI source; `learning.css` contains scoped styles. Run `node scripts/check-learning.cjs` after building `output/astralis-test.exe`, with the existing Playwright setup. The journey uses disposable storage on port **18089** and verifies guides, launch targets, role restrictions, checklist persistence, keyboard/focus behavior, desktop/390px layouts and unchanged encounter state.

See [the 25 November checkpoint](ASTRALIS_PROGRESS_2026-11-25.md) for scope, reference-access limitations and validation.

For complete continuation context, implementation details, validation results and known limitations, read [ASTRALIS_HANDOFF.md](ASTRALIS_HANDOFF.md).

Latest checkpoint: [ASTRALIS_PROGRESS_2026-11-25.md](ASTRALIS_PROGRESS_2026-11-25.md). The [15 November inventory](ASTRALIS_PROGRESS_2026-11-15.md) retains the consolidated game features and deeper implementation backlog.

## Game profiles and reference tools — 15 November 2026

The tabletop header now has **Game to play** and **Rules**. The GM selects a saved, live-shared profile: Astralis Basic, D&D 5e (2014 or 2024), Warhammer 40,000, Pathfinder 2e, Fate Core or a generic percentile game. Selection preserves the encounter and existing map scale. Rules provides tailored server-resolved checks, manual sheet templates, optional scene-scale presets and a searchable rulebook directory. Profiles do not replace the basic encounter combat/movement engine.

- D&D: normal/advantage/disadvantage ability checks with modifier and difficulty. Pathfinder: four degrees of success including natural 1/20 degree adjustment. Warhammer: 1–100 d6 and a 2+–6+ success threshold. Fate: 4dF plus modifier. Percentile: 1d100 at or below a target. Game-specific exceptions and subsequent effects remain manual.
- The existing dice tray now offers **All players**, **You and GM**, and **Only you**. Private rolls are filtered through the existing event projection, including against the GM for another account's self-only roll.
- **Ruler** opens line, square, circle, cone and beam options, center snapping/free coordinates, sharing on release, explicit sharing and clearing. Drag or click two points. Circle labels report radius; cone labels report depth. Measurements use the scene metric and scale; they do not spend movement. Beam width is one cell; templates may extend beyond the map.
- **FX** opens burst, aimed breath, aimed beam and nova in eleven element colors. Effects are temporary visual particles, synchronized through saved events, with reduced-motion support; they do not inflict damage. With token sight enabled, GM FX stays GM-only and a player's shared measurement is limited to that player and the GM, with visible endpoints required.
- **Light** opens the existing server-enforced token sight/darkness settings and links to lights, doors and darkvision. Light colors remain source markers; illumination is binary and explored-map memory is not implemented.

Unchanged D&D [SRD 5.1](frontend/rules/dnd-srd-5.1.pdf) and [SRD 5.2.1](frontend/rules/dnd-srd-5.2.1.pdf) are bundled under CC BY 4.0, with [attribution](frontend/rules/ATTRIBUTION.txt) and [download sources/checksums](frontend/rules/manifest.json). Other rulebooks link to publisher or publisher-endorsed resources, including Warhammer downloads, Archives of Nethys and Fate SRD. Commercial rulebooks, codexes and complete paid compendiums are not bundled. Warhammer is an edition-neutral pool helper; check the edition/revision of documents on the publisher's downloads page.

New backend: `cmd/astralis/systems.go` and `systems_test.go`. New UI: `frontend/systems.js` / `systems.css`. `node scripts/check-systems.cjs` runs the new real-backend GM/player journey on port 18088 with disposable data. The downloader `node scripts/fetch-open-rules.cjs` verifies existing PDFs and writes the manifest; `--refresh` downloads again and requires network access. Browser tests accept the existing `PLAYWRIGHT_MODULE` override.

The main runtime is now Go. Python remains optional for future LLM connectivity and is not required to run the game.

## Run locally

```powershell
go run ./cmd/astralis
```

Open <http://localhost:8080/>. Health is available at `/health`.

The Go service serves the functional frontend in `frontend/` and the latest v2 design archive under `/design/`. It persists local state in `data/astralis.json` and exposes registration/sign-in, campaigns, invitations, characters, maps, sessions, snapshots, event replay, and authoritative commands for movement, interaction, combat, and turns.

The authoritative path is:

```text
frontend → Go HTTP API → server-side game rules → persisted snapshot + event
```

The server validates player movement, dice, damage, permissions and turns. Game masters can explicitly place tokens and edit HP through separately authorized tools.

The redesigned frontend includes a public landing page, account registration and sign in, campaign creation, saved session resume, and a tactical table with server positions, blocked cells, actor selection, movement, attacks, and turns. The illustrated village on the public page is labeled as a preview. Authentication lasts seven days and survives server restarts. The campaign library discovers saved sessions from the server after sign-in.

Validation commands:

```powershell
go test ./...
go vet ./...
node scripts/check-frontend.cjs
```

The frontend contract checks use API response fixtures. The Go journey test uses temporary storage and does not modify saved campaigns.

The tabletop includes a character inspector, party and NPC roster, active turn display, blocked movement controls, a combat target panel, snapshot refresh, and an encounter log containing server resolved attack rolls, damage, movement, and discoveries. The RPG guide is available at `/#/guide`.

The public page leads with an interactive tabletop preview rather than a portrait or WebGL backdrop. The campaign workspace includes search, name/date sorting, saved-session status, and campaign cards backed by the API. The game workspace includes a map tool rail, drag panning, zoom, fit to view, grid controls, an active-turn bar, and separate Actions, Characters, and Log panels. On mobile, the map remains scrollable and the inspector opens as a dismissible panel. Map view state survives character selection and game actions.

The landing page also includes overlapping tabletop and character previews, accessible feature tabs for Maps/Characters/Combat, details of the included encounter, and learning resources. The Learn menu opens the RPG glossary, a seven-step first-session guide at `/#/getting-started`, or a three-step interactive preview walkthrough. The walkthrough uses the actual public preview controls; it does not create an account or alter a campaign. Character stats and starter-encounter counts describe the included encounter rather than published game-system support.

Reference patterns were reviewed from public sites and official documentation: [Roll20](https://roll20.net/) for a product-led hero and campaign organization, [Alchemy](https://help.alchemyrpg.com/en/articles/9714185-gamemaster-orientation) for scene and party panels, [Foundry](https://foundryvtt.com/article/tutorial/) for the map tool rail and tabbed sidebar, and [Owlbear Rodeo](https://docs.owlbear.rodeo/docs/scenes/) for map-focused navigation and scene controls. Quest Portal's public page was also reviewed. Captures and source links are in `output/references/`; Owlbear's browser capture encountered a security check, so its official documentation supplied the reference instead. Astralis keeps its own assets, copy, and game API; no competitor feature claims, customer counts, or licensed game support were added.

`node scripts/check-cinematic.cjs` runs isolated Playwright UI tests against the local server and writes desktop and 390px screenshots to `output/screenshots/`. It verifies campaign filtering/sorting, sidebar tabs, map pan/zoom/fit, preview controls, account and campaign flows, and encounter actions. It requires Playwright; `PLAYWRIGHT_MODULE` can point to an existing Playwright package. API responses in this browser test are fixtures, so it does not change saved campaign data. The separate Go tests verify the real API using temporary storage.

## Try the tabletop tools

Open <http://localhost:8080/#/login> and choose **Try a demo table**. No email or password is needed. Each entry creates a separate campaign with a GM account and an active encounter. Reloading the page or restarting the server preserves this sandbox while its browser sign-in remains valid. Signing out and choosing demo again creates a new sandbox.

The map toolbar provides Select, Pan, Ruler, Line, Area, Block, Fog, Place, Ping, Grid, Fit and Sync. The ruler uses orthogonal grid-space distance. Line and Area use two cell clicks. Block toggles traversable terrain; movement checks routes, occupied cells and the remaining turn allowance. Fog is a shared planning overlay, not hidden player information or dynamic lighting. Place lets the GM move the token selected in Tools without spending its turn.

The **Tools** tab contains:

- Server-resolved dice rolls with a result card and saved event log.
- Riverside/blank background switching and PNG/JPEG map upload (5 MB, 20 megapixel maximum), fitted to the 20 × 14 grid.
- Persistent line/area annotations with undo, terrain edits and fog planning marks.
- Token creation, name/HP editing, and placement for party members and hostile NPCs.
- Shared session notes, chat, and campaign invitation codes.

Players register separately and redeem an invitation in their campaign library. They can resume the shared table, chat, roll dice and ping. GM editing tools are restricted in both the UI and API. WebSockets deliver live session updates and member presence, with automatic reconnect and sequence replay. Five-second polling remains a fallback. Active form drafts are preserved. Local voice/video is now available through direct peer connections. A hosted lobby and published game-system compendiums remain unimplemented.

Passwords use salted PBKDF2-SHA256; credentials and seven-day tokens persist in the local database. Existing accounts created by the earlier backend may need to be recreated if their password was never persisted. Session reads and commands now check campaign membership, and map/session creation checks ownership. Turn order cycles living party tokens, resets movement and attacks, and advances the round on wrap; hostile NPCs remain GM controlled.

To validate the new features against a real backend with disposable data:

```powershell
go test ./...
go vet ./...
go build -o output/astralis-test.exe ./cmd/astralis
node scripts/check-table-tools.cjs
```

The browser journey launches its own server on port 18081, creates two isolated browser contexts, tests the tools and invitation/player flow, restarts that server to verify persistence, and removes its temporary database. It does not touch `data/astralis.json`. Set `PLAYWRIGHT_MODULE` to an installed Playwright package if needed. `ASTRALIS_DATA_FILE` overrides the normal database location for testing.

## Live multiplayer continuation — 12 November 2026

The browser sends a unique `command_id` and `expected_state_version` with every action. Accepted command receipts survive restarts; retrying an accepted request returns the original result and a current snapshot without applying the action again. A stale action returns HTTP 409 with a current snapshot for review. A lost network response is retried once with the same ID. Older API clients may still omit both fields, but then receive neither safeguard.

The live endpoint is `/api/v1/sessions/{id}/live?after={sequence}`. It authenticates the first WebSocket message (`{"token":"..."}`), enforces same-origin browser connections and campaign membership, sends `session.sync` with replay events/current snapshot, then broadcasts `session.updated` and `presence`. Tokens never appear in the URL. Presence counts unique connected accounts on that session; several tabs from one account count once. Reconnect applies the authoritative snapshot and retains the server event log. Incoming updates wait while a form field is focused. This transport currently targets one Go server process.

After building `output/astralis-test.exe`, run `node scripts/check-live.cjs` with the same Playwright setup as above. It uses port 18082 and temporary storage to test two-member presence, live chat, draft preservation, a lost-response retry without rerolling dice, reconnect after a backend restart, route cleanup, mobile overflow, and polling fallback. Go tests also cover concurrent version conflicts, persisted deduplication, replay, membership, token revocation and origin checks. See [ASTRALIS_HANDOFF.md](ASTRALIS_HANDOFF.md) for remaining work and limits.

## Character control

In the table's **Characters** panel, the GM can select a party character, choose a campaign member under **Controlled by**, and save. Choose **Game master only** to clear an assignment. Invite players first; **Refresh players** loads members who joined after the panel was opened. Assignments are saved for that session and shared live.

Players can move, attack, investigate, or end a turn only for their assigned party characters, subject to the existing turn/range/action rules. Other tokens remain available for inspection. The GM retains control of every party character and scene-editing tools; hostile NPCs remain GM-controlled. Existing saved tokens without assignments default to GM-only control. Server checks apply to direct API requests as well as the interface, including a retry made after control was revoked.

`GET /api/v1/campaigns/{id}/members` returns campaign-scoped IDs, display names and roles for signed-in members. The GM submits `token.assign` through the normal session command endpoint with `token_id` and `controller_id` (empty to clear). The server validates the controller against campaign membership and publishes `token.assigned` with the updated snapshot. Global character-record CRUD/assignment and full character sheets remain future work.

`node scripts/check-character-control.cjs` uses temporary storage on port 18083 and three isolated browser contexts to verify assignment, GM override, allowed player actions, other-player restrictions, reassignment/clearing, forged-request rejection, restart persistence and desktop/mobile controls. Build `output/astralis-test.exe` first and use the Playwright setup above.

## Expanded tabletop workspace

Open a saved/demo table and choose the new **Table** inspector tab. The existing Actions, Characters, Tools and Log panels remain available.

| Section | Working features |
| --- | --- |
| Scenes | Create a blank scene or duplicate the current one; rename, switch and delete archived scenes. The whole table follows the GM. Maps, tokens, terrain, fog marks, drawings, grid settings and initiative are saved per scene. |
| Drawings | Line, rectangle, ellipse, cone, freehand, text and labeled pins; stroke color/width; map, token, GM-private and foreground layers; remove a drawing or undo the selected layer. |
| Tokens | Rotation, 1–4 cell sizes, six status markers, movement locks, initiative modifiers, layers, duplication and deletion. Dragging a token submits an authorized move; GM dragging uses placement. Larger footprints affect collision checks. |
| Art | Import PNG/JPEG images up to 2 MB into a ten-image library, apply artwork to a token or use it as a map. Unused library originals are GM-only. |
| Lighting | Optional square-grid token line of sight through assigned living characters, with a radius of 1–30 cells. Blocked terrain stops rays, including closed diagonal corners. Custom map images are masked before player delivery; hidden tokens and events are filtered on the server. |
| Turns | Server-resolved party initiative and start-in-order controls; pause/resume player movement. Hostile NPCs remain GM-controlled. |
| Chat | Speak as your account or an authorized character; Enter sends and Shift+Enter adds a line. `/roll`, `/r`, `/gmroll`, `/gr`, `/w gm`, `/w "Player name"`, `/w member-id`, `/em` and `/me`. Dice syntax is `NdS±bonus`, with server limits. |
| Journal | Create/edit/delete plain-text handouts, keep them GM-private or share them with players. |
| Music | GM sets a shared HTTP(S) soundtrack; listeners explicitly enable playback, control their own volume, and start at the current track position. Source/browser playback restrictions still apply. |
| Media | Explicitly join/leave a direct voice/video call, optionally include camera video, and mute/unmute the microphone. No device capture starts until Join and browser permission. Leaving/disconnecting stops capture. |
| Macros | Save personal dice macros, run public server-resolved rolls and delete definitions. Other accounts, including the GM, cannot read personal definitions. |
| Settings | Grid dimensions from 2–100 cells in either direction, ruler scale/units and orthogonal/straight-line/diagonal distance; share the last ruler measurement, zoom, fit and read-only generic player preview. Gameplay movement remains orthogonal pathfinding. |

Privacy is applied to snapshots, session listings, events, live sync/broadcast, stale-version responses and duplicate-command responses. Whisper content is delivered only to its sender and recipient; a GM does not automatically receive whispers between other players. GM-layer objects and private handouts are omitted from player data. If sight is enabled, players without an assigned living character see a dark map. The included Riverside artwork is also public as the landing-page sample; uploaded custom originals are not sent through the player snapshot when masked.

The lighting model uses opaque grid-cell terrain, doors/windows, token sight, optional light-dependent visibility and darkvision. Explored-map memory, blended colored illumination, hex grids, published compendiums/modules, cards, advanced dice expressions and production hosting/storage remain future work. Direct calling requires localhost or HTTPS and a network that permits peer connections; internet-wide relay/TURN support is not configured. The continuation below describes the latest character, crop and random-table tools.

The comparison used the supplied editor screenshot and Roll20's official [toolbar](https://help.roll20.net/hc/en-us/articles/360039674753-Toolbar-Overview), [layers](https://help.roll20.net/hc/en-us/articles/360039675053-Layers), [measurement](https://help.roll20.net/hc/en-us/articles/360039674913-Measure-Tool), [token](https://help.roll20.net/hc/en-us/articles/360039674573-Token-Features), and [chat](https://help.roll20.net/hc/en-us/articles/360039675093-Text-Chat) documentation. The authenticated editor was not available through a connected browser; no claim is made that its signed-in controls were inspected live.

After rebuilding the test executable, run `node scripts/check-table-workspace.cjs` with the Playwright setup above. It starts port 18084 with temporary storage and verifies the expanded UI, private content, persistence and desktop/390px layouts. Audio and voice/video checks use synthetic audio and fake browser devices; they do not access personal microphones/cameras or establish internet-wide call reliability. Backend tests additionally verify privacy across all read paths, masked map pixels, wall occlusion, historical-event visibility, token footprints and session-scoped media signaling.

## Character and GM toolkit — 14 November 2026

The **Table** panel now includes **Sheets**, **Environment**, **Quests**, and **Tables** alongside existing journals and scene tools.

- **Sheets:** GM-created system-neutral characters, assigned-player editing, private/shared visibility, attributes and background, inventory/spell/resource notes, plus structured entries with quantities and capacities. Prepared spells can spend uses; exhausted entries reject further use. Editing notes preserves tracked quantities. Sheet rolls use the existing server dice system and are public. Combat stats and spell effects remain manual.
- **Environment:** cell-based open/closed, locked and secret doors; transparent windows that still block movement; optional scene darkness, colored light-source markers with wall-aware illumination, and per-token darkvision (0–30 cells). Players can toggle visible unlocked doors next to their assigned living character. Close actions reject occupied doorways. Lighting requires token sight in **Lighting**. Colors identify light sources; illumination is binary rather than physically blended.
- **Map setup:** percentage crop bounds and quarter-turn rotation for imported custom maps. The server generates a new PNG, limited to 2048 pixels per side and 5 MB. Cropping replaces the current image; keep the original file to undo. Grid dimensions/scale remain under **Settings**. Doors and lights must fit when shrinking the grid.
- **Quests:** descriptions, objective checklists written with `[ ]`/`[x]`, planned/active/completed/failed status, and GM-private/shared visibility. The GM manages quest progress; players read shared quests.
- **Tables:** weighted random results, public/private tables and draws, draws with/without replacement, depletion checks and reset. The GM creates/edits/deletes and draws; players read shared tables/results.
- **Invitations:** configurable expiry and join limits, owner-only listing/revocation in **Tools**. New defaults are seven days and ten joins; existing saved legacy codes retain compatibility. Redeeming never overwrites an existing member's role or consumes another use for existing membership. Revoking a code prevents future joins; it does not remove existing members.

Sheets, quests and random tables belong to the saved session and survive scene switching. Environment data belongs to its scene. These features do not connect the older campaign-wide `/characters` records to scene spawning.

Research used official [Foundry actors](https://foundryvtt.com/article/actors/), [rollable tables](https://foundryvtt.com/article/roll-tables/), [journals](https://foundryvtt.com/article/journal/), [Quest Portal sheets](https://www.questportal.com/features/character-sheets), [Owlbear fog](https://docs.owlbear.rodeo/docs/fog/), [Roll20 lighting/doors](https://roll20partners.zendesk.com/hc/en-us/articles/10269638920599-Setting-Up-Dynamic-Lighting), and [Alchemy GM tools](https://help.alchemyrpg.com/en/articles/9714185-gamemaster-orientation). This is an Astralis implementation of useful workflows; licensed content and full competitor parity are not included.

Run `node scripts/check-adventure.cjs` after rebuilding `output/astralis-test.exe`. It uses temporary storage on port 18085 and isolated browser contexts to verify assigned sheets, tracked spell use, shared quests, weighted draws, doors/lights/darkvision, cropping, invitation revocation, restart persistence and desktop/mobile layouts. See the final handoff section for confirmed checks and remaining work.

## Map interaction polish

The tabletop now has a compact tool rail/inspector, an on-map drawing flyout with quick color/width/layer controls, smooth freehand strokes, live shape/token drag previews, shared-token movement transitions, and a floating party turn tracker. GM tracker controls edit initiative scores/order and roll/sort/start/advance turns without granting extra actions. Existing drawings and two-click tools remain compatible.

Hold **Space** or the **middle mouse button** while dragging to pan. **Ctrl/Command + wheel** zooms around the pointer. **Y** toggles the turn tracker and **W** toggles the inspector when canvas focus is outside form controls. Ordinary wheel scrolling and the existing zoom/fit buttons remain available. Escape cancels unfinished drawing gestures. Reduced-motion settings disable token movement animation.

`node scripts/check-canvas.cjs` tests these interactions against a temporary database on port 18086, including live previews, precise stored strokes, anchored zoom, tracker editing, shared movement animation, desktop/mobile layouts and restart persistence. Rebuild the test executable first. The latest verified development runtime is `output/astralis-polished.exe`; see the handoff for process/health evidence and reference-access limitations.

## Contextual editor controls

The rail now groups secondary tools under **More** and provides direct **Dice**, **Turns**, **Scenes** and **Layer** access. Double-click or right-click a token for its contextual panel. GMs can change HP (absolute or `-5`/`+5`), rotation, status markers, movement lock and token layer there; player panels preserve existing permissions. Quick dice rolls and scene switching use the saved API. Turn tracker entries support drag reordering as well as keyboard buttons. The Table panel's section selector provides direct navigation to all workspace tools.

`node scripts/check-editor.cjs` verifies these workflows, player restrictions and restart persistence using disposable storage on port 18087. Build `output/astralis-test.exe` first. The latest verified local runtime is now **`output/astralis-editor.exe`**; reload the browser after restarting to load the new assets.

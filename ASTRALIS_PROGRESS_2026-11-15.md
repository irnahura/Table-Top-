# Astralis progress and continuation checkpoint

Saved: 15 November 2026.

Project: `C:\Users\Mayank\OneDrive\Documents\GitHub\Table top`.

This document consolidates the current feature set, the latest completed work, validation evidence and next steps. [ASTRALIS_HANDOFF.md](ASTRALIS_HANDOFF.md) retains the detailed implementation history and protocol notes. [README.md](README.md) contains running instructions. No additional application changes were made while saving this checkpoint.

## Current project

Astralis uses a Go backend in `cmd/astralis/` and a framework-free frontend in `frontend/`. Saved local accounts, campaigns and sessions live in `data/astralis.json`. The older Python backend and root `astralis.exe` are not the current implementation.

Shared actions follow the authoritative command pipeline: authentication and membership checks, actor/GM permissions, command IDs, state-version checks, validated state changes, event recording, JSON persistence and live broadcast. Player snapshots and events use a common privacy projection.

## Completed feature inventory

| Area | Current implementation |
| --- | --- |
| Accounts and campaigns | Registration, sign-in, demo tables, campaign library, campaign/session creation and resume, persisted authentication and membership checks. |
| Invitations | Expiry, usage limits, owner listing/revocation and redemption; legacy invitation compatibility. |
| Multiplayer | Authenticated WebSocket updates, presence, reconnect/replay, safe duplicate-command retries, conflict handling and polling fallback. |
| Character control | GM assignment of session party actors to campaign members; server-enforced movement, interaction, attacks and turn permissions. |
| Basic encounter | Server-resolved party turns, movement/path budgets, terrain and occupancy checks, basic attacks, HP and encounter history. |
| Maps and scenes | Image import, blank/sample backgrounds, art library, scene creation/duplication/switching, grid dimensions/units/metrics, percentage cropping and quarter-turn image rotation. |
| Drawings and layers | Line, rectangle, ellipse, cone, smooth freehand, text and pins; map/token/GM/foreground layers; colors, widths, undo and removal. |
| Editor interaction | Grouped tool rail, drawing flyout, gesture previews, token dragging, animated accepted movement, Space/middle-button pan, pointer-anchored zoom and inspector collapse. |
| Token editor | Context menu/double-click actions, HP adjustments, rotation, movement locks, status markers, layers, portraits, size, duplication and deletion. |
| Initiative | Party initiative rolls, score edits, sorting, keyboard/drag reordering, floating tracker, shared visibility and preservation of active action budgets. |
| Sheets | Private/shared sheets with assignment, editable notes/stats, inventory/spells/resources and tracked entries with server-validated use/preparation/depletion. |
| Journals and quests | Private/shared text handouts, quest descriptions/objectives/status and GM editing. |
| Random tables | Weighted outcomes, replacement/no-replacement draws, reset and private/shared definitions/results. |
| Chat and dice | Account/authorized-character identity, emotes, whispers, public/GM/self-only dice, simple formulas and personal macros. |
| Audio and calling | Shared audio URL/playback intent with local opt-in; authenticated direct WebRTC signaling, microphone/camera controls and call cleanup. |
| Visibility | Assigned-token sight, wall-aware light sources, darkness requirement, darkvision, cell doors/windows, secret/locked doors and server-masked custom map images. |
| Game profiles and rules | Header game selector, tailored checks, manual sheet templates, scale presets, searchable reference directory and bundled open D&D SRDs. |
| Ruler | Five shapes, center/free positioning, drag/click-click gestures, local preview, optional sharing, explicit sharing and local clearing. |
| FX | Burst, aimed breath, aimed beam and nova in eleven colors, synchronized visual particles, recent-event playback and reduced-motion support. |

These are working baseline features. They do not imply full Roll20 parity or complete automation of the selected published game.

## Latest completed work: game profiles, rules, ruler and FX

The latest request supplied Roll20 screenshots for dynamic lighting, effects, dice and measurement, plus the Roll20 editor URL. Existing lighting, doors, rulers and dice were extended instead of recreated.

### Header and game tools

- **Game to play** offers Astralis Basic, D&D 5e 2014, D&D 5e 2024, Warhammer 40,000, Pathfinder 2e, Fate Core and Generic percentile.
- The GM's selection persists across restart and scene switching and updates connected players. Players can read the selected profile but cannot change it.
- Selection preserves token values, encounter state and current scene scale. Applying a profile's suggested scale is an explicit separate action.
- **Rules** provides quick references, book search, system-specific checks and private manual sheet templates. Form preferences survive rerenders per profile.
- D&D supports ordinary d20 ability checks with normal/advantage/disadvantage. Pathfinder supports four degrees of success and natural-face adjustments. Warhammer supports up to 100 d6 with a 2+–6+ threshold. Fate rolls 4dF plus modifier. Generic percentile rolls 1d100 at or below a target.
- The quick dice tray and system checks offer All players, You and GM, and Only you. Another player's self-only results are also hidden from the GM.

### Books

- Included unchanged [SRD 5.1 PDF](frontend/rules/dnd-srd-5.1.pdf), 403 pages, and [SRD 5.2.1 PDF](frontend/rules/dnd-srd-5.2.1.pdf), 364 pages.
- Both retain their publisher legal pages and use CC BY 4.0 with [attribution](frontend/rules/ATTRIBUTION.txt). The [manifest](frontend/rules/manifest.json) records sources, sizes and SHA-256 checksums.
- Other references link to publisher or publisher-endorsed resources, including Warhammer downloads, Archives of Nethys and Fate SRD. Commercial books and codexes are not bundled.
- The Warhammer profile is an edition-neutral dice helper. Users must check the edition and revision of downloaded rules.

### Map tools

- Ruler shapes: line, square, circle, cone and a one-cell-wide beam. Circle labels report radius; cone labels report depth. Square uses opposite corners and labels the distance between endpoints.
- Measurements use the scene metric and scale. Endpoints stay inside the map; template geometry can extend beyond it. Measurement does not spend movement or apply damage.
- FX modes: burst, breath, beam and nova. Elements: fire, charm, acid, death, holy, blood, frost, slime, smoke, water and magic. Effects are visual; damage and other consequences remain manual.
- Spatial events carry their originating scene. Switching scenes clears local ruler geometry and does not draw another scene's effects.
- With token sight active, GM FX/measurements are GM-only. A player's shared measurement requires visible endpoints and is delivered only to that player and the GM.
- **Light** exposes token sight and darkness together, with access to the existing doors/lights/darkvision editor. Lighting still uses binary illumination and colored source markers.

### Files changed or added for this work

| Files | Purpose |
| --- | --- |
| `cmd/astralis/systems.go` | Profile selection, system checks, private quick dice and extended FX/ruler command handling. |
| `cmd/astralis/systems_test.go` | Rules boundaries, permissions, privacy, rejected-state immutability and persistence tests. |
| `cmd/astralis/table.go` | Command dispatch, persisted profile field and profile preservation across scenes. |
| `frontend/systems.js`, `frontend/systems.css` | Header selector, rules/checks, dice audiences, ruler/FX/lighting flyouts and responsive layout. |
| `frontend/canvas.js`, `frontend/index.html` | Integration and loading of the new module. |
| `frontend/tabletop.js` | Shape-aware sharing and legacy-renderer fallback without duplicate FX/ruler overlays. |
| `frontend/rules/` | Bundled open PDFs, attribution and source/checksum manifest. |
| `scripts/fetch-open-rules.cjs` | Verify existing PDFs or download with `--refresh`, then write the manifest. |
| `scripts/check-systems.cjs` | Real-backend GM/player browser journey on disposable storage, port 18088. |
| `scripts/check-table-tools.cjs`, `scripts/check-table-workspace.cjs` | Existing journeys adjusted to close ruler options before clicking obscured map cells. |
| `README.md`, `ASTRALIS_HANDOFF.md` | Usage, implementation history, validation and remaining limits. |

## Validation completed during implementation

- `go test ./...` and `go vet ./...` passed.
- JavaScript syntax checks and `node scripts/check-frontend.cjs` passed.
- All nine browser journeys passed during the implementation: `check-systems.cjs`, `check-editor.cjs`, `check-canvas.cjs`, `check-table-tools.cjs`, `check-table-workspace.cjs`, `check-adventure.cjs`, `check-live.cjs`, `check-character-control.cjs` and `check-cinematic.cjs`.
- The new journey covered live header selection, system checks, populated sheet templates, book search and PDF serving, private dice, ruler geometry/free coordinates/sharing/clear, aimed shared FX, lighting, scene isolation, restart persistence and desktop/390px layouts.
- Deterministic Pathfinder tests cover difficulty boundaries and natural 1/20 degree changes. Other backend coverage includes GM permissions, private-roll audiences, D6 pools, Fate values, measurement distances and sight restrictions.
- Screenshots were reviewed and corrected for mobile header/panel layout, duplicate ruler labels, stale shared ruler overlays and gesture click suppression.
- Test accounts and interactions used disposable databases. Saved `data/astralis.json` was not used for test data.
- Browser connection discovery returned no available browser; installed isolated Playwright supplied local verification. The signed-in Roll20 editor was not inspected, and no personal browser profile was accessed.
- These checks were completed in the preceding implementation work; documentation saving did not rerun the application suites.

Latest screenshots: [desktop](output/screenshots/astralis-systems-desktop.png) and [mobile](output/screenshots/astralis-systems-mobile.png).

## Local runtime checkpoint

Last verified during the preceding implementation turn:

- URL: <http://localhost:8080/>.
- Executable: `output/astralis-systems.exe`, started hidden from the project root.
- PID at verification: `28908`; recheck before managing the process.
- Storage: `data/astralis.json`.
- Health: `/health` returned `astralis-go / ok`.
- `/rules/dnd-srd-5.2.1.pdf` returned HTTP 200.
- Saved data's checksum was unchanged across the runtime refresh.
- Logs: `output/astralis-systems.stdout.log` and `output/astralis-systems.stderr.log`.

The previous Go-run Astralis process was stopped only after verifying its identity, health and exact serving of this project's source asset. Older executables remain on disk. Runtime availability was not rechecked solely for this documentation update.

## Remaining work and suggested continuation order

1. **Sheet/token linkage:** connect session sheets to map actors through explicit GM assignment; define ownership, stat synchronization and deletion behavior before implementing deeper system automation.
2. **Published-system mechanics:** add actual system-specific movement, combat, initiative and resource rules incrementally. The current profile selector configures reference tools; the encounter engine still uses Astralis rules. D&D spell effects, Warhammer rerolls/wound/save chains/army phases/missions and other system exceptions remain manual.
3. **Compendiums:** searchable structured open-license rules, supported imports and clearly scoped commercial licensing. A universal book catalog and full-text PDF search are not implemented.
4. **Advanced maps and lighting:** interactive image-to-grid calibration, hex grids, arbitrary wall geometry, explored-map memory and blended colored light.
5. **Editor depth:** drawing selection/resizing, custom/NPC initiative entries, cards, advanced dice expressions, programmable FX and journal organization.
6. **Production foundations:** propagate persistence failures, transactional storage/PostgreSQL migrations, durable asset storage, membership removal/roles, account recovery/rotation/rate limits and observability. JSON storage, inline images and retained full snapshots remain scale limits.
7. **Internet calling:** hosted discovery and STUN/TURN support. Prior synthetic media tests do not establish real-device or internet-wide call reliability.

## Resume instructions

Read this checkpoint, then the latest sections of `ASTRALIS_HANDOFF.md`, and inspect current source and working-tree changes. Preserve existing edits and saved campaigns. Historical `design/` deletions and many untracked implementation directories predate this checkpoint; do not reset or stage them broadly.

Verify the listener and executable before restarting anything. If no development server is running, start from the project root:

```powershell
go run ./cmd/astralis
```

For implementation verification:

```powershell
go test ./...
go vet ./...
node scripts/check-frontend.cjs
go build -o output/astralis-test.exe ./cmd/astralis
$env:PLAYWRIGHT_MODULE='C:\Users\Mayank\AppData\Roaming\Python\Python314\site-packages\playwright\driver\package'
node scripts/check-systems.cjs
```

Confirm that the Playwright installation still exists before using the example override. Browser journeys launch their own temporary-data servers; do not point them at saved campaign storage.

No commit, push or deployment was made. This checkpoint is saved locally for continuation.

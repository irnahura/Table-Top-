# Astralis guide and usability checkpoint

Date: 25 November 2026.

## Request and reference access

The user asked for guides to the website and all tools, more user-friendly features, and continuation using https://www.youtube.com/watch?v=ueflwPxQTRc&t=26s as a reference.

The video watch page and oEmbed request could not be retrieved. Browser setup returned `No browser is available`; recovery discovery returned `[]`. The video was not watched, and no transcript or frame analysis was available. Do not present a complete video-derived feature inventory as verified. The current source, latest project handoff and official Roll20 toolbar documentation informed this pass: https://help.roll20.net/hc/en-us/articles/360039674753-Toolbar-Overview . Installed isolated Playwright provided local browser verification.

## Implemented

- Public handbook at `/#/help`: 31 original guides, full-text search across instructions, player/GM filters, empty-result feedback and direct URLs for individual guides.
- Coverage: accounts/demo/campaigns/invitations; navigation and map controls; movement/combat/ownership; token editing; scenes; map import/crop; art; drawings/pins/layers; blocked terrain/fog planning; ruler; ping; FX; initiative; dice audiences; system checks/reference books; sheets/inventory/spells/resources; token sight/darkness; doors/windows/lights; chat/whispers; journals/notes; quests; weighted tables; macros; music; calling; settings/player preview; log/reconnect/troubleshooting.
- In-table **Help & tools**, rail Guide interception, contextual **How to use this** in all five inspector panels and the selected Table section.
- **Open this tool** dispatches to existing controls without submitting gameplay commands. It opens collapsed inspectors and mobile panels, and reveals secondary tools when necessary. GM launch controls are unavailable to players and all launch controls are disabled during player preview.
- **?** opens help; **Ctrl/Command + K** opens the tool finder. Text inputs retain normal typing. A native modal provides focus containment, Escape dismissal and focus return; the modal lives outside the table render tree.
- Six-step walkthrough highlights the rail/map/Actions/Characters/Table/help controls, with Back/Next/Finish/close. Inspector steps reveal the relevant local panel, including on mobile. No walkthrough step issues a saved command.
- Manually checked, role-specific preparation checklists in table help. Progress is local browser storage keyed by account, campaign and role, with unavailable-storage feedback. It is not synced or automatic completion tracking.
- Handbook links in Learn, mobile navigation, campaign sidebar, footer, getting-started guide and glossary.
- Corrected outdated FAQ/glossary claims that sheets, lighting, journals and published-game reference tools were absent. Complete published-system mechanics remain manual as documented.

## Files

- `frontend/learning.js`: single shared guide dataset, handbook rendering, contextual modal, tool launch, tour and local checklist.
- `frontend/learning.css`: scoped desktop/mobile styles matching the existing dark/gold interface.
- `frontend/app.js`, `frontend/index.html`: module loading, route/navigation and render hooks; corrected help copy.
- `scripts/check-learning.cjs`: real-backend isolated guide journey on port 18089.
- `README.md`: current usage and verification instructions.

## Confirmed validation

- JavaScript syntax checks and `node scripts/check-frontend.cjs` passed.
- Built `output/astralis-test.exe` from current Go source.
- `check-learning.cjs` passed: all 31 public guide routes, search/filter/no-results, no public launch actions, persisted checklist across reload, focus return, keyboard shortcuts and normal text entry, tour, every available launch target, role restrictions, reopening collapsed inspector, desktop/390px layouts and unchanged saved encounter state.
- Existing `check-editor.cjs`, `check-systems.cjs` and `check-cinematic.cjs` passed, including real gameplay mutations, private dice, lighting, ruler/FX, persistence, landing/auth/campaign flows, map controls, movement/combat and desktop/mobile panels. The cinematic journey uses API fixtures; the editor/systems journeys use real disposable backends.
- Help screenshots under `output/screenshots/`: `astralis-help-mobile.png`, `astralis-help-table-mobile.png`, `astralis-help-table-desktop.png`. Table modal screenshots were visually reviewed; the tour action was moved above long instructions to improve discoverability on small screens.
- Tests used temporary storage and isolated browser contexts. Saved `data/astralis.json` was not used for test accounts. No backend source or game-rule changes were made.
- The Go test/vet suites were not rerun for this frontend-only change. Existing browser regressions above were run; do not claim all nine historical journeys were rerun.

## Local runtime

No listener was present on port 8080. Built `output/astralis-learning.exe` and started it hidden from this project root with port 8080 and explicit `data/astralis.json` storage. PID at verification: **24800**. `/health` returned `astralis-go / ok`, and the HTTP-served `learning.js` exactly matched current source. The updated handbook is available at http://localhost:8080/#/help . Logs are `output/astralis-learning.stdout.log` and `.stderr.log`. Recheck process/listener identity before managing this runtime in a future session. No existing process was stopped.

## Continuation boundaries

This pass improves onboarding, discovery and guidance across the existing features. It does not establish implementation of every feature in the inaccessible video. Full sheet/token linkage, published-system automation, compendiums, advanced lighting/maps and production foundations remain open as described in the 15 November checkpoint. Keep saved campaigns and unrelated edits intact. No commit, push or deployment was requested or performed.

For further changes, maintain guide content beside the affected tool so help does not drift again. Verify selected tool names, player access and examples against actual source and browser behavior.

## Continuation: Roll20 adventure guide audit, course and linked sheets

The user supplied Roll20's written adventure-creation overview and asked to compare all functionality, learn from the guide structure, implement improvements and use available MCPs. Reviewed the overview, eleven linked written chapters and resources. Canonical article URLs recovered the non-combat/reward/resources pages when shortened links failed. This was a written-guide review; no embedded-video playback is claimed.

Full findings and source links: [ASTRALIS_ROLL20_ADVENTURE_AUDIT_2026-11-25.md](ASTRALIS_ROLL20_ADVENTURE_AUDIT_2026-11-25.md). It distinguishes working baseline features, manual workflows and missing capabilities.

### Implemented

- Added `frontend/adventure-guide.js` / `.css`, integrated through the existing learning module and Learn navigation. The handbook now has **32** entries.
- `/#/help/adventure` provides an eleven-lesson preparation course, desktop table of contents, mobile chapter picker, previous/next links, local completion checkmarks, current-table content counts and explicit limitations per lesson.
- `/#/help/adventure/resources` provides seven original Astralis text templates: outline, location key, combat encounter, obstacle/clue/puzzle worksheet, rewards/aftermath, player briefing and session-zero agreement. Copy, individual download and complete-pack download work. No Roll20 story/map or commercial content was imported.
- GM template preparation stages an explicitly private Journal draft, routes to the existing form and requires review/save. An unfinished Journal draft is preserved instead of overwritten. Preparing a template does not save game state by itself.
- Added optional `Actor.sheet_id`, GM-only `sheet.link` and `sheet.place`, Map tokens controls on opened sheets and Open linked sheet in token contextual menus. New party tokens inherit sheet control once; existing token stats/control are not changed by linking. Hostile NPCs remain GM-controlled. Stats and HP remain separate from sheet text.
- Sheet links persist across scenes/restarts. Deleting a sheet clears active and archived links without deleting tokens. Player projection removes references to private inaccessible sheets. Link/placement history is GM-only because it can reference hidden tokens. Authorization is checked before command-receipt lookup.
- New-token defaults prefer an unoccupied central cell, avoiding floating tools over cell 0,0. Server validation remains authoritative.
- Corrected public guide/course spacing below fixed navigation; added a browser check for mobile header overlap. Reviewed desktop/mobile screenshots and regenerated them after adjustments.

### Validation completed for the continuation

- `go test ./...` and `go vet ./...` passed. New `sheet_tokens_test.go` covers link/placement permissions, invalid-input immutability, inherited versus unchanged control, private link projection, hidden-token event privacy, scene/restart persistence, unlinking and deletion cleanup.
- Frontend contracts and changed JavaScript syntax checks passed. Current test executable was rebuilt.
- All **eleven** browser journeys passed: `check-adventure-course`, `check-learning`, `check-adventure`, `check-editor`, `check-systems`, `check-canvas`, `check-table-tools`, `check-table-workspace`, `check-character-control`, `check-live`, and `check-cinematic`. The new course journey was rerun after the final header-spacing test adjustment and passed.
- `check-adventure-course.cjs` uses temporary port **18090** and disposable storage/accounts. It verifies all lesson routes, progress, template downloads, mobile navigation, explicit private drafts, draft preservation, linking/placement, token-to-sheet opening, player restrictions and restart persistence. `check-learning.cjs` now verifies 32 handbook entries.
- Tests did not use saved `data/astralis.json` for accounts or actions. The cinematic suite uses API fixtures; other listed integration journeys use disposable real backends. Synthetic voice/video testing does not establish internet-wide call reliability.
- Playwright MCP successfully inspected the actual local course and mobile chapter navigation/contextual help. The Browser skill runtime had no connected browser. Plugin Management search for Framer/Selenium returned no plugin result; no Framer operation, Selenium verification or plugin installation was performed. Other plugins may be available in the plugin directory.
- Screenshots: `output/screenshots/astralis-course-desktop.png`, `astralis-course-mobile.png`, and a Playwright MCP capture `astralis-course-mcp-mobile.png`. The standalone course screenshots include the corrected navigation spacing; the earlier MCP capture predates the final scroll-position/layout check.

### Current local runtime (supersedes earlier runtime section)

Built `output/astralis-adventure.exe`. Verified the 8080 listener was this project's `output/astralis-learning.exe` before stopping only that process. Started the new build hidden from this project root on port **8080**, with explicit `data/astralis.json` storage. PID at verification: **9440**. `/health` returned `astralis-go / ok`, the HTTP-served `adventure-guide.js` matched source, and the saved database checksum was unchanged across the refresh. Logs are `output/astralis-adventure.stdout.log` and `.stderr.log`. Recheck listener/process identity on future continuation. Reload the browser to load new scripts/styles.

No commit, push or deployment was made. Full published-system automation, compendium imports, interactive grid alignment, NPC/custom initiative, manual reveal masking and independent GM/player scene navigation remain gaps. The audit records the broader backlog and the exact scope of the implemented session sheet links.

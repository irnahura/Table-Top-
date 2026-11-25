# Adventure-creation workflow audit

Date: 25 November 2026. Scope: compare the linked Roll20 written guide with current Astralis source, implement a connected preparation course and close the sheet-to-token navigation/placement gap. This is a workflow comparison, not a claim of full Roll20 parity.

Reference: [Creating your adventure — Overview](https://help.roll20.net/hc/en-us/articles/29257395844247-Creating-your-adventure-Overview), its eleven chapter links and its resources page. The written pages were retrieved; no embedded video playback is claimed. Canonical chapter URLs were used when shortened resource/non-combat/reward links failed.

## Coverage after this continuation

| Reference workflow | Astralis implementation | Status and boundary |
| --- | --- | --- |
| 1. Game lobby | Account registration/sign-in, campaign creation, launch/resume, invitations, live profile selection. | Working baseline. Campaign cover editing, modules and email invitations remain absent. |
| 2. Tabletop basics | Map rail, select/pan, drawing tools, layers, dice, inspector sections, scenes, chat, art, journals, music, macros and tables. | Working baseline. Separate GM browsing and player scene locations, folders, card decks and 3D dice are absent. |
| 3. Adventure outline | GM-private plain-text handouts; new original outline, briefing and session-agreement templates. | Working manual preparation. Templates stage reviewable drafts; rich text and folders remain absent. |
| 4. Map creation | Included Riverside artwork, blank grid/map-layer drawing, PNG/JPEG import and reusable art. | Partial. Use an external map editor for custom image production. No dungeon generator or Dungeon Scrawl/One Page Dungeon project import. |
| 5. Map setup | Named/duplicate scenes, grid dimensions/scale/metric, crop/rotation, GM-private text/pins, blocked terrain, doors, token sight and lights. | Partial. Interactive pixel/grid alignment and hex grids remain absent. GM scene switching moves the whole table. |
| 6. Combat setup | Manual token creation, sheet-to-token placement/linking, GM-layer staging, duplicate tokens, HP/status editing and party initiative. | Partial. No published-system encounter calculator, monster compendium drag/drop, NPC initiative or complete system combat automation. |
| 7. Non-combat encounters | Private notes/pins/drawings, shared handouts, quests, system check helpers and original obstacle worksheet. | Working manual preparation/adjudication. No automatic trap, puzzle, save or damage triggers. |
| 8. Rewards | Inventory/resource entries and quantities, weighted tables, quests and original rewards/aftermath worksheet. | Manual bookkeeping. No automatic loot distribution, item effects or progression calculation. |
| 9. Heroes | Assigned private/shared sheets, tracked entries, explicit links to existing tokens, placement from sheets and token-menu sheet opening. | Session-level reference links implemented here. No full character builder, licensed system sheet or automatic sheet/stat synchronization. |
| 10. Live play | Authoritative basic party movement/combat, chat/whispers, dice audiences, ruler shapes, FX, token sight, reconnect/replay and saved resume. | Partial. No ruler waypoints, independent GM/player scenes, manual hide/reveal mask or published-system action economy. Fog planning remains shared. |
| 11. Group setup | Expiring/limited/revocable invitation codes, player/token assignments, optional music/calling and a session-agreement worksheet. | Invite a known group. No public matchmaking, calendar scheduling, membership removal or internet TURN relay. |
| Resources | Seven original Astralis text templates; copy, individual download, complete pack download and GM-private Journal draft preparation. | Working. The templates do not import a module or change encounters automatically. No Roll20 story, map or commercial content was copied. |

## New connected course

Open `/#/help/adventure`. Eleven lessons have a desktop chapter list, mobile lesson picker, previous/next navigation, practical steps, completion criteria, tool links and explicit feature boundaries. `/#/help/adventure/resources` provides original worksheets and the Riverside “Missing Lantern” example. Completion checkmarks are local learning notes keyed by account/campaign; they are separate from saved game state.

Current-table preparation details show actual saved counts of visible sheets, links, assigned living party actors and handouts. They do not infer that an adventure is complete or safe to run from those counts alone.

Template actions open the existing Journal form with sharing off. The user must review and save. Existing unfinished Journal drafts are preserved, and the user is told to finish them before staging another template.

## Sheet/token contract

- `Actor.sheet_id` is optional and defaults to no link in older saves. A sheet can have several tokens, including across scenes.
- GM-only `sheet.link` accepts `sheet_id`, `token_id`, and boolean `unlink`. Linking replaces that token’s reference; unlink requires a matching reference. Existing stats and control are unchanged.
- GM-only `sheet.place` accepts `sheet_id`, `name`, `hp`, integer `x`/`y`, boolean `hostile`, and layer `tokens` or `gm`. It reuses token-creation limits, occupancy, terrain and input validation. New party tokens inherit the sheet controller once; hostile NPCs remain GM-controlled. Basic defaults are movement 6, attack +3, damage 4. Sheet text is not parsed.
- Sheet reassignment does not silently reassign previously placed tokens. Use Characters to manage their control. HP and combat state remain independent per token.
- Players receive a link only when the referenced sheet is visible to that account. GM-layer tokens and placement/link history stay private. Existing common projection applies across snapshots, replay, live updates and command responses.
- Deleting a sheet clears its references in active and archived scenes without deleting tokens. Links survive scene changes, reload and server restart.
- A visible linked sheet opens from the token contextual menu, including for authorized players. GM linking/placement controls are absent from player sheets.

## Tools and evidence

Playwright MCP successfully inspected the live public course, counted its eleven lessons, navigated the mobile picker at 390px and opened contextual guide help. The Browser skill’s runtime discovery returned no connected browser; direct Playwright MCP was used after that check. Plugin Management search for Framer/Selenium returned no available plugin result; no installation, Framer action or Selenium test is claimed.

Backend tests cover permissions, rejected-state immutability, valid/invalid linking, placement, inherited versus unchanged authority, private reference projection, hidden-token event history, scene/restart persistence and deletion cleanup. The new real-backend course browser journey covers public lessons, progress, downloads, private draft preparation, unsaved draft preservation, linking/placement, player access, token-to-sheet opening and persistence. Tests use disposable accounts and databases.

## Remaining implementation priorities

1. Define structured game-specific sheet fields before adding optional explicit stat synchronization.
2. Interactive image-to-grid alignment and ruler waypoints.
3. Structured open-license compendium records and supported imports.
4. NPC/custom initiative and incrementally scoped published-system mechanics.
5. Separate GM scene preparation/player scene control and real manual reveal masking.
6. Existing production persistence, auth lifecycle, roles, hosting and call-relay work.

These are identified gaps, not work reported as complete by this continuation. The working course is designed around capabilities that actually exist.

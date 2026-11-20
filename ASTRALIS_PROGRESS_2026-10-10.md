# Astralis Progress Handoff — 2026-10-10

## Completed today

The Astralis project was continued from the latest design archive and backend implementation plan. The main runtime was moved from the Python prototype to Go because Go is installed and .NET is not available.

Latest design source of truth:

- `Astralis-Game-Table/astralis-home-v2.html`
- `Astralis-Game-Table/astralis-onboarding-v2.html`
- `Astralis-Game-Table/astralis-tabletop-v2.html`
- `Astralis-Game-Table/DESIGN-MANIFEST.json`
- `Astralis-Game-Table/DESIGN-HANDOFF.md`
- `Astralis-Game-Table/Astralis_Open_Design_Spec_v2.md`
- `Astralis-Game-Table/brand-spec.md`

## Go backend

Added `go.mod` and `cmd/astralis/main.go`.

The Go service provides:

- HTTP server on port `8080`.
- `GET /health` health check.
- Static frontend from `frontend/`.
- Latest design/assets under `/design/`.
- JSON persistence in `data/astralis.json`.
- Registration and login.
- Campaign creation and listing.
- Campaign invitations and redemption.
- Character creation.
- Map creation/configuration.
- Session creation.
- Session snapshots.
- Event replay using `?after={sequence}`.
- Server-authoritative move, interaction, attack, and end-turn commands.

Gameplay rules validate map bounds, blocked cells, movement budget, active turns, interaction range, attack range, hostile targets, dice, damage, and state changes. The client does not provide final position, dice, damage, or turn results.

Accepted commands increment session sequence and state version and persist event payloads with type, actor, and timestamp.

## Frontend

Added:

- `frontend/index.html`
- `frontend/styles.css`
- `frontend/app.js`

The frontend includes the Astralis neutral/yellow/cyan/coral visual language, responsive shell, overview dashboard, campaign library, registration, campaign creation, map/session setup, functional tabletop tokens, initiative panel, selected-character panel, move/attack/interact/end-turn controls, snapshot refresh, and error/status messaging.

The Go server maps `/design/astralis-tabletop.html` to the latest `astralis-tabletop-v2.html`.

## Documentation and startup

`README.md` now documents the Go runtime. Start with `go run ./cmd/astralis` or run the compiled `.\\astralis.exe`, then open `http://localhost:8080/`.

## Validation completed

`gofmt -w cmd/astralis/main.go` and `go build ./...` passed.

The live smoke test passed:

- Health response: `ok`.
- Frontend response: HTTP `200`.
- Registration and login.
- Campaign creation.
- Map creation.
- Session creation.
- Authoritative movement command.
- Event sequence `1`.
- Position changed from `(2,2)` to `(3,2)`.

The first command test found a double-body-read bug; it was fixed and the complete test passed afterward.

## Python status

The earlier Python prototype remains in `backend/` as optional reference code and a possible future LLM-connectivity adapter. It is not required to run the main application.

Python/LLM services must remain non-authoritative. They may propose dialogue, narration, summaries, or intent, but must not decide movement, permissions, dice, damage, turns, or persistent state.

## Next continuation steps

1. Add WebSocket transport for presence and live event broadcast.
2. Add `command_id` deduplication and expected state-version checks.
3. Add complete session membership and character ownership checks.
4. Add invitation expiry, revocation, and usage limits.
5. Replace JSON persistence with PostgreSQL migrations.
6. Replace development SHA-256 passwords with Argon2id or bcrypt.
7. Add secure cookie/token rotation and rate limiting.
8. Add map image upload, grid calibration, collision annotation, and asset validation.
9. Add proper map rendering from uploaded assets.
10. Add reconnect-from-sequence behavior.
11. Add frontend character, invitation, map-editor, chat, quest, inventory, and loading surfaces as separate routes.
12. Add unit, integration, security, contract, and browser E2E tests.
13. Add Docker Compose with Go API and PostgreSQL.
14. Add structured logs, request IDs, metrics, and readiness checks.

## Repository note

Pre-existing deletions under the old `design/` directory were left untouched. The current implementation uses the `Astralis-Game-Table/` design archive as its source of truth.

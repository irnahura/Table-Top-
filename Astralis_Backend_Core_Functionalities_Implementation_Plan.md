# Astralis — Backend & Core Game Functionality Implementation Plan

**Status:** Ready for implementation  
**Goal:** Build a secure, persistent, multiplayer RPG backend and deliver one complete playable vertical slice.

## 1. Objective

The design phase is complete. Implementation should focus on a working end-to-end loop before expanding the feature set.

The first playable milestone must let users:
1. Register and sign in.
2. Create a campaign and invite another player.
3. Create/select a character.
4. Upload a map image and configure its grid and basic collision.
5. Start and join a multiplayer session.
6. See both players and their characters.
7. Move a character using server-validated commands.
8. Interact with a basic NPC/object.
9. Start a turn-based encounter.
10. Resolve a basic attack using server-side rules and randomness.
11. Persist the result and recover it after reconnect.

## 2. Non-Negotiable Architecture

### Server authority
The server is authoritative for position, movement legality, collision, turns, attacks, dice, damage, resources, inventory, encounter state, permissions, and world mutations. The client sends **commands**, never final outcomes.

Incorrect: “set character position to X/Y.”  
Correct: “request movement to this target”; the server validates and calculates the result.

### AI is not authoritative
AI may later propose dialogue, narration, summaries, or structured intent. It must not directly write game state or decide authoritative damage, dice, movement, permissions, or combat legality. The MVP must work when AI services are unavailable.

### Modular monolith first
Use one backend application and one PostgreSQL database. Do not begin with microservices, Kubernetes, a message broker, or distributed simulation. Extract components only when measured requirements justify it.

### Keep the engine independent
Rules must be testable without HTTP, WebSockets, database connections, React, PixiJS, or an AI provider.

## 3. Recommended Stack

| Layer | MVP choice | Purpose |
|---|---|---|
| Web app | Next.js + TypeScript | UI and application pages |
| World renderer | PixiJS + WebGL | Isometric map, tokens, camera, effects |
| API | FastAPI + Python | HTTP endpoints and session lifecycle |
| Realtime | WebSockets | Session commands, events and presence |
| Game engine | Pure Python package | Movement, collision, dice, turns and combat |
| Database | PostgreSQL | Durable application and game state |
| Asset storage | S3-compatible object storage | Maps, textures and character assets |
| Local environment | Docker Compose | Repeatable development setup |
| Tests | pytest + frontend test tooling | Unit, integration, contract and E2E tests |

If the existing repository already uses a working stack, inspect it before changing it. Preserve working infrastructure and adapt these module boundaries rather than rewriting without evidence.

Do not add Redis, a vector database, a broker, or Kubernetes by default.

## 4. Architecture

```text
Next.js Client ── HTTP ──────┐
      │                      ▼
      └── WebSocket ── Backend Application
                            │
                 Auth / Permissions / Sessions
                            │
                     Application Services
                            │
                       Game Engine
              Movement / Collision / Combat
                Dice / Turns / Interaction
                            │
                     Persistence Layer
                            │
                        PostgreSQL

Map and character files → Object Storage
```

HTTP is for account, campaign, character, map and asset management. WebSockets are for active session commands and events. The game engine owns rules. The renderer only presents state.

## 5. Repository Structure

```text
astralis/
├── apps/
│   ├── web/
│   │   ├── app/
│   │   ├── components/
│   │   ├── features/
│   │   ├── game-client/
│   │   │   ├── renderer/
│   │   │   ├── camera/
│   │   │   ├── input/
│   │   │   └── prediction/
│   │   └── tests/
│   └── api/
│       ├── app/
│       │   ├── api/
│       │   ├── websocket/
│       │   ├── middleware/
│       │   └── services/
│       └── tests/
├── packages/
│   ├── game_engine/
│   │   ├── movement/
│   │   ├── collision/
│   │   ├── combat/
│   │   ├── dice/
│   │   ├── turns/
│   │   ├── interaction/
│   │   ├── state/
│   │   └── events/
│   ├── contracts/
│   └── content/
│       ├── map_import/
│       └── asset_validation/
├── migrations/
├── tests/{integration,contract,e2e,load}/
├── docker-compose.yml
├── .env.example
└── README.md
```

## 6. Core Data Model

Use UUIDs for public identifiers, UTC timestamps, migrations, foreign keys and indexes.

### User
`id`, `email`, `password_hash` or identity reference, `display_name`, `status`, timestamps.

### Campaign
`id`, `owner_user_id`, `name`, `description`, `visibility`, `ruleset_id`, timestamps.

### CampaignMember
`campaign_id`, `user_id`, `role` (owner, GM, player, spectator), `joined_at`.

### Character
`id`, `owner_user_id`, optional `campaign_id`, `name`, `class_name`, `level`, attributes/resources/equipment/status data, `version`, timestamps.

### Map
`id`, `owner_user_id`, optional `campaign_id`, `name`, source asset ID, dimensions, grid configuration, spatial configuration, `version`, timestamps.

### MapLayer / MapObject
Represent visual/gameplay layers, collision, zones, spawn points, interactables, transforms and asset references.

### GameSession
`id`, `campaign_id`, `map_id`, status, owner, `current_sequence`, `state_version`, timestamps.

### SessionParticipant
`session_id`, `user_id`, `character_id`, connection status, last acknowledged sequence.

### GameState
Current session snapshot and `state_version`. Use structured domain tables for durable entities; do not put the entire application into one JSON field.

### GameEvent
`id`, `session_id`, monotonic `sequence_number`, `event_type`, actor, payload, timestamp. Add a unique constraint on `(session_id, sequence_number)`.

### Asset
`id`, `owner_user_id`, `asset_type`, storage key, content type, size, checksum, metadata, processing status, timestamp.

### Invitation
Campaign ID, creator, hashed token, expiry, usage limit, revocation state.

Use JSONB for flexible configuration, not as a replacement for relational structure. Store binary files in object storage, not PostgreSQL rows.

## 7. Authentication and Permissions

Implement registration, sign-in/out, secure session/token handling, profile editing, password recovery, campaign membership and invitations.

Rules:
- Only campaign owners/GMs can change campaign configuration.
- Only authorized users can join private campaigns.
- Players can control only assigned characters.
- Spectators cannot issue gameplay commands.
- Asset upload/edit/delete requires ownership or explicit permission.
- Every HTTP endpoint and WebSocket command checks authorization server-side.
- Never trust user IDs, roles, character ownership or permissions asserted only by the client.
- Rate-limit authentication, password recovery, invitation redemption and uploads.
- Prefer secure HttpOnly cookies where appropriate; protect cookie-authenticated mutations against CSRF.

## 8. HTTP API

Use `/api/v1`.

```text
POST/PATCH/GET  /api/v1/auth/...
GET/POST        /api/v1/campaigns
GET/PATCH/DELETE /api/v1/campaigns/{id}
GET/POST        /api/v1/campaigns/{id}/members
POST            /api/v1/campaigns/{id}/invitations
POST            /api/v1/invitations/{token}/redeem

GET/POST        /api/v1/characters
GET/PATCH/DELETE /api/v1/characters/{id}
POST            /api/v1/campaigns/{id}/characters/{character_id}/assign

GET/POST        /api/v1/maps
GET/PATCH/DELETE /api/v1/maps/{id}
POST            /api/v1/maps/{id}/validate
POST            /api/v1/maps/{id}/duplicate

POST            /api/v1/assets/upload-intent
POST            /api/v1/assets/{id}/complete
GET/DELETE      /api/v1/assets/{id}

POST            /api/v1/campaigns/{id}/sessions
GET             /api/v1/sessions/{id}
POST            /api/v1/sessions/{id}/join
POST            /api/v1/sessions/{id}/leave
GET             /api/v1/sessions/{id}/snapshot
GET             /api/v1/sessions/{id}/events?after={sequence}
```

Use signed direct-to-object-storage uploads for larger files. Verify ownership, content, size and checksum before marking an asset ready. Active gameplay actions should use the session command pipeline, not bypass it with ordinary CRUD endpoints.

## 9. WebSocket Protocol

All messages use a versioned envelope.

### Client command

```json
{
  "protocol_version": 1,
  "message_type": "command",
  "command_id": "uuid",
  "session_id": "uuid",
  "expected_state_version": 42,
  "payload": {
    "type": "character.move",
    "character_id": "uuid",
    "target": { "x": 12, "y": 18 }
  }
}
```

### Acknowledgement

```json
{
  "protocol_version": 1,
  "message_type": "command.ack",
  "command_id": "uuid",
  "status": "accepted",
  "sequence_number": 104,
  "state_version": 43
}
```

### Authoritative event

```json
{
  "protocol_version": 1,
  "message_type": "event",
  "session_id": "uuid",
  "sequence_number": 104,
  "state_version": 43,
  "event_type": "character.moved",
  "payload": {
    "character_id": "uuid",
    "from": { "x": 10, "y": 18 },
    "to": { "x": 12, "y": 18 }
  }
}
```

Required behavior:
- Authenticate the connection and authorize session membership.
- Validate schemas and message size.
- Deduplicate by `command_id`.
- Assign monotonic per-session sequence numbers.
- Check state versions and return structured rejection reasons.
- Persist accepted state changes and events atomically.
- Reconnect from the last acknowledged sequence; replay missing events or send a snapshot.
- Heartbeat connections and expire stale presence.
- Apply per-user/session rate limits.
- Never trust client timestamps for gameplay outcomes.

## 10. Game Engine Modules

### Command dispatcher
Validate schema, actor, session, permissions and state version; invoke the engine; produce typed events; persist the result; acknowledge or reject. Do not put every rule inside the dispatcher.

### Movement
- Coordinate and grid conversion
- Square-grid movement for MVP
- Movement budget and movement mode
- Path, terrain, collision and occupancy checks
- Server-derived final position
- Explicit GM-authorized teleport command
- Reject illegal or out-of-range movement

Keep game coordinates independent of isometric screen projection.

### Collision and spatial queries
Start with blocked cells/polygons, static obstacles, occupancy, basic line-of-sight blockers, interaction range and path validation. Do not start with a fully general physics sandbox.

### Dice
- Resolve rolls on the server.
- Record roll result and context in the event.
- Never accept a client-supplied result.
- Inject a seeded/test RNG for deterministic tests; never expose production RNG control to clients.

### Turns and combat
MVP turn system: encounter start/end, initiative order, active participant, movement/action budget and end-turn validation.

Combat flow:
1. Validate session, actor and turn.
2. Validate action/resource availability.
3. Validate target, range and line of sight.
4. Resolve attack and dice on the server.
5. Apply damage/status/resource changes.
6. Persist state and event atomically.
7. Broadcast the result.
8. Optionally generate narration after the result commits.

### Interaction
Support inspect, talk to a basic NPC, open a simple container and activate a simple trigger. Validate range, permissions, visibility and object state.

### Inventory
Start with item definitions, character inventory, equip/unequip and one basic consumable. Validate every mutation on the server.

### Visibility
Implement basic line of sight and optional fog-of-war state. Filter hidden entities and GM-only data before sending state to clients; never rely on hiding secret data in the UI.

### Events and snapshots
Example event types:

```text
session.started
participant.joined
participant.left
character.spawned
character.moved
object.interacted
encounter.started
turn.advanced
attack.resolved
character.damaged
character.defeated
inventory.item_equipped
quest.updated
session.paused
session.ended
```

Use PostgreSQL events for reconnect, history and audit, plus snapshots for efficient recovery. Full event sourcing of every domain is not required for MVP.

## 11. Map and Asset Import

### MVP map formats
- PNG
- JPEG
- WebP

Treat the image as a visual layer. Store grid, collision, zones, spawn points and objects separately. Add structured formats such as Tiled JSON after image-map import is stable.

### Map setup
Allow creators to configure name, dimensions, grid type (square/hex/none), grid dimensions/calibration, spawn points, blocked cells/polygons and basic zones.

### Asset security
- Enforce size, dimensions and pixel-count limits.
- Verify file signatures, not only extensions.
- Use random storage keys and private storage by default.
- Decode/process files in a controlled worker/process.
- Reject malformed and decompression-bomb images.
- Generate thumbnails and optimized variants.
- Strip unnecessary metadata when appropriate.
- Do not execute uploaded scripts.
- Check ownership for every operation.
- Track `pending`, `processing`, `ready` and `failed` states.

Use a versioned, declarative content manifest. Do not allow arbitrary uploaded code to execute on game servers.

## 12. Client Responsibilities and Reconciliation

Client owns camera, zoom, hover, selection presentation, panel state, animation playback, cosmetic effects and temporary movement previews.

Server owns actual positions, HP/resources, dice/combat outcomes, inventory, quests, turns, permissions, hidden state and world mutations.

For MVP movement:
1. Show a pending movement preview.
2. Submit the command.
3. Wait for acknowledgement.
4. Animate toward the authoritative result.
5. On rejection, restore authoritative position and show a clear reason.

Interpolate other players between authoritative updates. Never let prediction change durable state.

## 13. Networking and Reconnect

Track RTT, jitter where measurable, disconnects, command acknowledgement delay and event gaps. Show clear states: `CONNECTED`, `DEGRADED`, `RECONNECTING`, `DISCONNECTED`.

On reconnect:
1. Re-authenticate.
2. Re-authorize session membership.
3. Submit last acknowledged sequence.
4. Replay retained events or send a full snapshot.
5. Resume command processing.

High latency alone is not evidence of cheating. Never grant extra authority to high-latency clients.

## 14. AI Integration Boundary

AI is not required for the first playable loop. Later, use a provider-independent adapter for dialogue, narration, campaign summaries or intent interpretation.

Requirements:
- Validate outputs against schemas.
- AI cannot write directly to game tables.
- AI cannot decide damage, dice, legal movement or permissions.
- AI sees only information the user is allowed to know.
- AI failure cannot block movement or combat.
- Do not add embeddings/vector search or complex agent orchestration until campaign-memory needs are proven.

## 15. Errors, Logging and Observability

Use structured logs with request ID, session/campaign ID, command ID, sequence, error category and duration. Never log passwords, tokens or signed URLs.

Stable error codes should include:

```text
AUTH_REQUIRED
PERMISSION_DENIED
SESSION_NOT_ACTIVE
STALE_STATE
INVALID_COMMAND
OUT_OF_RANGE
BLOCKED_PATH
NOT_YOUR_TURN
INSUFFICIENT_RESOURCE
ASSET_INVALID
RATE_LIMITED
```

Add health/readiness endpoints and track API errors, WebSocket disconnects, command latency, database latency, failed asset processing and rejected commands.

## 16. Testing Plan

### Unit tests
Movement, grid conversion, collision, range, line of sight, dice, initiative, turns, attack legality, damage, inventory, permissions and manifest validation.

### Integration tests
Authentication, campaign membership, invitations, character assignment, asset upload, session lifecycle, WebSocket commands, transactional state/event writes, replay, snapshots and concurrent commands.

### Security tests
Attempt to access another user's private campaign, control an unassigned character, forge position/damage, replay a command, send oversized messages, upload invalid files and read hidden GM state.

### Required end-to-end scenario

```text
User A creates campaign and character
       ↓
User A uploads/configures a map
       ↓
User B joins by invitation
       ↓
Both join the session
       ↓
User A moves a character
       ↓
User B sees authoritative movement
       ↓
Encounter starts and an attack resolves
       ↓
Both clients see the same result
       ↓
User B disconnects and reconnects
       ↓
User B receives the correct current state
```

This is the release gate for the first multiplayer milestone.

## 17. Implementation Phases

### Phase 0 — Repository audit
- Inspect current repository, documents, dependencies and deployment.
- Record what already works.
- Confirm framework versions and contracts.
- Add `.env.example`, migrations, formatting/linting and tests.

**Done when:** baseline is documented and the app/API run locally without replacing working features unnecessarily.

### Phase 1 — Backend foundation
- FastAPI setup and configuration
- PostgreSQL and migrations
- Error format and structured logs
- Health/readiness
- Authentication and profile
- Authorization helpers

**Done when:** migrations work from an empty database, users can authenticate, and unauthorized access is rejected.

### Phase 2 — Campaigns and characters
- Campaign CRUD
- Membership and roles
- Invitations
- Character CRUD and assignment
- Server-side stats/resources validation

**Done when:** an owner invites a player and only authorized users can access/operate assigned characters.

### Phase 3 — Maps and assets
- Storage abstraction
- Upload lifecycle and validation
- Asset metadata/status
- Map CRUD, grid, layers and spawn points
- Collision annotations and map validation

**Done when:** a user uploads/configures/saves a map and unauthorized users cannot alter it.

### Phase 4 — Pure game engine
- Grid/coordinate utilities
- Movement and collision
- Dice
- Turns and initiative
- Basic combat
- Interaction and inventory
- Typed events and state versioning

**Done when:** engine tests run without API/database and invalid actions never mutate state.

### Phase 5 — Authoritative session server
- Session lifecycle
- WebSocket auth/presence
- Command dispatcher
- Deduplication and version checks
- Sequence numbers
- Transactional state/event writes
- Snapshot and reconnect
- Rate limits

**Done when:** two clients share a session, duplicate commands apply once, illegal/stale commands are rejected and reconnect restores state.

### Phase 6 — First playable loop
- WebSocket client manager
- Snapshot loading
- PixiJS map and tokens
- Pan/zoom and selection
- Movement preview/reconciliation
- Interaction
- Initiative and basic attack
- HP feedback, chat/system messages and reconnect UI

**Done when:** two real browser clients play the same encounter and converge on the same server state.

### Phase 7 — Hardening and OJT demo
- E2E/security/network tests
- Logs and basic metrics
- Error/loading/empty states
- Demo campaign and reset script
- Backup/restore procedure
- Setup and API documentation

**Done when:** the required E2E scenario passes and a new developer can run the project from the README.

## 18. Explicitly Defer

Do not build these before the core loop works:
- Microservices, Kubernetes and distributed simulation
- Arbitrary creator-uploaded code or physics scripts
- Full rigid-body physics sandbox
- Full 3D renderer
- AI map segmentation
- AI-controlled campaign mutation
- Advanced autonomous NPC planning
- Vector database/semantic memory
- Marketplace payments
- Cross-campaign shared world simulation
- ML-based anti-cheat
- Matchmaking and voice/video chat

Define clean interfaces where useful, but do not implement speculative infrastructure.

## 19. Definition of Done

- [ ] Authentication and authorization work.
- [ ] Campaigns, roles and invitations work.
- [ ] Characters can be created and assigned.
- [ ] Users can upload, configure and save map images.
- [ ] Multiple clients can join a session.
- [ ] Server validates every gameplay command.
- [ ] Movement is authoritative and collision-aware.
- [ ] Basic turn-based combat works.
- [ ] Dice and outcomes are server-generated.
- [ ] State and events persist reliably.
- [ ] Reconnect and recovery work.
- [ ] Two clients see consistent game state.
- [ ] Core play works without AI services.
- [ ] Unit, integration, security and E2E tests pass.
- [ ] Setup and demo instructions are documented.

## 20. Implementation Order

```text
Inspect repository
       ↓
Backend + database foundation
       ↓
Authentication and permissions
       ↓
Campaigns and characters
       ↓
Map and asset import
       ↓
Pure deterministic game engine
       ↓
Authoritative session server
       ↓
WebSocket protocol and reconnect
       ↓
PixiJS integration
       ↓
Movement and interaction
       ↓
Turn-based combat
       ↓
Persistence and end-to-end testing
       ↓
Polished playable OJT demo
```

**Engineering north star:** The client makes the game feel responsive. The server makes the game trustworthy. The event history makes the world persistent.

# Astralis --- Open Design Specification

## 1\. Product Direction

Astralis is a browser-based persistent online tabletop RPG designed as a
**game first, VTT second**.

Visual direction: - Isometric 2.5D RPG presentation - Modern minimal
application UI - Warm fantasy environments - High-contrast contextual
actions - Lightweight glass/layered panels - Responsive multiplayer-game
interactions - Persistent campaign/world information

**Core principle:** The world is the interface.

The map should dominate the screen. UI should appear when useful, remain
visually quiet when not needed, and expose complexity progressively.

\---

## 2\. Visual Reference

Use the supplied color-guide reference and generated UI concept as the
primary visual direction:

* White/off-white application chrome
* Bright yellow primary interaction color
* Cyan secondary/system color
* Coral/red danger/action color
* Very light grey surfaces
* Dark charcoal typography
* Rounded cards
* Soft shadows
* Minimal borders
* Large visual map area
* Circular character portraits/tokens
* Contextual action buttons

The reference is a design direction, not a requirement to reproduce
every component exactly.

\---

## 3\. Color System

The palette is inspired by physical color-guide swatches: clean,
saturated accent colors against a neutral paper-like interface.

### Primary

`#FFD21A` --- Astralis Yellow

Use for primary CTA, selected navigation, active tools, player
selection, important highlights and progress indicators.

### Secondary

`#18C6D9` --- Astralis Cyan

Use for magic, information, system feedback, movement previews,
interactive world elements and network/status indicators.

### Danger

`#FF6B72` --- Astralis Coral

Use for attack, damage, enemy selection, destructive actions and
critical warnings.

### Neutral

``` text
Background       #F7F7F5
Surface          #FFFFFF
Surface Muted    #F0F1EF
Border           #E5E6E3
Text Primary     #171717
Text Secondary   #6B6D69
Text Muted       #969894
```

Do not turn the entire interface into a rainbow UI. Accent colors must
communicate meaning consistently.

\---

## 4\. Typography

Primary UI font: - Inter - Geist - Manrope

Hierarchy:

``` text
Display      32–48px
Page Title   24–32px
Section      16–20px
Body         14–16px
Secondary    12–14px
Metadata     11–12px
```

Fantasy/display typography should be used sparingly for campaign names,
major locations, chapter titles and cinematic moments.

\---

## 5\. Application Shell

``` text
┌──────────────────────────────────────────────────────────┐
│ GLOBAL NAVIGATION                                        │
├───────────────┬──────────────────────────────┬───────────┤
│ CAMPAIGN      │                              │ PLAYER    │
│ SIDEBAR       │          GAME WORLD          │ INSPECTOR │
│               │            / MAP              │           │
│               │                              │           │
├───────────────┴──────────────────────────────┴───────────┤
│ CHAT / SESSION                         ACTION BAR        │
└──────────────────────────────────────────────────────────┘
```

The world/map should receive roughly 70--80% of the visual attention
during gameplay.

\---

## 6\. Global Navigation

``` text
Astralis

Play
Campaigns
Characters
Library
Community

Search
Notifications
Profile
```

Keep global navigation quiet. The user should immediately understand
where they are, what campaign they are in and what they are currently
doing.

\---

## 7\. Campaign Sidebar

``` text
THE ASHEN VALE
Campaign

Map
Characters
Quests
Journal
Inventory
Settings

SESSION
● 4 Online

Mayank       Ranger · Lv.5
Aria         Mage · Lv.5
Kael         Fighter · Lv.5
Lira         Cleric · Lv.5
```

Support collapse-to-icon mode.

\---

## 8\. Game World

The central map is the primary experience.

Support: - Isometric rendering - 2D rendering - Zoom and pan - Camera
focus - Grid toggle - Fog of war - Movement previews - Path previews -
Area-of-effect previews - Character tokens - NPCs - Environmental
objects - Spell effects - Projectiles - Lighting - Shadows

The renderer must remain independent from game rules.

Recommended implementation:

``` text
Next.js
    ↓
Game Client
    ↓
PixiJS / WebGL
    ↓
Authoritative Game Server
```

\---

## 9\. Contextual Interaction

Avoid exposing every possible action simultaneously.

### Character

Move, Attack, Interact, Use Item, Dash, Hide, Cast

### NPC

Talk, Inspect, Trade, Follow, Threaten

### Object

Open, Inspect, Take, Move, Destroy

### Enemy

Attack, Inspect, Target, Ability

The interface should progressively reveal complexity.

\---

## 10\. Player Inspector

``` text
┌─────────────────────────────┐
│ Portrait                    │
│                             │
│ Mayank                      │
│ Human Ranger · Lv.5         │
│                             │
│ HP    █████████░ 82/100     │
│ Mana  ██████░░░ 45/70       │
│                             │
│ STR 14  DEX 18  CON 13      │
│ INT 12  WIS 16  CHA 10     │
│                             │
│ Actions Inventory Spells    │
│                             │
│ \[ Move ] \[ Attack ]         │
│ \[ Interact ]                │
└─────────────────────────────┘
```

Actions should be large enough to understand without reading
documentation.

\---

## 11\. Quest Panel

``` text
CURRENT QUEST

The Missing Caravan

✓ Reach the old bridge
□ Investigate the river camp
□ Find the caravan's ledger
```

Keep quest information compact and immediately readable.

\---

## 12\. Chat and Session Layer

Bottom-left panel:

``` text
Chat
Combat Log
System

Aria
Should we check the bridge first?

Kael
I'll move ahead and look for traps.

Lira
I'll keep an eye on the forest side.

\[ Type a message... ]
```

Chat should not obscure the game world and can collapse automatically.

\---

## 13\. Combat Mode

Combat should transform the interface without navigating to a separate
page.

``` text
EXPLORATION
      ↓
TACTICAL MODE
```

Add: - Grid - Movement range - Attack range - Line of sight -
Initiative - Target indicators - Action economy - Combat log

The visual world remains the same.

\---

## 14\. Initiative

``` text
INITIATIVE

1  Mayank          20
2  Bandit Scout    17
3  Lira            14
4  Aria            12
5  Kael            10
6  Bandit Archer    8
```

Semantic colors: - Yellow = active player - Coral = hostile - Cyan =
system/ability state

\---

# 15\. User-Uploaded Map System

Users must be able to use their own maps.

### MVP formats

``` text
PNG
JPG / JPEG
WebP
```

The uploaded image is a **visual layer**, not the authoritative map.

``` text
Uploaded Image
      +
Grid
      +
Collision
      +
Zones
      +
Spawn Points
      +
Elevation
      +
Entities
      ↓
Structured Game Map
```

### Map Import Flow

``` text
Upload Map
     ↓
Preview
     ↓
Map Configuration
     ↓
Grid Calibration
     ↓
Gameplay Annotation
     ↓
Save Map
     ↓
Play
```

\---

## 16\. Map Calibration

Allow:

``` text
Grid Type

○ Square
○ Hex
○ None
```

For square/hex maps: - Detect grid automatically where possible - Allow
manual correction - Allow user-defined grid dimensions - Allow
pixel-to-cell calibration

Example:

``` text
Grid spacing: 64px
Columns: 40
Rows: 30
```

\---

## 17\. Map Annotation

Map editor tools:

``` text
Select
Wall
Collision
Zone
Spawn
Enemy Spawn
Point
Elevation
Hazard
Water
```

Annotations create gameplay metadata without modifying the original
image.

\---

## 18\. AI-Assisted Map Detection

Future feature:

``` text
Analyze Map

Detected:

Grid              98%
Walls             81%
Doors             74%
Rooms             89%
Stairs            63%
```

Always provide:

``` text
Apply
Review
Cancel
```

AI detection is advisory. The deterministic game engine remains
authoritative.

\---

# 19\. Loading Experience

## Design Direction

The loading screen should have the restrained futuristic, cinematic
feeling of a premium narrative-game boot sequence, inspired by the
*Detroit: Become Human* loading-screen atmosphere without copying its
exact interface, characters, animations, branding or assets.

The experience should feel: - Minimal - White/off-white - Clinical -
Premium - Quiet - Slightly futuristic - Human-centered - Cinematic

This deliberately contrasts with the colorful fantasy gameplay UI.

\---

## 20\. Astralis Loading Screen

Concept:

``` text
                    ASTRALIS

                      ◇

                 INITIALIZING
                  WORLD STATE

               ● ● ● ● ○ ○ ○

              Connecting players...

              WORLD SYNCHRONIZATION
```

Use generous negative space.

A small geometric Astralis mark can sit near the center.

\---

## 21\. Loading Animation

Suggested sequence:

``` text
0.0s
Logo appears

0.5s
Thin geometric lines assemble

1.0s
Small system indicator appears

1.5s
Status text changes

2.0s
World synchronization begins

2.5s
Campaign title fades in

3.0s
Map begins transitioning into view
```

Status sequence:

``` text
INITIALIZING WORLD
CONNECTING SESSION
LOADING CAMPAIGN
SYNCING CHARACTERS
RESTORING WORLD STATE
READY
```

Do not use a generic spinning loader as the primary animation.

\---

## 22\. Loading Screen Color

Keep the loading screen mostly:

``` text
#F7F7F5
#171717
```

Use Astralis Yellow only for progress indicators, active status and
small geometric accents.

Cyan may appear briefly during synchronization.

Coral should generally not appear during normal loading.

\---

## 23\. Loading Screen Transition

The loading screen should transition into gameplay rather than simply
disappear.

``` text
LOADING SCREEN
      ↓
Campaign title
      ↓
Location title
      ↓
Camera moves into map
      ↓
HUD fades in
      ↓
Player control enabled
```

Example:

``` text
THE ASHEN VALE

RIVERSIDE VILLAGE

        ↓

Camera enters the world

        ↓

HUD appears

        ↓

CONTROL READY
```

The goal is to make the user feel that they are entering a world, not
opening a webpage.

\---

## 24\. Responsive Design

Desktop is the primary target.

Minimum recommended gameplay target:

``` text
1280 × 720
```

Preferred:

``` text
1440 × 900
1920 × 1080
```

Tablet: - Collapsible sidebar - Bottom action bar - Simplified inspector

Mobile MVP: - Campaign management - Character management - Chat -
Journal - Inventory - Session joining

Full tactical gameplay can remain desktop-first.

\---

## 25\. Design Principles

### World First

The map is the primary interface.

### Context Over Clutter

Show actions when relevant.

### Progressive Disclosure

Expose advanced mechanics when the user needs them.

### Semantic Color

Yellow, cyan and coral communicate different meanings consistently.

### Readability Over Decoration

Fantasy styling must never make information difficult to understand.

### Responsive Interaction

Clicks, selections, movement previews and actions should feel immediate.

### Cinematic Transitions

Important state changes should feel like part of the game.

### Consistency

Every interaction follows the same visual language.

\---

## 26\. Component Library

``` text
Button
IconButton
Card
Panel
Modal
Tooltip
Popover
Tabs
ProgressBar
HealthBar
ManaBar
Avatar
CharacterToken
QuestItem
ActionButton
InventorySlot
AbilityCard
ChatMessage
InitiativeItem
MapToolbar
MapMarker
LoadingScreen
Toast
Notification
ContextMenu
```

Each component should support applicable: - Default - Hover - Active -
Disabled - Loading - Selected - Danger - Success

states.

\---

## 27\. UX Goal

A new player should understand the basic gameplay loop without
documentation:

``` text
See world
   ↓
Select character
   ↓
See available actions
   ↓
Choose action
   ↓
See immediate result
   ↓
Continue exploring
```

The interface should teach the game through interaction.

\---

## 28\. Design Anti-Patterns

Avoid: - Dense VTT-style toolbars - Too many floating panels - Permanent
grids - Excessive fantasy ornamentation - Excessive glassmorphism - Neon
AI aesthetics - Generic SaaS dashboard styling - Tiny buttons - Long
text blocks during gameplay - 20+ visible actions - UI covering the
map - Technical implementation terminology in player-facing UI -
Decorative numbered onboarding labels such as
`numbered onboarding labels` - Generic AI-themed placeholder copy or
fake product sections

The player should think:

> "I'm playing an RPG."

Not:

> "I'm operating a game management application."

\---

## 29\. MVP Design Scope

Build these first:

1. Loading screen
2. Campaign shell
3. Isometric map
4. Character tokens
5. Left campaign navigation
6. Player inspector
7. Contextual action buttons
8. Quest panel
9. Chat
10. Initiative panel
11. Map upload
12. Grid calibration
13. Basic map annotation
14. Exploration mode
15. Tactical mode
16. Responsive camera
17. Zoom/pan
18. Basic transitions

\---

## 30\. North-Star Visual Statement

**Astralis should look like a premium fantasy game that happens to run
in the browser.**

The UI should disappear when the world becomes interesting.

The world should remain readable when the UI becomes complex.

The loading experience should establish the opposite side of the
product: calm, precise, cinematic and futuristic before transitioning
into the colorful, living fantasy world.

# 31\. Recommended Design Order

Design the product in this order so the visual language is established
from the core game outward.

### Foundation

1. Loading / World Transition
2. Gameplay / World View
3. Tactical Combat State
4. Map Editor / Map Setup

### Player Experience

5. Character Sheet
6. Character Selection
7. Inventory / Equipment
8. Quest Log
9. Journal / Campaign Chronicle

### Multiplayer / Campaign

10. Campaign Lobby
11. Home / Campaign Dashboard
12. Campaign Browser

### Supporting Product

13. Landing Page
14. Authentication
15. Profile
16. Map Library
17. Community / Shared Worlds
18. Settings

The **gameplay screen should be designed first**, because its map,
token, action, panel, typography and interaction components establish
the language used throughout the application.



# 31\. Design Style System

Astralis should use a **controlled combination of design styles**, not one style everywhere.

The product should feel like:

**Premium game UI + editorial minimalism + soft glass layering + cinematic fantasy**

Do not use a generic "fantasy UI" treatment across every surface.

## 31.1 Primary Style — Premium Minimal

Use this as the foundation for the application shell.

Characteristics:

* Clean whitespace
* Strong hierarchy
* Large visual areas
* Simple iconography
* Rounded but not overly soft components
* Thin borders
* Restrained shadows
* High-quality typography
* Minimal decorative elements

Use for:

* Navigation
* Campaign dashboard
* Character management
* Settings
* Library
* Community
* Authentication

Reference direction:

```text
Premium
Minimal
Editorial
Modern
Quiet
```

\---

## 31.2 Secondary Style — Glass \& Layered

Use for gameplay overlays rather than the entire application.

Characteristics:

* Semi-transparent white surfaces
* Slight background blur
* Soft shadows
* 12–20px corner radius
* Thin translucent borders
* Layered depth
* High contrast text

Use for:

* Player inspector
* Quest panel
* Chat
* Initiative
* Context menus
* Map controls
* Item inspection
* Dialogue panels

Important:

**Do not make the entire game world glassmorphic.**

The map remains solid and visually rich; UI floats above it.

\---

## 31.3 Game World Style — Stylized Isometric Fantasy

The world should use a cohesive stylized 2.5D art direction.

Characteristics:

* Isometric perspective
* Hand-painted or painterly environments
* Slightly exaggerated shapes
* Rich environmental detail
* Soft atmospheric lighting
* Readable silhouettes
* Clear character separation
* Controlled saturation

Avoid:

* Photorealistic textures
* Extremely dark environments
* Hyper-realistic AAA rendering
* Random asset-pack styles
* Excessive outlines

The world should feel like a premium stylized RPG rather than a tabletop board.

\---

## 31.4 Tactical Style — Clean Game Readability

Combat should temporarily increase visual clarity.

Use:

* Crisp grid
* Clear movement cells
* High-contrast target indicators
* Thin range boundaries
* Strong active-player indicator
* Clear AoE shapes
* Simple combat feedback
* Minimal HUD clutter

The tactical layer should sit on top of the world rather than replace it.

\---

## 31.5 Cinematic Style — Futuristic Minimal

Use for:

* Loading
* Major story moments
* Chapter transitions
* Major discoveries
* Campaign start
* World-state transitions

Characteristics:

* Large negative space
* Thin geometric elements
* Subtle motion
* Large typography
* Minimal information
* Off-white/charcoal palette
* Small controlled accent color

The atmosphere can be inspired by premium cinematic game interfaces such as *Detroit: Become Human*, but should remain an original Astralis visual language.

\---

# 32\. Color Palette System

Astralis should use **four functional palettes**.

## 32.1 Core UI Palette

```text
Paper             #F7F7F5
White             #FFFFFF
Soft Grey         #F0F1EF
Border            #E4E5E2

Ink               #171717
Charcoal          #292A28
Secondary Text    #686A66
Muted Text        #969894
```

These colors should represent approximately **75–85% of the normal application UI**.

The interface should therefore feel clean even when accent colors are active.

\---

## 32.2 Astralis Accent Palette

Inspired by the supplied color-guide reference.

```text
ASTRALIS YELLOW
#FFD21A

ASTRALIS CYAN
#18C6D9

ASTRALIS CORAL
#FF6B72
```

### Yellow

Meaning:

```text
Primary action
Selection
Active player
Important progress
Confirmation
Interactive focus
```

### Cyan

Meaning:

```text
Information
Magic
Movement preview
System state
Interactive world object
Connection status
```

### Coral

Meaning:

```text
Combat
Enemy
Damage
Danger
Destructive action
Critical warning
```

Never use these three colors randomly.

\---

# 33\. Extended Semantic Palette

The core palette can be extended when gameplay requires additional states.

```text
Success Green       #48B88A
Warning Amber       #F2A93B
Error Red           #E55353
Purple              #8B72D9
Neutral Blue        #6686A8
```

These are secondary semantic colors and should appear much less frequently than Yellow, Cyan and Coral.

Example:

```text
Success Green → quest completed
Warning Amber → uncertain interaction
Error Red     → invalid action
Purple        → special/rare ability
```

Do not introduce a new color when an existing semantic color already communicates the state.

\---

# 34\. Game-State Color Mapping

The entire product should follow one color language.

|State|Color|
|-|-|
|Primary interaction|Yellow|
|Active player|Yellow|
|Selected entity|Yellow|
|Movement|Cyan|
|Magic|Cyan|
|Information|Cyan|
|Enemy|Coral|
|Damage|Coral|
|Danger|Coral|
|Completed|Green|
|Warning|Amber|
|Error|Red|
|Rare/special|Purple|

This mapping must remain consistent across:

* Map
* Character sheet
* Inventory
* Quests
* Combat
* Notifications
* Chat
* Loading states

\---

# 35\. Map Color Treatment

The uploaded map or environment art should **not** be recolored to match the UI palette.

Instead:

```text
WORLD ART
Natural / fantasy colors

       +

UI
Yellow / Cyan / Coral

       +

GAMEPLAY OVERLAY
Semantic highlights
```

This distinction is important.

The world should look like a world.

The UI should look like Astralis.

\---

# 36\. UI Surface System

Use five surface levels.

```text
Surface 0
Page background
#F7F7F5

Surface 1
Cards / panels
#FFFFFF

Surface 2
Elevated panels
rgba(255,255,255,0.88)

Surface 3
Gameplay overlay
rgba(255,255,255,0.76)

Surface 4
Focused modal
rgba(255,255,255,0.96)
```

Use blur only on surfaces that sit above the game world.

\---

# 37\. Border and Shadow Style

Borders should be subtle.

Default:

```text
1px solid #E4E5E2
```

Gameplay overlays can use:

```text
1px solid rgba(255,255,255,0.65)
```

Shadows should communicate elevation rather than decoration.

Preferred character:

```text
soft
wide
low opacity
```

Avoid:

* heavy black shadows
* glowing cards
* neon outlines
* excessive drop shadows

\---

# 38\. Shape Language

Use a consistent rounded geometry.

Recommended:

```text
Small controls       10px
Buttons              12px
Cards                16px
Major panels         20px
Modal                24px
Character portraits  50% / circular
```

Avoid mixing:

* extremely sharp rectangles
* excessive pill buttons
* ornate fantasy frames
* inconsistent corner radii

The product should feel modern first.

\---

# 39\. Icon Style

Use a single primary icon family.

Preferred:

**Simple outline icons with moderate stroke weight.**

Use icons for:

* Navigation
* Actions
* Equipment
* Combat
* Map tools
* Status

Avoid mixing:

* 3D icons
* emoji
* pixel icons
* ornate fantasy icons
* glossy icons

inside the same UI.

Fantasy detail belongs primarily in the game world and artwork.

\---

# 40\. Button Style

### Primary

```text
Yellow background
Dark text
12px radius
Medium/bold label
```

Example:

```text
\[ Continue Campaign ]
```

### Secondary

```text
White / transparent
Dark text
Thin border
```

### Informational

```text
Cyan background
Dark text
```

### Danger

```text
Coral background
Dark/white text depending on contrast
```

Avoid gradients for normal buttons.

\---

# 41\. Card Style

Cards should be:

* Simple
* Spacious
* Image-led when relevant
* Lightly elevated
* Easy to scan

Example campaign card:

```text
┌──────────────────────────────┐
│                              │
│       WORLD IMAGE            │
│                              │
├──────────────────────────────┤
│ The Ashen Vale               │
│ Riverside Village            │
│                              │
│ Last played · 2h ago         │
└──────────────────────────────┘
```

Do not add decorative badges simply to fill empty space.

\---

# 42\. Animation Style

Animation should feel **responsive and intentional**, not flashy.

Use:

```text
Fast UI response       120–180ms
Panel transition       180–250ms
Modal transition       220–300ms
World transition       400–800ms
Cinematic transition   800–1500ms
```

Preferred motion:

* Fade
* Scale 98% → 100%
* Slide 8–16px
* Camera pan
* Camera zoom
* Soft parallax
* Subtle world movement

Avoid:

* Excessive bouncing
* Constant particle effects
* Large UI rotations
* Oversized spring animations
* Animation on every hover

\---

# 43\. Gameplay Feedback Style

Actions should provide immediate visual feedback.

### Movement

```text
Selected character
        ↓
Cyan reachable area
        ↓
Path preview
        ↓
Character moves
```

### Attack

```text
Select enemy
        ↓
Coral target indicator
        ↓
Attack range
        ↓
Action
        ↓
Damage feedback
```

### Magic

```text
Select spell
        ↓
Cyan range
        ↓
AoE preview
        ↓
Confirm
        ↓
Spell effect
```

The color language should make these states understandable without text.

\---

# 44\. Loading Screen Style Palette

The loading screen should intentionally use a reduced palette.

```text
Background     #F7F7F5
Primary text   #171717
Secondary      #686A66
Accent         #FFD21A
Optional sync  #18C6D9
```

Do not use Coral during normal loading.

The reduced palette creates contrast with the full-color fantasy world that follows.

\---

# 45\. Dark Mode

Dark mode can be supported later, but should not become the primary visual direction during the first design pass.

If implemented:

```text
Background     #101111
Surface        #171918
Elevated       #202321
Text           #F5F5F2
Muted          #9B9E99
Yellow         #FFD21A
Cyan           #18C6D9
Coral          #FF6B72
```

The same semantic color mapping must remain unchanged.

\---

# 46\. Design Style Summary

The final visual system should combine:

```text
FOUNDATION
Premium Minimal
        +
EDITORIAL
Clean Typography / White Space
        +
GAME UI
Contextual / Responsive
        +
LAYERING
Soft Glass / Floating Panels
        +
WORLD
Stylized Isometric Fantasy
        +
CINEMATIC
Futuristic Minimal Transitions
```

Avoid combining unrelated visual styles simply because they look impressive individually.

The goal is a single recognizable Astralis design language.

\---

# 47\. Palette Priority

When designing a screen, use the palette in this order:

```text
1. Neutral foundation
2. Yellow primary action
3. Cyan information/system
4. Coral danger/combat
5. Secondary semantic colors only when necessary
```

A typical screen should **not** use every accent color at once.

A good rule:

> If everything is highlighted, nothing is highlighted.

\---

# 48\. Design System Deliverables

The design phase should produce:

1. Color tokens
2. Typography tokens
3. Spacing tokens
4. Radius tokens
5. Shadow tokens
6. Icon rules
7. Button variants
8. Card variants
9. Panel variants
10. Input states
11. Modal states
12. Tooltip states
13. Map interaction states
14. Combat states
15. Loading states
16. Character states
17. Quest states
18. Notification states
19. Responsive breakpoints
20. Animation/motion rules

These should be defined before implementing all 18 product screens.

\---

# 49\. Final Style Direction

Astral should not look like:

* A conventional VTT
* A fantasy board-game website
* A generic SaaS dashboard
* A neon "AI product"
* A dark cyberpunk interface
* A collection of unrelated game assets

Astral should look like:

> \*\*A clean premium game interface surrounding a living, colorful fantasy world.\*\*

The visual contrast is intentional:

**Calm interface → colorful world → precise gameplay feedback → cinematic transitions.**


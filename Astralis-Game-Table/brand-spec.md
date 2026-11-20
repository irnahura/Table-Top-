# Astralis visual system

Source: `Astralis_Open_Design_Spec_v2.md`

## Tokens

```css
--bg:      oklch(97.2% 0.004 105); /* #F7F7F5 */
--surface: oklch(100% 0 0);       /* #FFFFFF */
--fg:      oklch(16.5% 0 0);       /* #171717 */
--muted:   oklch(49.5% 0.008 125);/* #6B6D69 */
--border:  oklch(91.5% 0.006 120);/* #E4E5E2 */
--accent:  oklch(86.9% 0.18 94);  /* #FFD21A */
```

Supporting semantic colors remain exact to the source spec: Astralis Cyan `#18C6D9`, Astralis Coral `#FF6B72`, Soft Grey `#F0F1EF`, Charcoal `#292A28`, and Muted Text `#969894`.

## Type

- Display: `Inter`, `system-ui`, `-apple-system`, `Segoe UI`, `sans-serif`
- Body: `Inter`, `system-ui`, `-apple-system`, `Segoe UI`, `sans-serif`
- Metadata: `ui-monospace`, `SFMono-Regular`, `Consolas`, `monospace`

## Observed rules

1. The world receives most of the attention; UI floats quietly above it.
2. Yellow means active intent and selection; cyan means movement/system state; coral means combat and danger.
3. Gameplay surfaces are light glass with restrained borders and soft elevation.
4. The map stays colorful and solid; the interface stays neutral and editorial.
5. Tactical clarity appears contextually instead of exposing a permanent VTT toolbar.

System summary: Astralis is a premium, minimal game shell around a living isometric fantasy world, with semantic color doing the teaching.

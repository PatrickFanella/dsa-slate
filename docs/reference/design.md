# DSA Slate Design Reference

DSA Slate uses the official DSA design system palette from `https://design.dsausa.org` as the visual foundation.

The UI now supports runtime theme switching with `data-theme="light" | "dark"` on `document.documentElement`.

## Palette

| Token | HEX |
|---|---:|
| DSA Red | `#EC1F27` |
| DSA Red Tint 1 | `#F04C53` |
| DSA Red Tint 2 | `#F4797E` |
| DSA Red Tint 3 | `#F7A5A9` |
| DSA Red Tint 4 | `#FBD2D4` |
| DSA Black | `#231F20` |
| DSA Black Tint 1 | `#3B3838` |
| DSA Black Tint 2 | `#605C5C` |
| DSA Black Tint 3 | `#8C8989` |
| DSA Black Tint 4 | `#C1C0BF` |
| White | `#FFFFFF` |

## App mapping

- `primary` maps to DSA Red.
- `foreground` maps to DSA Black.
- `background` and `surface` map to White.
- `surface-muted` maps to DSA Red Tint 4.
- `border` uses DSA Black Tint 4.
- `border-strong` uses DSA Black.
- `primary-hover` maps to DSA Black for accessible primary-action hover contrast.

## Dark theme

Dark mode keeps DSA Red as the accent while shifting the runtime tokens to DSA Black and black tints.

- `background` maps to DSA Black.
- `surface` maps to DSA Black Tint 1.
- `surface-muted` maps to DSA Black Tint 2.
- `border-strong` uses DSA Red Tint 2 for contrast and focus on dark surfaces.
- `primary-foreground` maps to White to keep red buttons readable.
- `shadow` uses red tint accents in dark mode and black tint shadows in light mode.

## Tailwind v4

Theme tokens live in `web/src/styles.css` using CSS-first Tailwind v4 configuration and `@theme inline`.

Use semantic utilities such as `bg-background`, `text-foreground`, `bg-surface`, `border-border`, `bg-primary`, and `text-primary-foreground`. Avoid raw color utilities like `bg-red-950` for app UI.

## Shape and depth

- Prefer `rounded-none` for cards, panels, badges, and controls.
- If a component needs softness, keep it minimal and intentional.
- Use hard borders and offset shadows instead of soft-radius elevation.

## Accessibility notes

- Do not rely on color alone for confidence or endorsement meaning.
- Keep source links underlined.
- Keep focus rings visible.
- Avoid red-on-green combinations.
- Keep essential text at readable sizes.

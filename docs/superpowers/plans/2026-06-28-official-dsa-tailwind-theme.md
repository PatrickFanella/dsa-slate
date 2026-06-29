# Official DSA Tailwind Theme Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Replace the current dark red draft with an official DSA Tailwind v4 theme foundation, square/neobrutalist component styling, and runtime light/dark theme switching.

**Architecture:** Keep Tailwind v4 CSS-first configuration in `web/src/styles.css` using `@import "tailwindcss"`, official DSA palette variables, runtime semantic variables, and `@theme inline` mappings. React owns a small theme toggle that writes `data-theme="light" | "dark"` to `document.documentElement`; components consume only semantic Tailwind utilities.

**Tech Stack:** Vite React TypeScript, Tailwind CSS v4 via `@tailwindcss/vite`, CSS custom properties, Vitest, existing `make verify` workflow.

---

## Tasks

### Task 1: Add official DSA Tailwind v4 semantic tokens

- [x] Define official DSA raw palette variables in `web/src/styles.css`.
- [x] Map semantic color/font tokens with Tailwind v4 `@theme inline`.
- [x] Keep light theme as the default token set.

### Task 2: Convert App shell to semantic DSA utilities

- [x] Convert hero, search, stats, sections, empty states, and API status from raw dark utilities to semantic utilities.
- [x] Preserve existing copy, source-linked workflow, API loading status, and privacy warning.

### Task 3: Convert cards, methodology, and source badges

- [x] Convert directory cards, methodology panels, and source badges to semantic utilities.
- [x] Keep source links underlined and focus-visible.

### Task 4: Document the official DSA theme mapping

- [x] Create `docs/reference/design.md` with palette, semantic mapping, Tailwind v4 usage, and accessibility notes.
- [x] Link the design reference from `README.md`.

### Task 5: Add neobrutalist shape and dark theme switching

- [x] Replace rounded UI classes with `rounded-none` and hard offset shadows.
- [x] Add `data-theme="dark"` semantic token overrides derived from the light DSA palette.
- [x] Add a header theme toggle and persist selection in `localStorage` when available.
- [x] Add regression assertions for semantic classes, square style, and theme toggle rendering.

### Task 6: Final verification

- [x] Run `make verify`.
- [x] Confirm old dark theme classes are absent from `web/src`:
  - `rounded-3xl`, `rounded-2xl`, `rounded-xl`, `rounded-full`
  - `red-950`, `red-900`, `red-800`, `rose-*`, `text-white`, `bg-white`

## Acceptance criteria

- Official DSA palette is represented as raw CSS variables in `web/src/styles.css`.
- Tailwind v4 semantic utilities are available through `@theme inline`.
- Light and dark themes switch via `data-theme` on the document root.
- UI uses square/neobrutalist styling with hard borders and offset shadows.
- Existing app behavior and copy remain intact.
- `make verify` passes.

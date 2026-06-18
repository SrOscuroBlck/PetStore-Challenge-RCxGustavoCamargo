---
name: new-component
description: Scaffold a new React component or screen following this project's conventions — strict typed props, thin presentation kept separate from the data/API layer, designed loading/empty/error states for anything async, accessible markup, and prefers-reduced-motion-aware Framer Motion. Use when creating any new UI component or route.
---

# New component

Build UI that meets the polish, a11y, and architecture bars in CLAUDE.md from the first commit.

## Conventions

- **Thin & typed.** A presentation component receives typed props and renders; it does not call the
  network. Data comes from generated Apollo hooks in a container/route, passed down. No `any`.
- **Co-locate** the component with its styles/tests; name files and the default export the same as the
  component.

## Required states (for anything async)
- **Loading** → a skeleton matching the final layout (avoid layout shift), not a bare spinner.
- **Empty** → a friendly, on-brand empty state.
- **Error** → a toast/banner with the human-readable message, derived from `extensions.code`.
- **Optimistic** → immediate animated feedback for purchase/add-to-cart; roll back on `UNAVAILABLE`.

## Accessibility
- Semantic elements (`button`/`a`/`nav`/`main`/ordered headings) — never a clickable `div`.
- Keyboard operable; manage focus on mount/route/modal/cart; visible focus ring.
- `aria-label` on icon-only controls; meaningful `alt` on images (`alt=""` if decorative).
- Contrast holds in light and dark mode.

## Motion
- Use Framer Motion for enter/exit, hover, and transitions — purposeful, not decorative.
- Gate every animation on reduced motion: `const reduce = useReducedMotion()` and skip/short-circuit
  transitions when true.

## Images
- `pictureUrl` straight into `<img>`; add `loading="lazy"`, width/height, and an `onError` placeholder.

## Finish
- Run `npm run typecheck`. For a screen with real interactions, consider the `ux-state-auditor` agent
  for a polish/a11y pass.

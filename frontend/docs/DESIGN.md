# Design system

The visual language for the customer storefront. **Design DNA: a warm, friendly pet-store** —
Robotic Crew's orange brand colour carried into a soft, rounded, photo-forward storefront in the
spirit of Chewy / Petco / Pets-at-Home and playful adoption sites. We keep RC's hot orange against
a warm cream page, fused with the proven shopping flows of pet commerce, aligned to the challenge
(browse available pets → purchase → cart → checkout).

This document is the source of truth for UI. Tokens live in `src/styles/tokens.css` +
`tailwind.config.ts`; motif utilities/components in `src/styles/index.css` and
`src/components/brand`. Build to these — don't invent ad-hoc colors, type, or spacing.

## 1. Brand pillars

- **Friendly & rounded** — soft corners, pill controls, gentle shadows. Approachable, never clinical
  or "techy". (We deliberately moved away from the earlier angular/industrial treatment.)
- **One hot accent** — a single vivid orange against warm cream + occasional navy `ink` chrome. Orange
  is the brand, the primary action, and section emphasis.
- **Warm & photo-forward** — pets are the heroes. Generous whitespace, large rounded imagery, warm copy.
- **Playful, not noisy** — the paw motif and species colours add personality; the catalog stays scannable.

## 2. Color

RGB-channel CSS variables. The app ships **light-only** (see §9); a `.dark` palette is defined for
possible future use but is not auto-applied. Tailwind names in parentheses.

| Token (Tailwind) | Light | Use |
|---|---|---|
| `bg` | `#FCF7F0` (warm cream) | Page background |
| `card` | `#FFFFFF` | Surfaces, cards |
| `border` | `#EDE5DB` (warm) | Hairlines, dividers |
| `fg` | `#26211E` (warm near-black) | Body + heading text |
| `muted` | `#756B63` (warm) | Secondary text |
| `ink` / `ink-fg` | `#1C232B` / `#F6F3EE` | **Dark chrome** (header, footer, feature bands) + its text |
| `primary` | `#F26522` | **Brand orange** — fills, borders, icons, large display text |
| `primary-strong` | `#D9551A` | Hover/pressed orange fills |
| `primary-fg` | `#291608` (deep brown) | Text/icon **on** an orange fill |
| `accent` | `#C2410C` | Orange **text** at body/label sizes (AA-safe) |
| `cat` / `dog` / `frog` | amber / blue / green | Soft species accents (tiles, badges, placeholders) |
| `danger` / `success` | `#C6281C` / `#197943` | Status |

### Apply orange generously (it's the brand)

The page reads **warm and orange**, not navy/white/gray. Large headings, solid pill buttons, eyebrows,
soft underlines, and accents are orange against cream. Pet photos and the soft species tints (cat/dog/
frog) supply the rest of the colour. Reserve dark `ink` for chrome (header/footer) and feature bands.

### Accessibility colour rule (non-negotiable)

Orange at body sizes on cream is ~3.3:1 — fine for large text/graphics, **not** normal text. So:

- **`primary`** (vivid orange) → fills, borders, icons, and **large** display text (≥ ~24px), incl. headings.
- **`accent`** (`#C2410C`) → orange-coloured **text** at body/label sizes on light (AA). This is what `.eyebrow` uses.
- **`primary-fg`** → text/icons placed **on** an orange fill. Primary buttons are dark-on-orange, **not** white-on-orange.
- Species `cat`/`dog`/`frog` are used as **soft fills/tints** (`/15`) and solid badges with readable text — not as body text.
- Never put small orange text on a light background. Body text is always `fg` / `muted`.

## 3. Typography

Two self-hosted families (bundled via `@fontsource`, no CDN — offline + CSP-safe):

- **Display — "Fredoka"** (`font-display`): friendly rounded face. Headings (`h1–h3` default to it),
  eyebrows, CTA labels, the wordmark. **Not** uppercase by default — friendly title/sentence case; only
  the small `.eyebrow` label stays uppercase.
- **Body/UI — "Nunito"** (`font-sans`, default): all paragraphs, descriptions, inputs, secondary UI.
  Warm and rounded with excellent legibility for the reading-heavy catalog.

Scale (Tailwind): display `text-5xl`/`text-6xl` (hero), `text-3xl` (section), `text-xl` (card titles);
body `text-base`/`text-sm`; eyebrow `text-xs` uppercase, tracking `0.14em`. Headings keep their natural
(rounded) tracking; body line-height comfortable.

## 4. Space, shape, elevation

- **Radius** `--radius: 1rem` (generously rounded). Cards use `rounded-2xl`; **buttons, inputs, pills,
  and filter chips are `rounded-full`**.
- **No chamfers** — the old angular `clip-corner` notch is gone. Everything is soft-rounded.
- **Container** `max-w-6xl`, padding `px-4 sm:px-6`. Mobile-first; everything works from ~320px.
- **Elevation** — soft warm shadows are now welcome (friendly, not techy): `shadow-soft` for resting
  cards/toasts, `shadow-lift` for hover/active emphasis. Borders are thin + warm.

## 5. Signature motifs — and how to use them

Codified as utilities/components so usage stays consistent:

| Motif | How | Where |
|---|---|---|
| **Paw mark** | `<Paw />` (`components/brand/Paw.tsx`), `currentColor` | Logo badge, image placeholders, hero/empty-state decoration, cart header |
| **Eyebrow label** | `.eyebrow` (uppercase orange `accent`, display, tracked) | Above every section/page heading |
| **Soft rule** | `.rule-soft` (rounded orange underline bar) on a heading | Under section/page headings |
| **Paw bullets** | `.list-paw` on a `<ul>` (SVG paw-mask markers) | Feature/benefit lists |
| **Species tint/badge** | `speciesTheme()` (`features/catalog/speciesTheme.ts`) | Card photo tint, species pill, placeholder icon colour |
| **Pill button / chip** | `rounded-full` + `shadow-soft` | All buttons and the species filter |
| **Dark feature band** | `bg-ink text-ink-fg` section | Footer, occasional emphasis bands |

## 6. Components

Build thin + typed; data fetched in containers, passed to presentation (CLAUDE.md). Specs:

- **Button** (`components/ui/Button.tsx`) — pill (`rounded-full`), display label (title case): `primary`
  (orange fill, dark label, soft→lift shadow), `secondary` (bordered, orange border/tint on hover),
  `ghost`, `onDark` (outline on ink). Tactile `active:scale-95` (motion-safe).
- **Header / top bar** — `bg-ink` band: paw `Logo` left; **cart (icon + orange count badge)** right.
  Sticky. Sign-in email + logout when authed.
- **Pet card** (#2) — rounded-2xl `card`, soft shadow, hover lift. Photo (`pictureUrl`, lazy, sized,
  paw placeholder on 404) tinted by species; species **badge** (coloured pill with animal icon); name (display)
  + age pill + short description (`muted`); **Buy** (primary) + **Add to cart** (secondary). Sold = dimmed + "Sold".
- **Species filter** — `rounded-full` chips with an animal icon per species (paw / cat / dog / frog SVGs); active = orange fill.
- **Cart drawer** (#3) — Radix Dialog from the right, rounded-left panel; `ink` header with paw; line
  items with rounded thumbnail; **Checkout** primary CTA; named-unavailable-pets error surfaced verbatim.
- **Toasts** (`components/ui/Toast`) — Radix; rounded `card` with a left accent border (`primary`/
  `danger`/`success`), display title. Purchase/checkout/error feedback.
- **Inputs** (`components/ui/TextField`) — `rounded-full`, bordered, `card` bg, labelled (display),
  `aria-invalid`/`-describedby`, orange focus border.
- **States** — every async surface designs loading (skeleton), empty (friendly paw card), and error
  (rounded banner/toast with the server's human-readable message). Never a blank screen or raw error.

## 7. Motion

Framer Motion, **purposeful**: card hover lift, list enter/exit (stagger), section fade-in-up
(`animate-fade-in-up`), `pop-in` for badges, cart open/close, optimistic purchase (pet animates out).
No gratuitous motion. **Always gate on reduced motion** via `useReducedMotion`
(`src/lib/a11y/useReducedMotion.ts`) — when true, disable transforms/opacity transitions and render
the final state. Pill buttons' `active:scale` is `motion-safe:` only.

## 8. Accessibility (the bar)

Semantic HTML (`header`/`nav`/`main`/`footer`, ordered headings); full keyboard operability; visible
`ring-primary` focus; managed focus on route change (done) and in dialogs (Radix); `aria-label` on
icon-only controls; decorative icons/paws are `aria-hidden`; meaningful `alt` (pet name) or `alt=""`
for decorative; the colour rule in §2; `prefers-reduced-motion` respected.

## 9. Theme

The app ships **light-only** — `index.html` sets `color-scheme: light` and no `dark` class is applied.
`darkMode: 'class'` and the `.dark` token block remain defined so a future toggle could re-enable a dark
theme without re-deriving the palette, but nothing applies it today. Don't hardcode colours — use tokens.

## 10. Do / Don't

- **Do** lead sections with an eyebrow + **orange** soft-rule heading; let pet photos dominate; use
  rounded pills + soft shadows; use orange generously on warm cream; sprinkle the paw motif.
- **Don't** use orange as body/label text on light (`accent` only); don't reintroduce chamfers/angular
  notches or all-uppercase display copy; don't introduce new hex values, fonts, or shadows outside the
  tokens; don't animate without a reduced-motion guard.

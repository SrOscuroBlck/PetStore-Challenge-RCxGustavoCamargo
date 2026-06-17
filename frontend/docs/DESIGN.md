# Design system

The visual language for the customer storefront. **Design DNA: an industrial-tech storefront** —
Robotic Crew's angular orange-on-navy identity carrying warm, photo-forward pet-commerce UX. We borrow
RC's brand patterns deliberately (it signals we follow patterns) and fuse them with the proven shopping
flows of Chewy / Petco / PetSmart, aligned to the challenge's needs (browse available pets → purchase →
cart → checkout).

This document is the source of truth for UI. Tokens live in `src/styles/tokens.css` +
`tailwind.config.ts`; motif utilities in `src/styles/index.css`. Build to these — don't invent ad-hoc
colors, type, or spacing.

## 1. Brand pillars

- **Industrial precision** — angular chamfers, thin tech-frame outlines, squared display type.
- **One hot accent** — a single vivid orange against navy/charcoal and white. Orange is earned, not
  sprinkled; it marks the brand, the primary action, and section emphasis.
- **Warm & photo-forward** — pets are the heroes. Generous whitespace, large imagery, friendly copy.
- **Confident, not noisy** — motion and decoration are purposeful; the catalog stays scannable.

## 2. Color

RGB-channel CSS variables (theme-swappable). Tailwind names in parentheses.

| Token (Tailwind) | Light | Dark | Use |
|---|---|---|---|
| `bg` | `#FBF8F4` (warm cream) | `#0E171E` | Page background |
| `card` | `#FFFFFF` | `#16212A` | Surfaces, cards |
| `border` | `#E9E3DC` (warm) | `#2A3742` | Hairlines, dividers |
| `fg` | `#16212A` | `#E7ECEF` | Body text |
| `muted` | `#6B645E` (warm) | `#95A3AC` | Secondary text |
| `ink` / `ink-fg` | `#16212A` / `#EDF2F5` | `#0B1218` / `#EDF2F5` | **Dark chrome** (header, footer, hero/feature bands) + its text |
| `primary` | `#F26522` | `#F26522` | **Brand orange** — fills, borders, icons, large display text |
| `primary-strong` | `#D9551A` | `#FF7A38` | Hover/pressed orange fills |
| `primary-fg` | `#16212A` | `#16212A` | Text/icon **on** an orange fill |
| `accent` | `#C2410C` | `#FB8A4C` | Orange **text** at body/label sizes (AA-safe) |
| `danger` / `success` | `#C6281C` / `#197943` | `#F87171` / `#4ADE80` | Status |

### Apply orange generously (it's the brand)

RC is **orange-forward**: large headings, solid buttons, eyebrows, dividers, and accents are all orange,
against warm cream + navy. The page should read warm and orange — **not** navy/white/gray with hairline
accents. Concretely: **light-section page/section headings are `primary` orange** (they're large, so
the 3:1 large-text bar is met), primary buttons are solid orange, eyebrows/rule-tabs/badges are orange,
species tiles carry an orange tint, and the hero uses an orange wash. Pets (photos) supply the rest of
the colour. Reserve dark `ink` for chrome (header/footer) and the occasional feature band.

### Accessibility colour rule (non-negotiable)

Orange at 16px on the cream page is ~3.3:1 — fine for large text/graphics, **not** for normal text. So:

- **`primary`** (vivid orange) → fills, borders, icons, and **large** display text (≥ ~24px) — including
  section/page headings, which should be orange on light surfaces.
- **`accent`** (`#C2410C`) → for orange-coloured **text** at body/label sizes on light surfaces (AA ~4.6:1). This is what `.eyebrow` uses.
- **`primary-fg`** (ink) → text/icons placed **on** an orange fill (dark-on-orange ≈ 7:1). Primary buttons are ink-on-orange, **not** white-on-orange.
- On **dark (`ink`) surfaces**, bright orange and white both have ample contrast — use freely.
- Never put small orange text on a light background. Body text is always `fg` / `muted`.

## 3. Typography

Two self-hosted families (bundled via `@fontsource`, no CDN — offline + CSP-safe):

- **Display — "Chakra Petch"** (`font-display`): squared techno face. Headings (`h1–h3` default to it),
  eyebrows, CTA labels, the wordmark, prices/stats. Often **uppercase** for labels/buttons.
- **Body/UI — "Inter"** (`font-sans`, default): all paragraphs, descriptions, inputs, secondary UI.
  Maximum legibility for the reading-heavy catalog.

Scale (Tailwind): display `text-3xl`/`text-4xl` (page titles), `text-2xl` (section), `text-lg`
(card titles); body `text-base`/`text-sm`; eyebrow `text-xs` uppercase, tracking `0.18em`. Keep headings
`tracking-tight`; keep body line-height comfortable (`leading-relaxed` for descriptions).

## 4. Space, shape, elevation

- **Radius** `--radius: 0.5rem` (moderate — friendlier than RC's hard edges, less rounded than Chewy).
- **Chamfer** `--chamfer: 14px` — the RC notched corner, applied via `.clip-corner` / `.clip-corner-tr`.
- **Container** `max-w-6xl`, padding `px-4 sm:px-6`. Mobile-first; everything works from ~320px.
- **Elevation** subtle only: `shadow-sm`/`shadow-lg` for cards/toasts. The tech aesthetic favours
  **borders over shadows** — prefer a 1px `border-border` and an orange border on hover/active.

## 5. Signature motifs (RC) — and how to use them

Codified as utilities/components so usage stays consistent:

| Motif | How | Where |
|---|---|---|
| **Eyebrow label** | `.eyebrow` (uppercase orange `accent`, display, tracked) | Above every section/page heading |
| **Rule-tab** | `.rule-tab` (orange underline + raised tab) on a heading | Under section/page headings |
| **Chamfered corner** | `.clip-corner` / `.clip-corner-tr` | Cards, images, badges, tiles — **never focusable controls** (clip-path also clips the focus ring) |
| **Chevron bullets** | `.list-chevron` on a `<ul>` | Feature/benefit lists |
| **Tech-frame / bracket** | thin `border` panel; optional pointer | Hero/feature callouts |
| **Logo watermark** | oversized translucent `Logo` mark, `aria-hidden` | Decorative background on dark bands |
| **Dark feature band** | `bg-ink text-ink-fg` section | Hero, "why", emphasis bands between light content |

> Constraint: chamfers clip outlines, so the focus ring would be cut on a clipped control. Apply
> chamfers to **decorative/surface** elements; keep interactive controls on `rounded-sm` with a visible
> `ring-primary` focus.

## 6. Components

Build thin + typed; data fetched in containers, passed to presentation (CLAUDE.md). Specs:

- **Button** (`components/ui/Button.tsx`) — `primary` (orange fill, ink label), `secondary` (bordered,
  orange border on hover), `ghost`, `onDark` (outline on ink). Labels are uppercase display.
- **Header / top bar** — `bg-ink` band: `Logo` left; **search** + **cart (icon + count badge)** right
  (search/cart land in #2/#3). Sticky. Cart count is an orange badge.
- **Shop-by-species tiles** — `CAT | DOG | FROG` (from `PublicPet.species`). Chamfered `card` tiles,
  orange border + lift on hover; the petstore "who are you shopping for?" pattern, RC-styled.
- **Pet card** (#2) — chamfered photo (`pictureUrl`, lazy, sized, placeholder on 404); species **badge**
  (orange pill, ink text); name (display); age ("3 yrs") + short description (`muted`); **Buy** (primary)
  + **Add to cart** (secondary). Hover: lift + orange border. States: skeleton, sold (dimmed + "Sold").
- **Cart drawer** (#3) — Radix Dialog from the right; `ink` header, line items with thumbnail,
  **Checkout** primary CTA; animated success / named-unavailable-pets error.
- **Badges** — small uppercase display pills: species (orange), status (sold = muted, available = success).
- **Toasts** (`components/ui/Toast`) — Radix; `card` with a left accent border (`primary`/`danger`/
  `success`), display title. Used for purchase/checkout/error feedback.
- **Inputs** (`components/ui/TextField`) — bordered, `card` bg, labelled, `aria-invalid`/`-describedby`.
- **States** — every async surface designs loading (skeleton), empty (friendly, watermark), and error
  (banner/toast with the server's human-readable message). Never a blank screen or raw error.

## 7. Motion

Framer Motion, **purposeful**: card hover lift, list enter/exit (stagger), page/section fade-in-up
(`animate-fade-in-up`), cart open/close, optimistic purchase (pet animates out). No gratuitous motion.
**Always gate on reduced motion** via `useReducedMotion` (`src/lib/a11y/useReducedMotion.ts`) — when
true, disable transforms/opacity transitions and render final state.

## 8. Accessibility (the bar)

Semantic HTML (`header`/`nav`/`main`/`footer`, ordered headings); full keyboard operability; visible
`ring-primary` focus; managed focus on route change (done) and in dialogs (Radix); `aria-label` on
icon-only controls; meaningful `alt` (pet name) or `alt=""` for decorative; the colour rule in §2;
contrast holds in both themes; `prefers-reduced-motion` respected.

## 9. Dark mode

`darkMode: 'class'`; tokens already define the dark palette; `.dark` seeded from the OS preference
(no-flash inline script in `index.html`). A user toggle is a #4 task. Don't hardcode colours — use
tokens so both themes track automatically.

## 10. Do / Don't

- **Do** lead sections with an eyebrow + **orange** rule-tab heading; let pet photos dominate; use orange
  generously and deliberately — headings, primary actions, eyebrows, badges, tints — on warm cream.
- **Don't** use orange as body/label text on light (`accent` only); don't chamfer focusable controls;
  don't introduce new hex values, fonts, or shadows outside the tokens; don't animate without a reduced-
  motion guard.

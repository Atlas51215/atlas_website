# Atlas Design Tokens

Visual design reference for the Atlas blog platform. Inspired by the design language of [Fireart Studio](https://fireart.studio/) — deep blacks, bold orange accents, modern sans-serif type, generous spacing, and rounded corners. All fonts are free Google Fonts equivalents of Fireart's custom typefaces.

See `docs/style-guide.md` for Go coding conventions.

---

## Fonts

| Role | Font | Weight | Fallback |
|------|------|--------|----------|
| Headings & UI labels | Plus Jakarta Sans | 500, 600 | system-ui, sans-serif |
| Body UI (nav, inputs, buttons) | Inter | 400, 600 | system-ui, sans-serif |
| Prose body text | Georgia | 400 | Times New Roman, serif |

**Google Fonts import** (top of `style.css`):
```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600&family=Plus+Jakarta+Sans:wght@500;600&display=swap');
```

**CSS variables:**
```css
--font-ui:      'Inter', system-ui, sans-serif;
--font-heading: 'Plus Jakarta Sans', system-ui, sans-serif;
--font-body:    Georgia, 'Times New Roman', serif;
```

Georgia is kept for prose because long-form reading benefits from a serif. All chrome (nav, buttons, inputs, labels) uses Inter for a clean, modern feel.

---

## Type Scale

| Element | Size | Line Height | Font |
|---------|------|-------------|------|
| `h1` in main content | `2.25rem` (36px) | 1.25 | Plus Jakarta Sans |
| `h2` in main content | `1.65rem` (26.4px) | 1.3 | Plus Jakarta Sans |
| `h3` in main content | `1.3rem` (20.8px) | inherited | Plus Jakarta Sans |
| Body prose | `1rem` (16px) | 1.8 | Georgia |
| Nav links | `0.95rem` (15.2px) | — | Inter |
| Nav sub-links | `0.875rem` (14px) | — | Inter |
| Nav group labels | `0.75rem` (12px) | — | Inter 700 |
| Search input | `0.9rem` (14.4px) | — | Inter |
| Logo wordmark | `1.5rem` (24px) | — | Plus Jakarta Sans bold |

---

## Color Tokens

| Variable | Dark Value | Light Value | Usage |
|----------|-----------|-------------|-------|
| `--bg` | `#000000` | `#F5F5F6` | Page background |
| `--sidebar-bg` | `#141414` | `#FAFAFA` | Sidebar background |
| `--surface` | `#19191A` | `#FFFFFF` | Cards, inputs, elevated surfaces |
| `--text` | `#FAFAFA` | `#19191A` | Primary text |
| `--text-muted` | `#76757F` | `#76757F` | Secondary text, placeholders, labels |
| `--accent` | `#dc2626` | `#dc2626` | Links, active states, button backgrounds |
| `--accent-hover` | `#ef4444` | `#b91c1c` | Accent on hover |
| `--border` | `#323234` | `#F5F5F6` | Borders, dividers |
| `--input-bg` | `#19191A` | `#FFFFFF` | Form input backgrounds |
| `--input-border` | `#323234` | `#76757F` | Form input borders |
| `--nav-hover` | `#19191A` | `#F5F5F6` | Nav item hover background |

**Accessibility note:** The red accent (`#dc2626`) on a black background has a contrast ratio of ~4.5:1, which passes WCAG AA for both large and small text. In light mode on white, it passes AA for large text. Use the accent for interactive elements, headings, and hover states.

---

## Spacing

| Name | Value | Typical use |
|------|-------|-------------|
| Section | `3rem` (48px) | Main content top/bottom padding |
| Large gap | `2rem` (32px) | Nav section gaps |
| Standard | `1.5rem` (24px) | Heading margins, nav padding |
| Small | `0.75rem` (12px) | List margins |
| Micro | `0.2rem` (3.2px) | Nav item list gaps |

---

## Border Radius

| Variable | Value | Use |
|----------|-------|-----|
| `--radius-pill` | `120px` | Pill-shaped buttons (`.btn-pill`) |
| `--radius-card` | `16px` | Cards and large containers |
| `--radius-sm` | `8px` | Small interactive elements (toggle button, code blocks) |
| `--radius-input` | `5px` | Inputs, standard buttons |

---

## Shadows

| Variable | Value | Use |
|----------|-------|-----|
| `--shadow-sm` | `0 4px 8px rgba(0,0,0,0.08)` | Subtle lift — sidebar toggle, cards |
| `--shadow-deep` | `12px 12px 50px rgba(0,0,0,0.4)` | Mobile sidebar overlay |

Shadows are intentionally subtle in dark mode — `--shadow-sm` against a true black background is nearly imperceptible, providing lift without visual noise.

---

## Transitions

| Context | Duration | Variable |
|---------|----------|----------|
| Structural (sidebar open/close) | `0.3s ease` | — |
| Hover micro-interactions (links, buttons, inputs) | `0.18s ease` | — |
| Button background/color | `var(--transition)` = `0.4s ease` | `--transition` |

Hover states use `0.18s` rather than the full `0.4s` so nav links and inputs feel responsive, not sluggish.

---

## Buttons

```css
.btn          /* base: font, size, padding, cursor */
.btn-primary  /* orange background, white text, hover → black */
.btn-pill     /* applies --radius-pill for pill shape */
```

Combine classes: `class="btn btn-primary"` for standard, `class="btn btn-primary btn-pill"` for pill variant.

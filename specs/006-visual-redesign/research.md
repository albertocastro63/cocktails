# Research: Visual Redesign

## Color Palette

**Decision**: Amber accent (`amber-500` / `amber-600`) over a warm stone neutral base (`stone-50` backgrounds, `stone-900` headings, `stone-700` body text).  
**Rationale**: Amber is the natural color of cocktails — whiskey, honey syrup, aged spirits. It reads as warm, inviting, and premium without being loud. Stone neutrals are softer than pure grays and pair naturally with amber. This combination achieves the "modern-minimal, single bold accent" direction confirmed in the spec.  
**Alternatives considered**: Indigo (current — feels too corporate/SaaS), teal (too clinical), violet (lacks food/drink warmth).

## Navigation Bar

**Decision**: Dark navigation bar — `bg-stone-900` with white text, amber hover states on links, and an amber-filled "New Recipe" button.  
**Rationale**: A dark nav creates strong visual contrast with the light page body, immediately orienting users and making the site feel designed rather than default. This is a common pattern in modern recipe and food apps.  
**Alternatives considered**: Amber-filled nav — tested visually; amber at full saturation across the full nav width is heavy and competes with content. Dark nav with amber accents is more balanced.

## Page Backgrounds

**Decision**: All pages use `bg-stone-50` as the default body background instead of the browser/Tailwind default white.  
**Rationale**: Pure white (`#ffffff`) feels clinical. Stone-50 (`#fafaf9`) is imperceptibly warmer but prevents the "blank Word document" look. All cards remain `bg-white` so they visually pop off the stone background.  
**Alternatives considered**: Keeping default white — rejected because it undermines the "inviting" goal without adding implementation complexity.

## Card Style

**Decision**: Recipe cards use `bg-white rounded-2xl border border-stone-200 shadow-sm` with an amber left accent border (`border-l-4 border-l-amber-400`) and `hover:shadow-lg` transition.  
**Rationale**: `rounded-2xl` (16px radius) reads as modern vs. the current `rounded-lg`. The amber left border is a lightweight, distinctive accent that brands each card without overwhelming it. The shadow lift on hover gives satisfying feedback.  
**Alternatives considered**: Gradient card backgrounds — too busy for a list of many items. Plain border cards with no shadow — too flat, loses depth cue.

## Hero / Page Headers

**Decision**: The Home page features a full-width hero band (`bg-gradient-to-br from-stone-900 to-stone-800`) with a large white heading and amber-tinted subtext. All other pages use an in-page `h1` on the stone-50 background (no full-width hero for inner pages).  
**Rationale**: A dedicated hero only on the home page creates a clear hierarchy (landing vs. utility pages). Inner pages don't need the visual weight of a hero band — clean headings on the light background are sufficient.  
**Alternatives considered**: Hero band on every page — creates visual fatigue; the nav already provides top-of-page orientation.

## Typography

**Decision**: No custom web font added. Use Tailwind's default sans-serif stack (`font-sans`). Differentiate hierarchy through weight and size only: headings `text-3xl font-bold text-stone-900`, section labels `text-sm font-semibold uppercase tracking-widest text-amber-600`, body `text-stone-700`.  
**Rationale**: Adding a Google Font (e.g., Inter) would improve perceived quality but adds a network dependency and build complexity outside the scope of this feature. The existing system font stack is already high quality on modern OS. Uppercase tracked section labels are a widely used modern design pattern that creates strong hierarchy with zero additional assets.  
**Alternatives considered**: Plus Jakarta Sans or Inter via Google Fonts — deferred to a future typography enhancement; not blocked by this feature.

## Buttons

**Decision**: Primary buttons: `bg-amber-500 hover:bg-amber-600 text-white font-medium rounded-xl px-4 py-2 transition-colors`. Secondary/outline buttons: `border border-stone-300 text-stone-700 hover:border-amber-400 hover:text-amber-600 rounded-xl px-3 py-1 text-sm transition-colors`. Destructive buttons: `border border-red-300 text-red-600 hover:bg-red-50 rounded-xl px-3 py-1 text-sm`.  
**Rationale**: A single primary button style ensures visual consistency. The rounded-xl (12px) matches the card radius language. Amber fills for primary, outline for secondary, red-tinted for destructive — standard three-tier button hierarchy.

## Form Inputs

**Decision**: All inputs: `border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent`.  
**Rationale**: Amber focus ring ties inputs into the overall accent color system. `rounded-xl` matches button/card language.

## Tailwind Config

**Decision**: Add a `stone` and `amber` palette shortcut in `tailwind.config.js` — no custom color tokens needed since both `stone` and `amber` are default Tailwind palette colors. The `extend` block remains empty; classes are used directly.  
**Rationale**: Using built-in Tailwind colors avoids the need to maintain custom hex values. Both `stone` and `amber` are full palettes already in Tailwind v3.

## Scope of File Changes

Files modified (implementation phase):
- `frontend/tailwind.config.js` — no changes needed (stone + amber already in default palette)
- `frontend/src/index.css` — add `body { @apply bg-stone-50; }` base style
- `frontend/src/main.js` — nav redesign (dark stone, amber accents)
- `frontend/src/pages/Home.js` — hero band, updated headings
- `frontend/src/pages/RecipeList.js` — heading style
- `frontend/src/pages/RecipeDetail.js` — section label style, button style
- `frontend/src/pages/Login.js` — form card style, button
- `frontend/src/pages/RecipeForm.js` — input/button styles
- `frontend/src/components/RecipeCard.js` — card redesign
- `frontend/src/components/EmptyState.js` — updated empty state
- `frontend/src/components/SearchBar.js` — input style
- `frontend/src/components/IngredientList.js` — minor divider/text style

Test files updated to reflect new class assertions:
- `frontend/src/components/RecipeCard.test.js`
- `frontend/src/components/EmptyState.test.js`
- `frontend/src/components/SearchBar.test.js`
- `frontend/src/components/IngredientList.test.js`

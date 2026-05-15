# Research: Notes Rendered Markup Styling

**Feature**: 007-notes-markup-render | **Date**: 2026-05-14

## Finding 1: Current rendering state

**Decision**: The root cause is that `@tailwindcss/typography` is not installed, and the read-only surfaces (`RecipeDetail`, `Home`) have no typography classes at all.

**Evidence**:
- `MarkdownEditor.js:31` — preview div already has `prose` class, but the Typography plugin is absent from `tailwind.config.js` (plugins array is empty), so the class has no effect.
- `RecipeDetail.js:93` — notes div has `text-stone-700` only; no typography styling.
- `Home.js:95` — notes div has `text-stone-700` only; no typography styling.
- `tailwind.config.js` — `plugins: []`; no Typography plugin configured.
- `package.json` — `@tailwindcss/typography` is absent from both `dependencies` and `devDependencies`.

**Rationale**: A single install + config change unlocks the `prose` class that is already written in `MarkdownEditor.js`, and two class-string updates bring `RecipeDetail` and `Home` into sync.

---

## Finding 2: Styling approach selection

**Decision**: Use `@tailwindcss/typography` with the `prose prose-stone max-w-none` class combination on all rendered notes containers.

**Rationale**:
- `@tailwindcss/typography` is the standard, maintained Tailwind solution for styling arbitrary HTML generated from markdown. It handles all element types (h1–h6, ul, ol, li, blockquote, code, pre, hr, a, strong, em, table) in a single class.
- `prose-stone` applies the stone color palette (stone-900 for headings, stone-700 for body text), which aligns with the existing design system and the clarification to use standard typography colors rather than amber foreground colors.
- `max-w-none` removes the default prose max-width constraint (`65ch`), which would otherwise impose a content width narrower than the surrounding containers.
- Existing `text-stone-700` on the notes div is superseded by `prose prose-stone` and should be removed to avoid redundancy.

**Alternatives considered**:

| Alternative | Why Rejected |
|-------------|-------------|
| Custom hand-written CSS for each element | Higher maintenance cost; must be kept in sync with Tailwind's responsive/dark-mode utilities; reinvents a wheel Tailwind already ships |
| Post-processing rendered HTML to inject Tailwind classes per element | Complex, fragile, and breaks DOMPurify's sanitisation pipeline |
| `prose-amber` or custom amber overrides | Rejected per clarification Q1 (use standard typography colors, not amber foreground on prose elements) |
| `prose-sm` modifier | Not needed; the notes field is not displayed in a compact context that warrants smaller type |

---

## Finding 3: Test strategy

**Decision**: Assert that the notes container element has the `prose` CSS class in jsdom unit tests; do not attempt to assert computed visual styles.

**Rationale**:
- jsdom does not compute CSS — it does not resolve Tailwind class names into applied styles. Visual assertions (`getComputedStyle`) would always return empty strings and provide no value.
- Asserting `element.classList.contains('prose')` (or `className.includes('prose')`) is a valid contract test: it verifies the implementation wires up the styling hook correctly, and the visual outcome follows from the plugin being installed.
- This approach is consistent with the existing `RecipeDetail.test.js` pattern that asserts `h1.className` contains `text-stone-900`.

**New tests to add**:
1. `MarkdownEditor.test.js` — after clicking Preview, the `[data-preview]` div has class `prose`.
2. `RecipeDetail.test.js` — when recipe has notes, the notes container div has class `prose`.
3. `Home.test.js` — when recipe has notes, the notes container div has class `prose`.

---

## Finding 4: Class change summary

| File | Old class string | New class string |
|------|-----------------|-----------------|
| `MarkdownEditor.js` (notes container) | `prose w-full border border-gray-200 rounded-lg px-3 py-2 min-h-[4.5rem] text-gray-700` | `prose prose-stone max-w-none overflow-x-auto w-full border border-gray-200 rounded-lg px-3 py-2 min-h-[4.5rem]` |
| `RecipeDetail.js` (notes container) | `text-stone-700` | `prose prose-stone max-w-none overflow-x-auto` |
| `Home.js` (notes container) | `text-stone-700` | `prose prose-stone max-w-none overflow-x-auto` |

The `text-gray-700` on the preview div and `text-stone-700` on the page divs are both superseded by `prose-stone`, which sets body text to stone-700 automatically. Removing them avoids specificity conflicts. `overflow-x-auto` ensures wide markdown tables scroll horizontally rather than overflowing their container (spec edge case: "markdown table must render with visible rows/columns and not break surrounding layout").

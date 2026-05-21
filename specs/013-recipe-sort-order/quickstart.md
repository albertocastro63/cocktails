# Quickstart: Recipe Sort Order

**Feature**: 013-recipe-sort-order
**Date**: 2026-05-21

---

## Prerequisites

- Frontend dev server running: `cd frontend && npm run dev`
- At least 3 recipes in the system with names starting with different letters
- A modern browser and optionally a screen reader (e.g., macOS VoiceOver: Cmd+F5)

---

## Verification Scenarios

### SC-001 — A→Z Sort Reorders the List

```
1. Open http://localhost:5173 (or your dev server URL)
2. Navigate to the all-recipes page
3. Click the "A→Z" sort button
4. Verify: recipes are listed alphabetically, A first, Z last
5. Verify: the "A→Z" button appears highlighted (amber background)
6. Verify: the "Z→A" button appears un-highlighted (white/stone background)
```

**Expected**: Recipe cards reorder immediately; first card's name comes first alphabetically.

---

### SC-002 — Z→A Sort Reorders the List

```
1. On the all-recipes page, click the "Z→A" sort button
2. Verify: recipes are listed in reverse alphabetical order, Z first, A last
3. Verify: the "Z→A" button appears highlighted
4. Verify: switching back to "A→Z" immediately reverses the order
```

**Expected**: Bidirectional switching works; active button always visually distinguished.

---

### SC-003 — Default State on Page Load

```
1. Navigate away from the all-recipes page and return (or hard-refresh)
2. Verify: neither "A→Z" nor "Z→A" is highlighted
3. Verify: recipes appear in the server-returned default order
```

**Expected**: No sort direction is pre-selected on load.

---

### SC-004 — Keyboard Navigation

```
1. On the all-recipes page, use Tab to navigate to the sort controls
2. Verify: the first sort button receives a visible focus ring (amber outline)
3. Press Tab again: focus moves to the second sort button
4. Press Enter or Space: the focused sort button activates, list reorders
5. Verify: you never needed a mouse
```

**Expected**: Both sort buttons are keyboard-accessible with visible focus indicators.

---

### SC-005 — Screen Reader Announcement (VoiceOver)

```
1. Enable VoiceOver (macOS: Cmd+F5)
2. Navigate to the sort buttons using Tab
3. Listen: VoiceOver should announce the button label and pressed state
   e.g. "A to Z, toggle button, not pressed" / "Z to A, toggle button, pressed"
4. Activate the button with Space; verify the state change is announced
```

**Expected**: Screen reader correctly identifies each button and its active/inactive state.

---

### SC-006 — Empty List Graceful Handling

```
1. In a test environment with no recipes (or filter search to return 0 results)
2. Attempt to activate the A→Z or Z→A sort button
3. Verify: the empty state message is still shown; no JavaScript errors in the console
4. Verify: both sort buttons remain visible and the layout is intact
```

**Expected**: Sorting an empty list produces no errors and no layout breakage.

---

### SC-007 — Sort Applies Alongside Search (integration)

```
1. Type a search term that returns 3+ recipes
2. Click "A→Z"
3. Verify: the filtered results are sorted alphabetically
4. Click "Z→A"
5. Verify: the filtered results reorder in reverse
```

**Expected**: Sort and search compose correctly — sort operates on the currently displayed (filtered) results.

---

## Performance Check

```bash
# In a browser devtools console on the recipes page with 50+ recipes loaded:
console.time('sort');
document.querySelector('[data-dir="asc"]').click();
console.timeEnd('sort');
```

**Expected**: Sort completes in under 200 ms (SC-001 criterion).

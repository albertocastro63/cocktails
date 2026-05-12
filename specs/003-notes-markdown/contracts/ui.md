# UI Component Contracts: Markdown Notes Editor

**Version**: v1 | **Date**: 2026-05-11 | **Branch**: `003-notes-markdown`

---

## `renderMarkdown(text)`

**Location**: `frontend/src/utils/markdown.js`  
**Type**: Pure utility function

### Signature

```
renderMarkdown(text: string) → string (sanitised HTML)
```

### Behaviour

| Input | Output |
|-------|--------|
| Empty string `""` | Empty string `""` |
| Plain text (no markdown) | Text wrapped in `<p>` tags |
| CommonMark markdown | Equivalent rendered HTML |
| `![alt](url)` (image) | Empty string — image tag stripped |
| `[text](url)` (link) | `<a href="url" target="_blank" rel="noopener noreferrer">text</a>` |
| `<script>alert(1)</script>` | Empty — stripped by DOMPurify |
| Any other raw HTML tags | Stripped by DOMPurify |

### Constraints

- MUST NOT return any `<img>` tags in output.
- All `<a>` tags MUST have `target="_blank"` and `rel="noopener noreferrer"`.
- Output MUST be safe to assign to `element.innerHTML` without XSS risk.
- Function MUST be synchronous (no async, no network calls).

---

## `MarkdownEditor({ name, placeholder, value })`

**Location**: `frontend/src/components/MarkdownEditor.js`  
**Type**: DOM component factory (returns an `HTMLElement`)

### Signature

```
MarkdownEditor({ name?: string, placeholder?: string, value?: string }) → HTMLElement
```

### Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `name` | string | `"notes"` | The `name` attribute on the internal `<textarea>` |
| `placeholder` | string | `""` | Placeholder text for the textarea |
| `value` | string | `""` | Initial markdown content to prefill |

### DOM Structure

```
div.markdown-editor                          ← root element returned by component
├── div.markdown-editor-toolbar              ← top bar containing toggle button
│   └── button[type="button"][aria-pressed]  ← "Preview" / "Edit" label toggles; aria-pressed reflects state
├── textarea[name][placeholder]              ← visible in edit mode, hidden in preview mode
└── div.markdown-preview                     ← visible in preview mode, hidden in edit mode
```

### State & Behaviour

| State | Textarea | Preview div | Button label | `aria-pressed` |
|-------|----------|-------------|--------------|----------------|
| Edit mode (initial) | visible | hidden | "Preview" | `"false"` |
| Preview mode | hidden (present in DOM) | visible, innerHTML = renderMarkdown(textarea.value) | "Edit" | `"true"` |

- Clicking the toggle button switches between states.
- The button MUST have `aria-pressed="false"` in edit mode and `aria-pressed="true"` in preview mode (WCAG 2.1 AA — SC 4.1.2 Name, Role, Value).
- Switching from preview → edit MUST restore the exact string that was in the textarea before entering preview.
- The textarea MUST remain in the DOM (hidden via CSS) during preview mode so that `form.querySelector('[name="notes"]').value` continues to return the correct raw markdown value.
- The component MUST emit no events; its value is consumed passively via the textarea's `name` attribute by the parent form.

### Edge Cases

- If `value` is empty and user presses Preview: show empty preview div without error.
- If `renderMarkdown` produces empty output: show empty preview div (do not fall back to raw text in this case — empty is correct for empty input).

---

## Integration Points

### `RecipeForm.js`

Replace the existing notes `<textarea name="notes">` construction with:

```js
form.appendChild(MarkdownEditor({
  name: 'notes',
  placeholder: 'Personal notes, substitutions, tips…',
  value: recipe?.notes || '',   // prefill on edit
}));
```

Form submission continues to read `form.querySelector('[name="notes"]').value` — no change required.

### `RecipeDetail.js` and `Home.js`

Replace:
```js
p.className = 'text-gray-700 whitespace-pre-wrap';
p.textContent = recipe.notes;
```

With:
```js
div.className = 'prose text-gray-700';
div.innerHTML = renderMarkdown(recipe.notes);
```

The `if (recipe.notes)` guard that hides the section when notes are empty is preserved.

# Data Model: Notes Rendered Markup Styling

**Feature**: 007-notes-markup-render | **Date**: 2026-05-14

## No data model changes

This feature is entirely display-only. The `notes` field on a recipe is already stored as raw markdown text and is unchanged by this feature. No new fields, entities, or relationships are introduced.

## Affected display contract

The only contract change is in how the `notes` string is presented to the user in the browser.

| Surface | Before | After |
|---------|--------|-------|
| Editor preview (`MarkdownEditor`) | Rendered HTML with no effective typography styles (plugin missing) | Rendered HTML with `prose prose-stone max-w-none` typography styles |
| Recipe detail page (`RecipeDetail`) | Rendered HTML wrapped in `<div class="text-stone-700">` (no element-level styling) | Rendered HTML wrapped in `<div class="prose prose-stone max-w-none">` |
| Homepage featured recipe (`Home`) | Rendered HTML wrapped in `<div class="text-stone-700">` (no element-level styling) | Rendered HTML wrapped in `<div class="prose prose-stone max-w-none">` |

The stored value and the API response for `recipe.notes` are identical before and after. The renderer (`renderMarkdown`) is unchanged.

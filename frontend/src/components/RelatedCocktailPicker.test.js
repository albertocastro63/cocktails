import { describe, it, expect } from 'vitest';
import { RelatedCocktailPicker } from './RelatedCocktailPicker.js';

const NAMES = [
  { id: 'lh', name: 'Left Hand' },
  { id: 'rh', name: 'Right Hand' },
  { id: 'neg', name: 'Negroni' },
];

function typeInto(el, value) {
  const input = el.querySelector('input[role="combobox"]');
  input.value = value;
  input.dispatchEvent(new Event('input', { bubbles: true }));
  return input;
}

function optionTexts(el) {
  return [...el.querySelectorAll('[role="option"]')].map((o) => o.textContent);
}

describe('RelatedCocktailPicker', () => {
  it('filters names by case-insensitive substring', () => {
    const el = RelatedCocktailPicker({ names: NAMES });
    typeInto(el, 'HAND');
    const opts = optionTexts(el);
    expect(opts).toContain('Left Hand');
    expect(opts).toContain('Right Hand');
    expect(opts).not.toContain('Negroni');
  });

  it('shows no options for empty input', () => {
    const el = RelatedCocktailPicker({ names: NAMES });
    typeInto(el, '');
    expect(optionTexts(el)).toHaveLength(0);
  });

  it('excludes the current recipe from suggestions', () => {
    const el = RelatedCocktailPicker({ names: NAMES, currentId: 'neg' });
    typeInto(el, 'negroni');
    expect(optionTexts(el)).not.toContain('Negroni');
  });

  it('excludes already-selected cocktails', () => {
    const el = RelatedCocktailPicker({ names: NAMES, selectedIds: ['lh'] });
    typeInto(el, 'hand');
    const opts = optionTexts(el);
    expect(opts).not.toContain('Left Hand');
    expect(opts).toContain('Right Hand');
  });

  it('adds a cocktail via ArrowDown + Enter and exposes selected ids', () => {
    const el = RelatedCocktailPicker({ names: NAMES });
    const input = typeInto(el, 'hand'); // options sorted: Left Hand, Right Hand
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    expect(el.getSelectedIds()).toContain('lh');
    expect(el.querySelector('[data-chip="lh"]')).toBeTruthy();
  });

  it('renders initial chips from selectedIds and can remove them', () => {
    const el = RelatedCocktailPicker({ names: NAMES, selectedIds: ['lh'] });
    const chip = el.querySelector('[data-chip="lh"]');
    expect(chip).toBeTruthy();
    chip.querySelector('button').click();
    expect(el.getSelectedIds()).not.toContain('lh');
  });
});

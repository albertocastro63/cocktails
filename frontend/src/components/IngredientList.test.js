import { describe, it, expect } from 'vitest';
import { IngredientList } from './IngredientList.js';

const ingredients = [
  { name: 'rum', quantity: '50', unit: 'ml' },
  { name: 'mint', quantity: '10', unit: 'leaves' },
  { name: 'sugar', quantity: '2', unit: '' },
];

describe('IngredientList', () => {
  it('renders each ingredient name', () => {
    const el = IngredientList({ ingredients });
    expect(el.textContent).toContain('rum');
    expect(el.textContent).toContain('mint');
    expect(el.textContent).toContain('sugar');
  });

  it('renders quantity and unit', () => {
    const el = IngredientList({ ingredients });
    expect(el.textContent).toContain('50');
    expect(el.textContent).toContain('ml');
    expect(el.textContent).toContain('10');
    expect(el.textContent).toContain('leaves');
  });

  it('renders empty state for no ingredients', () => {
    const el = IngredientList({ ingredients: [] });
    expect(el).not.toBeNull();
  });

  it('ingredient name span has text-stone-800 class not legacy text-gray-800', () => {
    const el = IngredientList({ ingredients });
    const nameSpan = el.querySelector('span.font-medium');
    expect(nameSpan).not.toBeNull();
    expect(nameSpan.className).toContain('text-stone-800');
  });
});

describe('IngredientList — base spirit highlight (T011)', () => {
  it('base spirit ingredient li contains (base spirit) label', () => {
    const ings = [
      { name: 'Rye', quantity: '60', unit: 'ml', is_base_spirit: true },
      { name: 'Vermouth', quantity: '30', unit: 'ml' },
    ];
    const el = IngredientList({ ingredients: ings });
    expect(el.textContent).toContain('(base spirit)');
  });

  it('base spirit ingredient name span has font-semibold text-stone-900 classes', () => {
    const ings = [
      { name: 'Rye', quantity: '60', unit: 'ml', is_base_spirit: true },
      { name: 'Vermouth', quantity: '30', unit: 'ml' },
    ];
    const el = IngredientList({ ingredients: ings });
    const boldSpan = el.querySelector('.font-semibold');
    expect(boldSpan).not.toBeNull();
    expect(boldSpan.className).toContain('text-stone-900');
    expect(boldSpan.textContent).toBe('Rye');
  });

  it('(base spirit) label appears only on the flagged ingredient, not on neighbours', () => {
    const ings = [
      { name: 'Rye', quantity: '60', unit: 'ml', is_base_spirit: true },
      { name: 'Vermouth', quantity: '30', unit: 'ml' },
      { name: 'Bitters', quantity: '2', unit: 'dashes' },
    ];
    const el = IngredientList({ ingredients: ings });
    const items = [...el.querySelectorAll('li')];
    const labelCount = items.filter(li => li.textContent.includes('(base spirit)')).length;
    expect(labelCount).toBe(1);
  });

  it('all items are rendered identically when no ingredient has is_base_spirit', () => {
    const ings = [
      { name: 'Rum', quantity: '50', unit: 'ml' },
      { name: 'Lime', quantity: '15', unit: 'ml' },
    ];
    const el = IngredientList({ ingredients: ings });
    expect(el.textContent).not.toContain('(base spirit)');
    expect(el.querySelector('.font-semibold')).toBeNull();
  });
});

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
});

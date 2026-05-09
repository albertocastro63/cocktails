import { describe, it, expect } from 'vitest';
import { RecipeCard } from './RecipeCard.js';

const recipe = {
  id: 'r1',
  name: 'Mojito',
  ingredients: [
    { name: 'rum', quantity: '50', unit: 'ml' },
    { name: 'mint', quantity: '10', unit: 'leaves' },
  ],
};

describe('RecipeCard', () => {
  it('renders the recipe name', () => {
    const el = RecipeCard({ recipe });
    expect(el.textContent).toContain('Mojito');
  });

  it('renders the ingredient count', () => {
    const el = RecipeCard({ recipe });
    expect(el.textContent).toContain('2');
  });

  it('links to the recipe detail page', () => {
    const el = RecipeCard({ recipe });
    const link = el.querySelector('a');
    expect(link).not.toBeNull();
    expect(link.getAttribute('href')).toContain('r1');
  });
});

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

describe('RecipeCard - ingredient popover', () => {
  it('shows popover with ingredient names on mouseenter', () => {
    const el = RecipeCard({ recipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = el.querySelector('[data-popover]');
    expect(popover).not.toBeNull();
    expect(popover.textContent).toContain('rum');
    expect(popover.textContent).toContain('mint');
  });

  it('hides popover on mouseleave', () => {
    const el = RecipeCard({ recipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    el.dispatchEvent(new MouseEvent('mouseleave'));
    const popover = el.querySelector('[data-popover]');
    expect(popover).toBeNull();
  });

  it('removes first tile popover when mouse leaves it and enters a second tile', () => {
    const el1 = RecipeCard({ recipe });
    const el2 = RecipeCard({ recipe });
    el1.dispatchEvent(new MouseEvent('mouseenter'));
    el1.dispatchEvent(new MouseEvent('mouseleave'));
    el2.dispatchEvent(new MouseEvent('mouseenter'));
    expect(el1.querySelector('[data-popover]')).toBeNull();
    expect(el2.querySelector('[data-popover]')).not.toBeNull();
  });

  it('tile link href is unchanged after popover is shown', () => {
    const el = RecipeCard({ recipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const link = el.querySelector('a');
    expect(link).not.toBeNull();
    expect(link.getAttribute('href')).toContain('r1');
  });
});

describe('RecipeCard - ingredient truncation', () => {
  it('shows all ingredients when recipe has exactly 5', () => {
    const r = { id: 'r2', name: 'T', ingredients: [
      { name: 'a' }, { name: 'b' }, { name: 'c' }, { name: 'd' }, { name: 'e' },
    ]};
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = el.querySelector('[data-popover]');
    expect(popover.textContent).toContain('a');
    expect(popover.textContent).toContain('e');
    expect(popover.textContent).not.toContain('…');
  });

  it('shows first 5 and ellipsis when recipe has exactly 6 ingredients', () => {
    const r = { id: 'r3', name: 'T', ingredients: [
      { name: 'a' }, { name: 'b' }, { name: 'c' }, { name: 'd' }, { name: 'e' }, { name: 'f' },
    ]};
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = el.querySelector('[data-popover]');
    expect(popover.textContent).toContain('a');
    expect(popover.textContent).toContain('e');
    expect(popover.textContent).not.toContain('f');
    expect(popover.textContent).toContain('…');
  });

  it('shows first 5 and ellipsis when recipe has more than 6 ingredients', () => {
    const r = { id: 'r4', name: 'T', ingredients: [
      { name: 'a' }, { name: 'b' }, { name: 'c' }, { name: 'd' }, { name: 'e' },
      { name: 'f' }, { name: 'g' },
    ]};
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = el.querySelector('[data-popover]');
    expect(popover.textContent).not.toContain('f');
    expect(popover.textContent).toContain('…');
  });
});

describe('RecipeCard - empty ingredients', () => {
  it('shows "No ingredients listed." when recipe has no ingredients', () => {
    const r = { id: 'r5', name: 'Empty', ingredients: [] };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = el.querySelector('[data-popover]');
    expect(popover).not.toBeNull();
    expect(popover.textContent).toContain('No ingredients listed.');
  });
});

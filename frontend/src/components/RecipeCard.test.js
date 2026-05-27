import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { RecipeCard } from './RecipeCard.js';

const recipe = {
  id: 'r1',
  name: 'Mojito',
  ingredients: [
    { name: 'rum', quantity: '50', unit: 'ml' },
    { name: 'mint', quantity: '10', unit: 'leaves' },
  ],
};

describe('RecipeCard - visual redesign', () => {
  it('card wrapper has rounded-2xl class', () => {
    const el = RecipeCard({ recipe });
    expect(el.className).toContain('rounded-2xl');
  });

  it('card has amber left border accent class border-l-amber-400', () => {
    const el = RecipeCard({ recipe });
    expect(el.className).toContain('border-l-amber-400');
  });
});

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
  beforeEach(() => {
    Object.defineProperty(window, 'scrollX', { value: 0, writable: true, configurable: true });
    Object.defineProperty(window, 'scrollY', { value: 0, writable: true, configurable: true });
  });

  afterEach(() => {
    document.body.querySelector('[data-popover]')?.remove();
  });

  it('shows popover with ingredient names on mouseenter', () => {
    const el = RecipeCard({ recipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover).not.toBeNull();
    expect(popover.textContent).toContain('rum');
    expect(popover.textContent).toContain('mint');
  });

  it('hides popover on mouseleave', () => {
    const el = RecipeCard({ recipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    el.dispatchEvent(new MouseEvent('mouseleave'));
    expect(document.body.querySelector('[data-popover]')).toBeNull();
  });

  it('removes first tile popover when mouse leaves it and enters a second tile', () => {
    const el1 = RecipeCard({ recipe });
    const el2 = RecipeCard({ recipe });
    el1.dispatchEvent(new MouseEvent('mouseenter'));
    el1.dispatchEvent(new MouseEvent('mouseleave'));
    el2.dispatchEvent(new MouseEvent('mouseenter'));
    expect(el1.querySelector('[data-popover]')).toBeNull();
    expect(document.body.querySelector('[data-popover]')).not.toBeNull();
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
  beforeEach(() => {
    Object.defineProperty(window, 'scrollX', { value: 0, writable: true, configurable: true });
    Object.defineProperty(window, 'scrollY', { value: 0, writable: true, configurable: true });
  });

  afterEach(() => {
    document.body.querySelector('[data-popover]')?.remove();
  });

  it('shows all ingredients when recipe has exactly 5', () => {
    const r = { id: 'r2', name: 'T', ingredients: [
      { name: 'a' }, { name: 'b' }, { name: 'c' }, { name: 'd' }, { name: 'e' },
    ]};
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
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
    const popover = document.body.querySelector('[data-popover]');
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
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.textContent).not.toContain('f');
    expect(popover.textContent).toContain('…');
  });
});

describe('RecipeCard - empty ingredients', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'scrollX', { value: 0, writable: true, configurable: true });
    Object.defineProperty(window, 'scrollY', { value: 0, writable: true, configurable: true });
  });

  afterEach(() => {
    document.body.querySelector('[data-popover]')?.remove();
  });

  it('shows "No ingredients listed." when recipe has no ingredients', () => {
    const r = { id: 'r5', name: 'Empty', ingredients: [] };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover).not.toBeNull();
    expect(popover.textContent).toContain('No ingredients listed.');
  });
});

describe('RecipeCard - base spirit popover highlight (T009)', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'scrollX', { value: 0, writable: true, configurable: true });
    Object.defineProperty(window, 'scrollY', { value: 0, writable: true, configurable: true });
  });

  afterEach(() => {
    document.body.querySelector('[data-popover]')?.remove();
  });

  it('base spirit ingredient li contains (base spirit) label', () => {
    const r = {
      id: 'r1',
      name: 'Manhattan',
      ingredients: [
        { name: 'Rye', quantity: '60', unit: 'ml', is_base_spirit: true },
        { name: 'Vermouth', quantity: '30', unit: 'ml' },
      ],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.textContent).toContain('(base spirit)');
  });

  it('base spirit ingredient li has font-semibold class on name span', () => {
    const r = {
      id: 'r1',
      name: 'Manhattan',
      ingredients: [
        { name: 'Rye', quantity: '60', unit: 'ml', is_base_spirit: true },
        { name: 'Vermouth', quantity: '30', unit: 'ml' },
      ],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    const boldSpan = popover.querySelector('.font-semibold');
    expect(boldSpan).not.toBeNull();
    expect(boldSpan.textContent).toBe('Rye');
  });

  it('(base spirit) label appears only on the flagged ingredient, not on neighbours', () => {
    const r = {
      id: 'r1',
      name: 'Manhattan',
      ingredients: [
        { name: 'Rye', quantity: '60', unit: 'ml', is_base_spirit: true },
        { name: 'Vermouth', quantity: '30', unit: 'ml' },
        { name: 'Bitters', quantity: '2', unit: 'dashes' },
      ],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    const items = [...popover.querySelectorAll('li')];
    const labelCount = items.filter(li => li.textContent.includes('(base spirit)')).length;
    expect(labelCount).toBe(1);
  });

  it('all items are rendered identically when no ingredient has is_base_spirit', () => {
    const r = {
      id: 'r1',
      name: 'Mojito',
      ingredients: [
        { name: 'Rum', quantity: '50', unit: 'ml' },
        { name: 'Mint', quantity: '10', unit: 'leaves' },
      ],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.textContent).not.toContain('(base spirit)');
    expect(popover.querySelector('.font-semibold')).toBeNull();
  });
});

describe('RecipeCard - edit/delete controls (T005)', () => {
  const ownerRecipe = { id: 'r1', name: 'Mojito', creator_id: 'user-1', ingredients: [] };

  it('hides edit/delete when currentUser is not provided', () => {
    const el = RecipeCard({ recipe: ownerRecipe });
    expect(el.querySelector('a[href*="edit"]')).toBeNull();
    expect(el.querySelector('button')).toBeNull();
  });

  it('hides edit/delete when currentUser is null', () => {
    const el = RecipeCard({ recipe: ownerRecipe, currentUser: null });
    expect(el.querySelector('a[href*="edit"]')).toBeNull();
    expect(el.querySelector('button')).toBeNull();
  });

  it('hides edit/delete when currentUser is not the owner and not admin', () => {
    const el = RecipeCard({ recipe: ownerRecipe, currentUser: { id: 'user-2', isAdmin: false } });
    expect(el.querySelector('a[href*="edit"]')).toBeNull();
    expect(el.querySelector('button')).toBeNull();
  });

  it('shows edit/delete when currentUser is the recipe owner', () => {
    const el = RecipeCard({ recipe: ownerRecipe, currentUser: { id: 'user-1', isAdmin: false } });
    expect(el.querySelector('a[href*="edit"]')).not.toBeNull();
    expect(el.querySelector('button')).not.toBeNull();
  });

  it('shows edit/delete when currentUser is admin regardless of ownership', () => {
    const el = RecipeCard({ recipe: ownerRecipe, currentUser: { id: 'other-user', isAdmin: true } });
    expect(el.querySelector('a[href*="edit"]')).not.toBeNull();
    expect(el.querySelector('button')).not.toBeNull();
  });

  it('hides edit/delete for legacy recipe (empty creator_id) when non-owner non-admin', () => {
    const legacy = { id: 'r-old', name: 'Old', creator_id: '', ingredients: [] };
    const el = RecipeCard({ recipe: legacy, currentUser: { id: 'user-1', isAdmin: false } });
    expect(el.querySelector('a[href*="edit"]')).toBeNull();
    expect(el.querySelector('button')).toBeNull();
  });

  it('shows edit/delete for legacy recipe when admin', () => {
    const legacy = { id: 'r-old', name: 'Old', creator_id: '', ingredients: [] };
    const el = RecipeCard({ recipe: legacy, currentUser: { id: 'admin-1', isAdmin: true } });
    expect(el.querySelector('a[href*="edit"]')).not.toBeNull();
    expect(el.querySelector('button')).not.toBeNull();
  });
});

describe('RecipeCard - popup overlay (body portal)', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'scrollX', { value: 0, writable: true, configurable: true });
    Object.defineProperty(window, 'scrollY', { value: 0, writable: true, configurable: true });
  });

  afterEach(() => {
    document.body.querySelector('[data-popover]')?.remove();
  });

  const overlayRecipe = {
    id: 'r-overlay',
    name: 'Negroni',
    ingredients: [{ name: 'Gin' }, { name: 'Campari' }],
  };

  it('on mouseenter, popover is appended to document.body', () => {
    const el = RecipeCard({ recipe: overlayRecipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    expect(document.body.querySelector('[data-popover]')).not.toBeNull();
  });

  it('on mouseenter, popover is NOT a descendant of the card element', () => {
    const el = RecipeCard({ recipe: overlayRecipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    expect(el.querySelector('[data-popover]')).toBeNull();
  });

  it('on mouseleave, popover is removed from document.body', () => {
    const el = RecipeCard({ recipe: overlayRecipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    el.dispatchEvent(new MouseEvent('mouseleave'));
    expect(document.body.querySelector('[data-popover]')).toBeNull();
  });

  it('only one popover in document.body when two cards hovered in sequence', () => {
    const el1 = RecipeCard({ recipe: overlayRecipe });
    const el2 = RecipeCard({ recipe: overlayRecipe });
    el1.dispatchEvent(new MouseEvent('mouseenter'));
    el1.dispatchEvent(new MouseEvent('mouseleave'));
    el2.dispatchEvent(new MouseEvent('mouseenter'));
    const popovers = document.body.querySelectorAll('[data-popover]');
    expect(popovers.length).toBe(1);
  });

  it('popover has position: fixed style', () => {
    const el = RecipeCard({ recipe: overlayRecipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.style.position).toBe('fixed');
  });

  it('clicking document.body removes the popover (click-elsewhere closure)', () => {
    const el = RecipeCard({ recipe: overlayRecipe });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    expect(document.body.querySelector('[data-popover]')).not.toBeNull();
    document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    expect(document.body.querySelector('[data-popover]')).toBeNull();
  });
});

// T015: RecipeCard garnish popover (US3)
describe('RecipeCard - garnish popover (T015)', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'scrollX', { value: 0, writable: true, configurable: true });
    Object.defineProperty(window, 'scrollY', { value: 0, writable: true, configurable: true });
  });

  afterEach(() => {
    document.body.querySelector('[data-popover]')?.remove();
  });

  it('shows garnishes in <em> in popover when ingredients.length < MAX_VISIBLE (5)', () => {
    const r = {
      id: 'r1',
      name: 'Old Fashioned',
      ingredients: [
        { name: 'bourbon', quantity: '60', unit: 'ml' },
        { name: 'sugar', quantity: '5', unit: 'ml' },
      ],
      garnishes: ['Express orange oil over the cocktail', 'Use orange peel to garnish'],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    const ems = [...popover.querySelectorAll('em')];
    expect(ems.length).toBe(2);
    expect(ems[0].textContent).toBe('Express orange oil over the cocktail');
  });

  it('does not show garnishes when ingredients.length >= MAX_VISIBLE (5)', () => {
    const r = {
      id: 'r1',
      name: 'Complex Cocktail',
      ingredients: [
        { name: 'a' }, { name: 'b' }, { name: 'c' }, { name: 'd' }, { name: 'e' },
      ],
      garnishes: ['Express orange oil over the cocktail'],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.querySelector('em')).toBeNull();
  });

  it('does not show garnishes when recipe has no garnishes', () => {
    const r = {
      id: 'r1',
      name: 'Simple Cocktail',
      ingredients: [{ name: 'vodka' }, { name: 'soda' }],
      garnishes: [],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.querySelector('em')).toBeNull();
  });

  it('does not show garnishes when recipe.garnishes is absent', () => {
    const r = {
      id: 'r1',
      name: 'Legacy Cocktail',
      ingredients: [{ name: 'rum' }, { name: 'lime' }],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.querySelector('em')).toBeNull();
  });

  it('ingredient list is not truncated (no ellipsis) when garnishes are shown', () => {
    const r = {
      id: 'r1',
      name: 'Small Recipe',
      ingredients: [{ name: 'bourbon' }, { name: 'ice' }],
      garnishes: ['Express orange oil'],
    };
    const el = RecipeCard({ recipe: r });
    el.dispatchEvent(new MouseEvent('mouseenter'));
    const popover = document.body.querySelector('[data-popover]');
    expect(popover.textContent).not.toContain('…');
  });
});

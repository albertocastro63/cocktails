import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRandomRecipe: vi.fn(),
}));

import { getRandomRecipe } from '../api/client.js';
import { Home } from './Home.js';

describe('Home page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('renders a loading state initially', () => {
    getRandomRecipe.mockReturnValue(new Promise(() => {}));
    const el = Home();
    document.body.appendChild(el);
    expect(document.body.textContent.toLowerCase()).toMatch(/loading/);
  });

  it('renders recipe name when recipe returns', async () => {
    getRandomRecipe.mockResolvedValue({ id: 'r1', name: 'Mojito', ingredients: [], steps: [], properties: {}, notes: '' });
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });
  });

  it('renders ingredients when recipe has ingredients', async () => {
    getRandomRecipe.mockResolvedValue({
      id: 'r1',
      name: 'Mojito',
      ingredients: [{ name: 'Rum', quantity: '50', unit: 'ml' }],
      steps: [],
      properties: {},
      notes: '',
    });
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Rum');
    });
  });

  it('renders steps when recipe has steps', async () => {
    getRandomRecipe.mockResolvedValue({
      id: 'r1',
      name: 'Mojito',
      ingredients: [],
      steps: ['Muddle mint leaves'],
      properties: {},
      notes: '',
    });
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Muddle mint leaves');
    });
  });

  it('renders notes as markdown HTML when recipe has notes', async () => {
    getRandomRecipe.mockResolvedValue({
      id: 'r1',
      name: 'Mojito',
      ingredients: [],
      steps: [],
      properties: {},
      notes: '**bold notes**',
    });
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('strong')).not.toBeNull();
    });
  });

  it('renders EmptyState when null (204)', async () => {
    getRandomRecipe.mockResolvedValue(null);
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/no recipe|empty|nothing|add one/);
    });
  });

  it('renders hero band with from-stone-900 gradient class', () => {
    getRandomRecipe.mockReturnValue(new Promise(() => {}));
    const el = Home();
    document.body.appendChild(el);
    const hero = el.querySelector('.from-stone-900');
    expect(hero).not.toBeNull();
  });

  it('notes container has prose prose-stone max-w-none overflow-x-auto classes for typography styling', async () => {
    getRandomRecipe.mockResolvedValue({
      id: 'r1',
      name: 'Mojito',
      ingredients: [],
      steps: [],
      properties: {},
      notes: '## Tips\n**Shake well.**',
    });
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const strong = document.body.querySelector('strong');
      expect(strong).not.toBeNull();
      const notesContainer = strong.closest('div');
      expect(notesContainer.className).toContain('prose');
      expect(notesContainer.className).toContain('prose-stone');
      expect(notesContainer.className).toContain('max-w-none');
      expect(notesContainer.className).toContain('overflow-x-auto');
    });
  });
});

describe('compact landing header on mobile (responsive classes)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
    getRandomRecipe.mockReturnValue(new Promise(() => {})); // keep hero rendered
  });

  it('hero wrapper is compact on phones and full at md+ (H1, SC-001/SC-003)', () => {
    const hero = Home().querySelector('.from-stone-900');
    expect(hero.className).toContain('py-4');
    expect(hero.className).toContain('md:py-16');
  });

  it('title shrinks on phones and restores at md+, text unchanged (H2/H4)', () => {
    const h1 = Home().querySelector('h1');
    expect(h1.className).toContain('text-xl');
    expect(h1.className).toContain('md:text-4xl');
    expect(h1.textContent).toBe('Cocktail Recipes');
  });

  it('subtitle shrinks on phones and restores at md+, token + text unchanged (H3/H5/H7)', () => {
    const sub = Home().querySelector('.from-stone-900 p');
    expect(sub.className).toContain('text-sm');
    expect(sub.className).toContain('md:text-lg');
    expect(sub.className).toContain('text-amber-400');
    expect(sub.textContent).toBe('Discover your next favorite drink');
  });

  it('lays out text + CTA in a row on phones, stacked at md+ (H8)', () => {
    const heroInner = Home().querySelector('.from-stone-900 > div');
    expect(heroInner.className).toContain('flex');
    expect(heroInner.className).toContain('md:block');
  });

  it('CTA keeps href/text but is smaller on phones and restored at md+ (H6/H9)', () => {
    const cta = Home().querySelector('a[href="#/recipes"]');
    expect(cta).not.toBeNull();
    expect(cta.textContent).toBe('All Recipes');
    // smaller padding + font on phones
    expect(cta.className).toContain('px-4');
    expect(cta.className).toContain('py-1.5');
    expect(cta.className).toContain('text-sm');
    // restored to today's size at md+
    expect(cta.className).toContain('md:px-6');
    expect(cta.className).toContain('md:py-2');
    expect(cta.className).toContain('md:text-base');
  });

  // FR-011: the random featured cocktail must never show a related-cocktails list.
  it('does not render a Related cocktails section for the random cocktail', async () => {
    getRandomRecipe.mockResolvedValue({
      id: 'r1', name: 'Mojito', ingredients: [], steps: [], properties: {},
      related: [{ id: 'lh', name: 'Left Hand' }], // even if present, Home must ignore it
    });
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => expect(document.body.textContent).toContain('Mojito'));
    expect(document.body.textContent).not.toContain('Related cocktails');
  });
});

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
});

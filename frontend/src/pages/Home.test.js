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

  it('renders RecipeCard when recipe returns', async () => {
    getRandomRecipe.mockResolvedValue({ id: 'r1', name: 'Mojito', ingredients: [] });
    const el = Home();
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
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
});

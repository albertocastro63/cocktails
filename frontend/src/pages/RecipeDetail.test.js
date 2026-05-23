import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipe: vi.fn(),
  getFavoriteStatus: vi.fn(),
  favoriteRecipe: vi.fn(),
  unfavoriteRecipe: vi.fn(),
}));

vi.mock('../api/auth.js', () => ({
  getToken: vi.fn(() => 'test-token'),
  getUserID: vi.fn(() => null),
  isAdmin: vi.fn(() => false),
}));

import { getRecipe, getFavoriteStatus, favoriteRecipe, unfavoriteRecipe } from '../api/client.js';
import { getUserID } from '../api/auth.js';
import { RecipeDetail } from './RecipeDetail.js';

const fullRecipe = {
  id: 'r1',
  name: 'Mojito',
  ingredients: [{ name: 'rum', quantity: '50', unit: 'ml' }],
  steps: ['Muddle mint', 'Add rum', 'Top with soda'],
  properties: { style: 'refreshing', base_spirit: 'rum' },
  creator_id: 'u1',
};

describe('RecipeDetail page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
  });
  afterEach(() => { document.body.innerHTML = ''; });

  it('renders loading state initially', () => {
    getRecipe.mockReturnValue(new Promise(() => {}));
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    expect(document.body.textContent.toLowerCase()).toMatch(/loading/);
  });

  it('renders recipe name and ingredients', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
      expect(document.body.textContent).toContain('rum');
    });
  });

  it('renders steps', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Muddle mint');
    });
  });

  it('renders all property pairs', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('style');
      expect(document.body.textContent).toContain('refreshing');
    });
  });

  it('renders notes as markdown HTML when recipe has notes', async () => {
    getRecipe.mockResolvedValue({ ...fullRecipe, notes: '**bold notes**' });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('strong')).not.toBeNull();
    });
  });

  it('does not render a notes section when notes is empty', async () => {
    getRecipe.mockResolvedValue({ ...fullRecipe, notes: '' });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('h2')).not.toBeNull();
      const h2s = [...document.body.querySelectorAll('h2')].map(h => h.textContent);
      expect(h2s).not.toContain('Notes');
    });
  });

  it('renders error state on failure', async () => {
    getRecipe.mockRejectedValue(new Error('not found'));
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent.toLowerCase()).toMatch(/error|not found|failed/);
    });
  });

  it('recipe title h1 has text-stone-900 class', async () => {
    getRecipe.mockResolvedValue(fullRecipe);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const h1 = document.body.querySelector('h1');
      expect(h1).not.toBeNull();
      expect(h1.className).toContain('text-stone-900');
    });
  });

  it('notes container has prose prose-stone max-w-none overflow-x-auto classes for typography styling', async () => {
    getRecipe.mockResolvedValue({ ...fullRecipe, notes: '## Tips\n**Shake well.**' });
    const el = RecipeDetail({ id: 'r1' });
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

describe('RecipeDetail - favorites', () => {
  const recipeByOther = {
    id: 'r1',
    name: 'Mojito',
    ingredients: [],
    steps: [],
    properties: {},
    creator_id: 'u-other',
  };

  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
    getRecipe.mockResolvedValue(recipeByOther);
    getFavoriteStatus.mockResolvedValue({ is_favorite: false });
    favoriteRecipe.mockResolvedValue(null);
    unfavoriteRecipe.mockResolvedValue(null);
  });

  afterEach(() => { document.body.innerHTML = ''; });

  // (1) logged in and recipe.creator_id !== getUserID() → button rendered
  it('renders Add to favorites button when logged in and recipe is not own', async () => {
    getUserID.mockReturnValue('u1');
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const btn = document.body.querySelector('button[aria-label="Add to favorites"]');
      expect(btn).not.toBeNull();
    });
  });

  // (2) getFavoriteStatus returns {is_favorite:true} → aria-pressed="true"
  it('renders button with aria-pressed="true" when recipe is already favorited', async () => {
    getUserID.mockReturnValue('u1');
    getFavoriteStatus.mockResolvedValue({ is_favorite: true });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      const btn = document.body.querySelector('button[aria-pressed="true"]');
      expect(btn).not.toBeNull();
    });
  });

  // (3) recipe.creator_id === getUserID() → no favorite button
  it('does not render a favorite button when recipe is own', async () => {
    getUserID.mockReturnValue('u-other');
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });
    expect(document.body.querySelector('button[aria-label="Add to favorites"]')).toBeNull();
    expect(document.body.querySelector('button[aria-label="Remove from favorites"]')).toBeNull();
  });

  // (4) not logged in → no favorite button
  it('does not render a favorite button when user is not logged in', async () => {
    getUserID.mockReturnValue(null);
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.textContent).toContain('Mojito');
    });
    expect(document.body.querySelector('button[aria-label="Add to favorites"]')).toBeNull();
  });

  // (5) clicking when is_favorite=false calls favoriteRecipe(id, token)
  it('calls favoriteRecipe when clicking the unfavorited button', async () => {
    getUserID.mockReturnValue('u1');
    getFavoriteStatus.mockResolvedValue({ is_favorite: false });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('button[aria-label="Add to favorites"]')).not.toBeNull();
    });
    document.body.querySelector('button[aria-label="Add to favorites"]').click();
    expect(favoriteRecipe).toHaveBeenCalledWith('r1', 'test-token');
  });

  // (6) when is_favorite=true and user clicks, unfavoriteRecipe(id, token) is called
  it('calls unfavoriteRecipe when clicking the favorited button', async () => {
    getUserID.mockReturnValue('u1');
    getFavoriteStatus.mockResolvedValue({ is_favorite: true });
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('button[aria-label="Remove from favorites"]')).not.toBeNull();
    });
    document.body.querySelector('button[aria-label="Remove from favorites"]').click();
    expect(unfavoriteRecipe).toHaveBeenCalledWith('r1', 'test-token');
  });

  // (7) optimistic update: mock favoriteRecipe as never-resolving; after click, aria-pressed="true" set immediately
  it('optimistically sets aria-pressed="true" before API resolves', async () => {
    getUserID.mockReturnValue('u1');
    getFavoriteStatus.mockResolvedValue({ is_favorite: false });
    favoriteRecipe.mockReturnValue(new Promise(() => {}));
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('button[aria-label="Add to favorites"]')).not.toBeNull();
    });
    document.body.querySelector('button[aria-label="Add to favorites"]').click();
    const btn = document.body.querySelector('button[aria-pressed="true"]');
    expect(btn).not.toBeNull();
  });

  // (8) when favoriteRecipe rejects, aria-pressed reverts to "false"
  it('reverts aria-pressed to false when favoriteRecipe rejects', async () => {
    getUserID.mockReturnValue('u1');
    getFavoriteStatus.mockResolvedValue({ is_favorite: false });
    favoriteRecipe.mockRejectedValue(new Error('Network error'));
    const el = RecipeDetail({ id: 'r1' });
    document.body.appendChild(el);
    await vi.waitFor(() => {
      expect(document.body.querySelector('button[aria-label="Add to favorites"]')).not.toBeNull();
    });
    document.body.querySelector('button[aria-label="Add to favorites"]').click();
    await vi.waitFor(() => {
      const btn = document.body.querySelector('button[aria-pressed="false"]');
      expect(btn).not.toBeNull();
    });
  });
});

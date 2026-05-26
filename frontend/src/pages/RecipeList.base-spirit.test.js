import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('../api/client.js', () => ({
  getRecipes: vi.fn(),
}));

import { getRecipes } from '../api/client.js';
import { parseBaseSpirit, normaliseWhisky, RecipeList } from './RecipeList.js';

describe('parseBaseSpirit', () => {
  it('extracts base spirit from "base spirit is gin"', () => {
    expect(parseBaseSpirit('base spirit is gin')).toEqual({ baseSpirit: 'gin', q: '' });
  });

  it('handles equals syntax "base spirit = gin"', () => {
    expect(parseBaseSpirit('base spirit = gin')).toEqual({ baseSpirit: 'gin', q: '' });
  });

  it('is case-insensitive on the keyword', () => {
    const result = parseBaseSpirit('BASE SPIRIT IS GIN');
    expect(result.baseSpirit).toBe('GIN');
    expect(result.q).toBe('');
  });

  it('leaves remaining ingredient term in q', () => {
    expect(parseBaseSpirit('absinthe base spirit is rye whiskey')).toEqual({
      baseSpirit: 'rye whiskey',
      q: 'absinthe',
    });
  });

  it('treats trailing-space value as empty baseSpirit', () => {
    const result = parseBaseSpirit('base spirit is ');
    expect(result.baseSpirit).toBe('');
    expect(result.q).toBe('');
  });

  it('uses first clause and strips all base-spirit clauses from q', () => {
    expect(parseBaseSpirit('base spirit is gin base spirit is rum')).toEqual({
      baseSpirit: 'gin',
      q: '',
    });
  });

  it('passes through plain ingredient search untouched', () => {
    expect(parseBaseSpirit('martini')).toEqual({ baseSpirit: '', q: 'martini' });
  });

  it('handles empty string', () => {
    expect(parseBaseSpirit('')).toEqual({ baseSpirit: '', q: '' });
  });
});

describe('normaliseWhisky', () => {
  it('replaces whisky with whiskey', () => {
    expect(normaliseWhisky('rye whisky')).toBe('rye whiskey');
  });

  it('replaces whisky in a longer phrase', () => {
    expect(normaliseWhisky('scotch whisky and lime')).toBe('scotch whiskey and lime');
  });

  it('preserves capitalisation', () => {
    expect(normaliseWhisky('Rye Whisky')).toBe('Rye Whiskey');
  });

  it('leaves already-normalised whiskey unchanged', () => {
    expect(normaliseWhisky('rye whiskey')).toBe('rye whiskey');
  });

  it('does not affect non-whisky terms', () => {
    expect(normaliseWhisky('gin')).toBe('gin');
  });

  it('does not replace whisky inside a compound word', () => {
    expect(normaliseWhisky('whiskysoaked')).toBe('whiskysoaked');
  });
});

describe('RecipeList base-spirit integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '';
    getRecipes.mockResolvedValue({ data: [], total: 0, page: 1, limit: 20 });
  });
  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('sends base_spirit param when search contains base spirit clause', async () => {
    vi.useFakeTimers();
    const el = RecipeList();
    document.body.appendChild(el);
    await Promise.resolve();

    const input = document.body.querySelector('input');
    input.value = 'base spirit is gin';
    input.dispatchEvent(new Event('input'));
    vi.advanceTimersByTime(350);

    expect(getRecipes).toHaveBeenCalledWith(
      expect.objectContaining({ base_spirit: 'gin' })
    );
    const call = getRecipes.mock.calls[getRecipes.mock.calls.length - 1][0];
    expect(call.q === undefined || call.q === '').toBe(true);
    vi.useRealTimers();
  });

  it('sends normalised whiskey form when user types whisky', async () => {
    vi.useFakeTimers();
    const el = RecipeList();
    document.body.appendChild(el);
    await Promise.resolve();

    const input = document.body.querySelector('input');
    input.value = 'rye whisky';
    input.dispatchEvent(new Event('input'));
    vi.advanceTimersByTime(350);

    expect(getRecipes).toHaveBeenCalledWith(
      expect.objectContaining({ q: 'rye whiskey' })
    );
    vi.useRealTimers();
  });

  it('normalises whisky in base_spirit value too', async () => {
    vi.useFakeTimers();
    const el = RecipeList();
    document.body.appendChild(el);
    await Promise.resolve();

    const input = document.body.querySelector('input');
    input.value = 'base spirit is rye whisky';
    input.dispatchEvent(new Event('input'));
    vi.advanceTimersByTime(350);

    expect(getRecipes).toHaveBeenCalledWith(
      expect.objectContaining({ base_spirit: 'rye whiskey' })
    );
    vi.useRealTimers();
  });
});

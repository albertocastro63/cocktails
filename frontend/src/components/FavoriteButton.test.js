import { describe, it, expect, vi, beforeEach } from 'vitest';
import { FavoriteButton } from './FavoriteButton.js';

describe('FavoriteButton', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });

  // (1) renders a <button> element
  it('renders a <button> element', () => {
    const el = FavoriteButton({ isFavorited: false, onToggle: vi.fn() });
    expect(el.tagName.toLowerCase()).toBe('button');
  });

  // (2) when isFavorited=false, aria-label is "Add to favorites" and aria-pressed="false"
  it('when isFavorited=false, aria-label is "Add to favorites" and aria-pressed="false"', () => {
    const el = FavoriteButton({ isFavorited: false, onToggle: vi.fn() });
    expect(el.getAttribute('aria-label')).toBe('Add to favorites');
    expect(el.getAttribute('aria-pressed')).toBe('false');
  });

  // (3) when isFavorited=true, aria-label is "Remove from favorites" and aria-pressed="true"
  it('when isFavorited=true, aria-label is "Remove from favorites" and aria-pressed="true"', () => {
    const el = FavoriteButton({ isFavorited: true, onToggle: vi.fn() });
    expect(el.getAttribute('aria-label')).toBe('Remove from favorites');
    expect(el.getAttribute('aria-pressed')).toBe('true');
  });

  // (4) when isFavorited=true, button has class text-red-500
  it('when isFavorited=true, button has class text-red-500', () => {
    const el = FavoriteButton({ isFavorited: true, onToggle: vi.fn() });
    expect(el.className).toContain('text-red-500');
  });

  // (5) when isFavorited=false, button has class text-stone-400
  it('when isFavorited=false, button has class text-stone-400', () => {
    const el = FavoriteButton({ isFavorited: false, onToggle: vi.fn() });
    expect(el.className).toContain('text-stone-400');
  });

  // (6) clicking the button calls onToggle() exactly once
  it('clicking the button calls onToggle exactly once', () => {
    const onToggle = vi.fn();
    const el = FavoriteButton({ isFavorited: false, onToggle });
    document.body.appendChild(el);
    el.click();
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  // (7) button has class focus-visible:outline-amber-500
  it('button has class focus-visible:outline-amber-500', () => {
    const el = FavoriteButton({ isFavorited: false, onToggle: vi.fn() });
    expect(el.className).toContain('focus-visible:outline-amber-500');
  });
});

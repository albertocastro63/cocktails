import { describe, it, expect } from 'vitest';
import { EmptyState } from './EmptyState.js';

describe('EmptyState', () => {
  it('renders the message prop', () => {
    const el = EmptyState({ message: 'No recipes found' });
    expect(el.textContent).toContain('No recipes found');
  });

  it('renders a default message when none provided', () => {
    const el = EmptyState({});
    expect(el.textContent.trim().length).toBeGreaterThan(0);
  });

  it('uses text-stone-500 class not legacy text-gray-400', () => {
    const el = EmptyState({ message: 'Nothing here' });
    expect(el.className).toContain('text-stone-500');
  });
});

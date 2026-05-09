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
});

import { describe, it, expect } from 'vitest';
import { navDestinations, partitionNav, isActive } from './destinations.js';

describe('navDestinations — per-state sets (U1-U3)', () => {
  const ids = (state) => navDestinations(state).map((d) => d.id);

  it('visitor sees All Recipes and Sign In', () => {
    expect(ids('visitor')).toEqual(['all-recipes', 'sign-in']);
  });

  it('regular user sees browse/create + sign out', () => {
    expect(ids('user')).toEqual(['all-recipes', 'my-recipes', 'new-recipe', 'sign-out']);
  });

  it('admin adds Users and admin Recipes', () => {
    const set = ids('admin');
    expect(set).toContain('users');
    expect(set).toContain('admin-recipes');
    expect(set).toEqual(['all-recipes', 'my-recipes', 'new-recipe', 'users', 'admin-recipes', 'sign-out']);
  });

  it('every destination carries a label and an icon', () => {
    for (const d of navDestinations('admin')) {
      expect(d.label, `${d.id} label`).toBeTruthy();
      expect(d.icon, `${d.id} icon`).toMatch(/<svg/);
    }
  });
});

describe('partitionNav — bounded direct items + overflow (U4-U6)', () => {
  it('keeps all items direct when count <= maxSlots, no More', () => {
    const { direct, overflow } = partitionNav(navDestinations('user'), 5);
    expect(direct).toHaveLength(4);
    expect(overflow).toHaveLength(0);
  });

  it('admin (6) splits into 4 direct + 2 overflow', () => {
    const { direct, overflow } = partitionNav(navDestinations('admin'), 5);
    expect(direct.map((d) => d.id)).toEqual(['all-recipes', 'my-recipes', 'new-recipe', 'users']);
    expect(overflow.map((d) => d.id)).toEqual(['admin-recipes', 'sign-out']);
    // bar never exceeds maxSlots (direct + the synthesized More slot)
    expect(direct.length + (overflow.length ? 1 : 0)).toBeLessThanOrEqual(5);
  });

  it('growth stays bounded: 7 items -> 4 direct + 3 overflow (SC-005)', () => {
    const grown = [
      ...navDestinations('admin'),
      { id: 'future', label: 'Future', icon: '<svg></svg>', href: '#/future', priority: 7, match: '/future' },
    ];
    const { direct, overflow } = partitionNav(grown, 5);
    expect(direct).toHaveLength(4);
    expect(overflow).toHaveLength(3);
  });
});

describe('isActive — active-state matcher (U7)', () => {
  const dest = (id) => navDestinations('admin').find((d) => d.id === id);

  it('matches exact route', () => {
    expect(isActive(dest('all-recipes'), '/recipes')).toBe(true);
    expect(isActive(dest('new-recipe'), '/recipes/new')).toBe(true);
  });

  it('does not activate All Recipes on the New Recipe route', () => {
    expect(isActive(dest('all-recipes'), '/recipes/new')).toBe(false);
  });

  it('action items (sign-out) are never active', () => {
    expect(isActive(dest('sign-out'), '/anything')).toBe(false);
  });
});

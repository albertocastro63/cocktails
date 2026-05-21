import { describe, it, expect, vi } from 'vitest';
import { SortButtonGroup } from './SortButtonGroup.js';

describe('SortButtonGroup', () => {
  it('renders a button labeled "A→Z" and a button labeled "Z→A"', () => {
    const el = SortButtonGroup({ currentDir: null, onSort: vi.fn() });
    const buttons = el.querySelectorAll('button');
    expect(buttons).toHaveLength(2);
    expect(buttons[0].textContent).toBe('A→Z');
    expect(buttons[1].textContent).toBe('Z→A');
  });

  it('both buttons have aria-pressed="false" when currentDir is null', () => {
    const el = SortButtonGroup({ currentDir: null, onSort: vi.fn() });
    const buttons = el.querySelectorAll('button');
    expect(buttons[0].getAttribute('aria-pressed')).toBe('false');
    expect(buttons[1].getAttribute('aria-pressed')).toBe('false');
  });

  it('A→Z button has aria-pressed="true" when currentDir is "asc"', () => {
    const el = SortButtonGroup({ currentDir: 'asc', onSort: vi.fn() });
    const buttons = el.querySelectorAll('button');
    expect(buttons[0].getAttribute('aria-pressed')).toBe('true');
    expect(buttons[1].getAttribute('aria-pressed')).toBe('false');
  });

  it('Z→A button has aria-pressed="true" when currentDir is "desc"', () => {
    const el = SortButtonGroup({ currentDir: 'desc', onSort: vi.fn() });
    const buttons = el.querySelectorAll('button');
    expect(buttons[0].getAttribute('aria-pressed')).toBe('false');
    expect(buttons[1].getAttribute('aria-pressed')).toBe('true');
  });

  it('clicking A→Z button calls onSort("asc")', () => {
    const onSort = vi.fn();
    const el = SortButtonGroup({ currentDir: null, onSort });
    el.querySelectorAll('button')[0].click();
    expect(onSort).toHaveBeenCalledWith('asc');
  });

  it('clicking Z→A button calls onSort("desc")', () => {
    const onSort = vi.fn();
    const el = SortButtonGroup({ currentDir: null, onSort });
    el.querySelectorAll('button')[1].click();
    expect(onSort).toHaveBeenCalledWith('desc');
  });

  it('active button has class bg-amber-100', () => {
    const el = SortButtonGroup({ currentDir: 'asc', onSort: vi.fn() });
    expect(el.querySelectorAll('button')[0].className).toContain('bg-amber-100');
  });

  it('inactive button has class bg-white', () => {
    const el = SortButtonGroup({ currentDir: 'asc', onSort: vi.fn() });
    expect(el.querySelectorAll('button')[1].className).toContain('bg-white');
  });
});

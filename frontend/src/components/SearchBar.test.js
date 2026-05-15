import { describe, it, expect, vi } from 'vitest';
import { SearchBar } from './SearchBar.js';

describe('SearchBar', () => {
  it('renders an input element', () => {
    const el = SearchBar({ onSearch: vi.fn() });
    expect(el.querySelector('input')).not.toBeNull();
  });

  it('calls onSearch with current value after debounce', async () => {
    vi.useFakeTimers();
    const onSearch = vi.fn();
    const el = SearchBar({ onSearch });
    const input = el.querySelector('input');

    input.value = 'mojito';
    input.dispatchEvent(new Event('input'));
    expect(onSearch).not.toHaveBeenCalled();

    vi.advanceTimersByTime(350);
    expect(onSearch).toHaveBeenCalledWith('mojito');
    vi.useRealTimers();
  });

  it('renders a clear button when value is present', () => {
    const el = SearchBar({ onSearch: vi.fn(), value: 'rum' });
    const btn = el.querySelector('button');
    expect(btn).not.toBeNull();
  });

  it('input has focus:ring-amber-400 class', () => {
    const el = SearchBar({ onSearch: vi.fn() });
    const input = el.querySelector('input');
    expect(input.className).toContain('focus:ring-amber-400');
  });
});

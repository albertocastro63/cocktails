import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Pagination } from './Pagination.js';

describe('Pagination component', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });

  it('renders Previous, Next, and a "Page X of Y" indicator', () => {
    const el = Pagination({ page: 2, totalPages: 5, onPage: () => {} });
    document.body.appendChild(el);
    expect(document.body.querySelector('[data-dir="prev"]').textContent).toMatch(/previous/i);
    expect(document.body.querySelector('[data-dir="next"]').textContent).toMatch(/next/i);
    expect(document.body.querySelector('[data-testid="page-indicator"]').textContent).toBe('Page 2 of 5');
  });

  it('disables Previous on the first page', () => {
    const el = Pagination({ page: 1, totalPages: 3, onPage: () => {} });
    document.body.appendChild(el);
    expect(document.body.querySelector('[data-dir="prev"]').disabled).toBe(true);
    expect(document.body.querySelector('[data-dir="next"]').disabled).toBe(false);
  });

  it('disables Next on the last page', () => {
    const el = Pagination({ page: 3, totalPages: 3, onPage: () => {} });
    document.body.appendChild(el);
    expect(document.body.querySelector('[data-dir="next"]').disabled).toBe(true);
    expect(document.body.querySelector('[data-dir="prev"]').disabled).toBe(false);
  });

  it('invokes onPage with the target page when an enabled button is clicked', () => {
    const onPage = vi.fn();
    const el = Pagination({ page: 2, totalPages: 5, onPage });
    document.body.appendChild(el);
    document.body.querySelector('[data-dir="next"]').click();
    document.body.querySelector('[data-dir="prev"]').click();
    expect(onPage).toHaveBeenNthCalledWith(1, 3);
    expect(onPage).toHaveBeenNthCalledWith(2, 1);
  });

  it('does not invoke onPage when a disabled button is clicked', () => {
    const onPage = vi.fn();
    const el = Pagination({ page: 1, totalPages: 2, onPage });
    document.body.appendChild(el);
    document.body.querySelector('[data-dir="prev"]').click();
    expect(onPage).not.toHaveBeenCalled();
  });

  it('renders no controls when there is a single page', () => {
    const el = Pagination({ page: 1, totalPages: 1, onPage: () => {} });
    document.body.appendChild(el);
    expect(document.body.querySelector('button')).toBeNull();
    expect(el.dataset.empty).toBe('true');
  });
});

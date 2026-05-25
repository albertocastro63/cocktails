import { describe, it, expect } from 'vitest';
import { Footer } from './Footer.js';

describe('Footer', () => {
  it('returns a <footer> element', () => {
    const el = Footer();
    expect(el.tagName.toLowerCase()).toBe('footer');
  });

  it('contains an <hr> element', () => {
    const el = Footer();
    const hr = el.querySelector('hr');
    expect(hr).not.toBeNull();
  });

  it('contains copyright text with current year', () => {
    const el = Footer();
    const year = new Date().getFullYear();
    expect(el.textContent).toContain(`© ${year} Cocktails`);
  });

  it('copyright year matches new Date().getFullYear()', () => {
    const el = Footer();
    const year = String(new Date().getFullYear());
    expect(el.textContent).toContain(year);
  });

  it('inner wrapper has max-w-4xl class for content-area constraint', () => {
    const el = Footer();
    const wrapper = el.querySelector('.max-w-4xl');
    expect(wrapper).not.toBeNull();
  });

  it('hr has mt-4 class and mb-0 class (no bottom margin)', () => {
    const el = Footer();
    const hr = el.querySelector('hr');
    expect(hr.className).toContain('mt-4');
    expect(hr.className).toContain('mb-0');
  });
});

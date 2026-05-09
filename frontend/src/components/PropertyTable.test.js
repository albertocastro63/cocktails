import { describe, it, expect } from 'vitest';
import { PropertyTable } from './PropertyTable.js';

describe('PropertyTable', () => {
  it('renders all key-value property pairs', () => {
    const properties = { style: 'tropical', base_spirit: 'rum', garnish: 'lime' };
    const el = PropertyTable({ properties });
    expect(el.textContent).toContain('style');
    expect(el.textContent).toContain('tropical');
    expect(el.textContent).toContain('base_spirit');
    expect(el.textContent).toContain('rum');
    expect(el.textContent).toContain('garnish');
    expect(el.textContent).toContain('lime');
  });

  it('renders nothing meaningful for empty properties', () => {
    const el = PropertyTable({ properties: {} });
    expect(el).not.toBeNull();
  });

  it('renders arbitrary unknown keys (FR-005 flexibility)', () => {
    const properties = { unknown_field: 'some value', another_custom: '42' };
    const el = PropertyTable({ properties });
    expect(el.textContent).toContain('unknown_field');
    expect(el.textContent).toContain('some value');
  });
});

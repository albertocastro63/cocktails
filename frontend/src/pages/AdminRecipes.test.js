import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AdminRecipes } from './AdminRecipes.js';

vi.mock('../api/client.js', () => ({
  downloadRecipeSchema: vi.fn(),
  exportRecipes: vi.fn(),
  importRecipes: vi.fn(),
}));

vi.mock('../api/auth.js', () => ({
  getToken: vi.fn(() => 'test-token'),
}));

beforeEach(() => {
  document.body.innerHTML = '';
  vi.clearAllMocks();
});

// US1: schema download button
it('renders a data-download-schema button', () => {
  const el = AdminRecipes();
  document.body.appendChild(el);
  const btn = document.body.querySelector('[data-download-schema]');
  expect(btn).not.toBeNull();
});

// US2: export button
it('renders a data-export-recipes button', () => {
  const el = AdminRecipes();
  document.body.appendChild(el);
  const btn = document.body.querySelector('[data-export-recipes]');
  expect(btn).not.toBeNull();
});

// US3: import control elements
it('renders a data-import-file input', () => {
  const el = AdminRecipes();
  document.body.appendChild(el);
  expect(document.body.querySelector('[data-import-file]')).not.toBeNull();
});

it('renders a data-import-submit button', () => {
  const el = AdminRecipes();
  document.body.appendChild(el);
  expect(document.body.querySelector('[data-import-submit]')).not.toBeNull();
});

it('renders a data-import-status element', () => {
  const el = AdminRecipes();
  document.body.appendChild(el);
  expect(document.body.querySelector('[data-import-status]')).not.toBeNull();
});

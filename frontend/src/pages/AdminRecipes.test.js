import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AdminRecipes } from './AdminRecipes.js';
import { downloadRecipeSchema, exportRecipes, importRecipes } from '../api/client.js';

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
  global.URL.createObjectURL = vi.fn(() => 'blob:mock-url');
  global.URL.revokeObjectURL = vi.fn();
  global.alert = vi.fn();
});

// Render checks
it('renders a data-download-schema button', () => {
  document.body.appendChild(AdminRecipes());
  expect(document.body.querySelector('[data-download-schema]')).not.toBeNull();
});

it('renders a data-export-recipes button', () => {
  document.body.appendChild(AdminRecipes());
  expect(document.body.querySelector('[data-export-recipes]')).not.toBeNull();
});

it('renders a data-import-file input', () => {
  document.body.appendChild(AdminRecipes());
  expect(document.body.querySelector('[data-import-file]')).not.toBeNull();
});

it('renders a data-import-submit button', () => {
  document.body.appendChild(AdminRecipes());
  expect(document.body.querySelector('[data-import-submit]')).not.toBeNull();
});

it('renders a data-import-status element', () => {
  document.body.appendChild(AdminRecipes());
  expect(document.body.querySelector('[data-import-status]')).not.toBeNull();
});

// Schema download button
it('clicking schema button calls downloadRecipeSchema and triggers a download', async () => {
  const blob = new Blob(['{}'], { type: 'application/json' });
  downloadRecipeSchema.mockResolvedValue(blob);

  document.body.appendChild(AdminRecipes());
  document.body.querySelector('[data-download-schema]').click();

  await vi.waitFor(() => expect(URL.createObjectURL).toHaveBeenCalledWith(blob));
  expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:mock-url');
});

it('clicking schema button shows alert on failure', async () => {
  downloadRecipeSchema.mockRejectedValue(new Error('Network error'));

  document.body.appendChild(AdminRecipes());
  document.body.querySelector('[data-download-schema]').click();

  await vi.waitFor(() => expect(global.alert).toHaveBeenCalledWith('Network error'));
});

it('schema button is re-enabled after a failed download', async () => {
  downloadRecipeSchema.mockRejectedValue(new Error('fail'));

  document.body.appendChild(AdminRecipes());
  const btn = document.body.querySelector('[data-download-schema]');
  btn.click();

  await vi.waitFor(() => expect(global.alert).toHaveBeenCalled());
  expect(btn.disabled).toBe(false);
  expect(btn.textContent).toBe('Download Schema');
});

// Export button
it('clicking export button calls exportRecipes and triggers a download', async () => {
  const blob = new Blob(['[]'], { type: 'application/json' });
  exportRecipes.mockResolvedValue(blob);

  document.body.appendChild(AdminRecipes());
  document.body.querySelector('[data-export-recipes]').click();

  await vi.waitFor(() => expect(URL.createObjectURL).toHaveBeenCalledWith(blob));
});

it('clicking export button shows alert on failure', async () => {
  exportRecipes.mockRejectedValue(new Error('Export failed'));

  document.body.appendChild(AdminRecipes());
  document.body.querySelector('[data-export-recipes]').click();

  await vi.waitFor(() => expect(global.alert).toHaveBeenCalledWith('Export failed'));
});

// Import button
it('shows error when import is clicked with no file selected', () => {
  document.body.appendChild(AdminRecipes());
  document.body.querySelector('[data-import-submit]').click();

  const status = document.body.querySelector('[data-import-status]');
  expect(status.textContent).toBe('Please select a JSON file.');
  expect(status.className).toContain('text-red-600');
});

it('shows error for a file containing invalid JSON', async () => {
  document.body.appendChild(AdminRecipes());

  const fileInput = document.body.querySelector('[data-import-file]');
  const file = new File(['not json!'], 'recipes.json', { type: 'application/json' });
  Object.defineProperty(fileInput, 'files', { value: [file], configurable: true });

  document.body.querySelector('[data-import-submit]').click();

  await vi.waitFor(() => {
    expect(document.body.querySelector('[data-import-status]').textContent)
      .toBe('The selected file is not valid JSON.');
  });
});

it('shows error when JSON file is not an array', async () => {
  document.body.appendChild(AdminRecipes());

  const fileInput = document.body.querySelector('[data-import-file]');
  const file = new File([JSON.stringify({ not: 'array' })], 'recipes.json', { type: 'application/json' });
  Object.defineProperty(fileInput, 'files', { value: [file], configurable: true });

  document.body.querySelector('[data-import-submit]').click();

  await vi.waitFor(() => {
    expect(document.body.querySelector('[data-import-status]').textContent)
      .toBe('The selected file must be a JSON array of recipes.');
  });
});

it('shows success message after a successful import', async () => {
  importRecipes.mockResolvedValue({ imported: 3, skipped: 1 });
  document.body.appendChild(AdminRecipes());

  const fileInput = document.body.querySelector('[data-import-file]');
  const file = new File([JSON.stringify([{ name: 'Mojito' }])], 'recipes.json', { type: 'application/json' });
  Object.defineProperty(fileInput, 'files', { value: [file], configurable: true });

  document.body.querySelector('[data-import-submit]').click();

  await vi.waitFor(() => {
    expect(document.body.querySelector('[data-import-status]').textContent)
      .toBe('3 recipes imported, 1 skipped.');
  });
});

it('shows error message when importRecipes throws', async () => {
  importRecipes.mockRejectedValue(new Error('Upload failed'));
  document.body.appendChild(AdminRecipes());

  const fileInput = document.body.querySelector('[data-import-file]');
  const file = new File([JSON.stringify([])], 'recipes.json', { type: 'application/json' });
  Object.defineProperty(fileInput, 'files', { value: [file], configurable: true });

  document.body.querySelector('[data-import-submit]').click();

  await vi.waitFor(() => {
    expect(document.body.querySelector('[data-import-status]').textContent)
      .toBe('Upload failed');
  });
});

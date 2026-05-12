import { describe, it, expect, beforeEach, afterEach } from 'vitest';

vi.mock('../utils/markdown.js', () => ({
  renderMarkdown: vi.fn((text) => `<p>${text}</p>`),
}));

import { vi } from 'vitest';
import { MarkdownEditor } from './MarkdownEditor.js';

describe('MarkdownEditor', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
  });

  afterEach(() => {
    document.body.innerHTML = '';
  });

  it('renders a textarea with the given name attribute', () => {
    const el = MarkdownEditor({ name: 'notes', placeholder: 'Enter notes…' });
    document.body.appendChild(el);
    const textarea = document.body.querySelector('textarea[name="notes"]');
    expect(textarea).not.toBeNull();
    expect(textarea.placeholder).toBe('Enter notes…');
  });

  it('renders a Preview toggle button', () => {
    const el = MarkdownEditor({ name: 'notes' });
    document.body.appendChild(el);
    const btn = document.body.querySelector('button');
    expect(btn).not.toBeNull();
    expect(btn.textContent).toBe('Preview');
  });

  it('toggle button has aria-pressed="false" in edit mode and aria-pressed="true" in preview mode', () => {
    const el = MarkdownEditor({ name: 'notes', value: 'hello' });
    document.body.appendChild(el);
    const btn = document.body.querySelector('button');
    expect(btn.getAttribute('aria-pressed')).toBe('false');
    btn.click();
    expect(btn.getAttribute('aria-pressed')).toBe('true');
    btn.click();
    expect(btn.getAttribute('aria-pressed')).toBe('false');
  });

  it('clicking Preview hides the textarea and shows rendered markdown', () => {
    const el = MarkdownEditor({ name: 'notes', value: '**bold**' });
    document.body.appendChild(el);
    const btn = document.body.querySelector('button');
    const textarea = document.body.querySelector('textarea');
    btn.click();
    expect(textarea.style.display).toBe('none');
    const preview = document.body.querySelector('[data-preview]');
    expect(preview).not.toBeNull();
    expect(preview.style.display).not.toBe('none');
    expect(btn.textContent).toBe('Edit');
  });

  it('clicking Edit after Preview restores the original markdown text in the textarea', () => {
    const el = MarkdownEditor({ name: 'notes', value: '**bold**' });
    document.body.appendChild(el);
    const btn = document.body.querySelector('button');
    const textarea = document.body.querySelector('textarea');
    btn.click();
    btn.click();
    expect(textarea.style.display).not.toBe('none');
    expect(textarea.value).toBe('**bold**');
    expect(btn.textContent).toBe('Preview');
  });

  it('textarea value is accessible via its name attribute while in preview mode (form can read it)', () => {
    const form = document.createElement('form');
    const el = MarkdownEditor({ name: 'notes', value: '# heading' });
    form.appendChild(el);
    document.body.appendChild(form);
    const btn = form.querySelector('button');
    btn.click();
    const textarea = form.querySelector('[name="notes"]');
    expect(textarea).not.toBeNull();
    expect(textarea.value).toBe('# heading');
  });
});

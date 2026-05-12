import { describe, it, expect } from 'vitest';
import { renderMarkdown } from './markdown.js';

describe('renderMarkdown', () => {
  it('returns empty string for empty input', () => {
    expect(renderMarkdown('')).toBe('');
  });

  it('renders plain text as paragraph', () => {
    const result = renderMarkdown('hello world');
    expect(result).toContain('hello world');
  });

  it('renders **bold** as <strong>', () => {
    const result = renderMarkdown('**bold**');
    expect(result).toContain('<strong>bold</strong>');
  });

  it('renders [text](url) link with target=_blank and rel=noopener noreferrer', () => {
    const result = renderMarkdown('[click](https://example.com)');
    expect(result).toContain('href="https://example.com"');
    expect(result).toContain('target="_blank"');
    expect(result).toContain('rel="noopener noreferrer"');
  });

  it('strips ![alt](url) image — no <img> in output', () => {
    const result = renderMarkdown('![alt text](https://example.com/img.png)');
    expect(result).not.toContain('<img');
  });

  it('strips <script>alert(1)</script> from output', () => {
    const result = renderMarkdown('<script>alert(1)</script>');
    expect(result).not.toContain('<script');
    expect(result).not.toContain('alert(1)');
  });
});

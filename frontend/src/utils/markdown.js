import { marked, Renderer } from 'marked';
import DOMPurify from 'dompurify';

const renderer = new Renderer();

renderer.image = () => '';

renderer.link = ({ href, title, text }) => {
  const titleAttr = title ? ` title="${title}"` : '';
  return `<a href="${href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
};

marked.use({ renderer });

export function renderMarkdown(text) {
  if (!text) return '';
  const html = marked.parse(text);
  return DOMPurify.sanitize(html, { FORBID_TAGS: ['img'], ADD_ATTR: ['target'] });
}

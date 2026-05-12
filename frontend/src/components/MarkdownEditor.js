import { renderMarkdown } from '../utils/markdown.js';

export function MarkdownEditor({ name = 'notes', placeholder = '', value = '' } = {}) {
  const root = document.createElement('div');
  root.className = 'markdown-editor';

  const toolbar = document.createElement('div');
  toolbar.className = 'flex items-center justify-between mb-1';

  const label = document.createElement('span');
  label.className = 'text-sm font-medium text-gray-700';
  label.textContent = 'Notes';
  toolbar.appendChild(label);

  const toggleBtn = document.createElement('button');
  toggleBtn.type = 'button';
  toggleBtn.textContent = 'Preview';
  toggleBtn.setAttribute('aria-pressed', 'false');
  toggleBtn.className = 'text-sm text-indigo-600 hover:text-indigo-800 border border-indigo-300 rounded px-2 py-0.5';
  toolbar.appendChild(toggleBtn);

  const textarea = document.createElement('textarea');
  textarea.name = name;
  textarea.placeholder = placeholder;
  textarea.value = value;
  textarea.rows = 3;
  textarea.className = 'w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none';

  const preview = document.createElement('div');
  preview.setAttribute('data-preview', '');
  preview.className = 'prose w-full border border-gray-200 rounded-lg px-3 py-2 min-h-[4.5rem] text-gray-700';
  preview.style.display = 'none';

  toggleBtn.addEventListener('click', () => {
    const isPreview = toggleBtn.getAttribute('aria-pressed') === 'true';
    if (isPreview) {
      textarea.style.display = '';
      preview.style.display = 'none';
      preview.innerHTML = '';
      toggleBtn.textContent = 'Preview';
      toggleBtn.setAttribute('aria-pressed', 'false');
    } else {
      textarea.style.display = 'none';
      preview.innerHTML = renderMarkdown(textarea.value);
      preview.style.display = '';
      toggleBtn.textContent = 'Edit';
      toggleBtn.setAttribute('aria-pressed', 'true');
    }
  });

  root.appendChild(toolbar);
  root.appendChild(textarea);
  root.appendChild(preview);
  return root;
}

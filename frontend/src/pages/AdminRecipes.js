import { downloadRecipeSchema, exportRecipes, importRecipes } from '../api/client.js';
import { getToken } from '../api/auth.js';

export function AdminRecipes() {
  const root = document.createElement('div');
  root.className = 'max-w-4xl mx-auto px-4 py-8 space-y-10';

  const heading = document.createElement('h1');
  heading.className = 'text-2xl font-bold text-stone-900';
  heading.textContent = 'Admin · Recipes';
  root.appendChild(heading);

  root.appendChild(buildSchemaSection());
  root.appendChild(buildExportSection());
  root.appendChild(buildImportSection());

  return root;
}

function buildSchemaSection() {
  const section = document.createElement('section');
  section.className = 'border border-stone-200 rounded-xl p-6 space-y-3';

  const h2 = document.createElement('h2');
  h2.className = 'text-lg font-semibold text-stone-800';
  h2.textContent = 'Recipe Schema';

  const desc = document.createElement('p');
  desc.className = 'text-stone-600 text-sm';
  desc.textContent = 'Download the JSON Schema document that defines the recipe structure for import/export.';

  const btn = document.createElement('button');
  btn.setAttribute('data-download-schema', '');
  btn.className = 'bg-amber-500 hover:bg-amber-600 text-stone-900 font-semibold px-4 py-2 rounded-xl text-sm';
  btn.textContent = 'Download Schema';

  btn.addEventListener('click', async () => {
    btn.disabled = true;
    btn.textContent = 'Downloading…';
    try {
      const blob = await downloadRecipeSchema(getToken());
      triggerDownload(blob, 'recipe-schema.json');
    } catch (e) {
      alert(e.message || 'Download failed');
    } finally {
      btn.disabled = false;
      btn.textContent = 'Download Schema';
    }
  });

  section.appendChild(h2);
  section.appendChild(desc);
  section.appendChild(btn);
  return section;
}

function buildExportSection() {
  const section = document.createElement('section');
  section.className = 'border border-stone-200 rounded-xl p-6 space-y-3';

  const h2 = document.createElement('h2');
  h2.className = 'text-lg font-semibold text-stone-800';
  h2.textContent = 'Export Recipes';

  const desc = document.createElement('p');
  desc.className = 'text-stone-600 text-sm';
  desc.textContent = 'Download all recipes as a JSON file. The file can be re-imported using the import control below.';

  const btn = document.createElement('button');
  btn.setAttribute('data-export-recipes', '');
  btn.className = 'bg-amber-500 hover:bg-amber-600 text-stone-900 font-semibold px-4 py-2 rounded-xl text-sm';
  btn.textContent = 'Export Recipes';

  btn.addEventListener('click', async () => {
    btn.disabled = true;
    btn.textContent = 'Exporting…';
    try {
      const blob = await exportRecipes(getToken());
      triggerDownload(blob, 'recipes-export.json');
    } catch (e) {
      alert(e.message || 'Export failed');
    } finally {
      btn.disabled = false;
      btn.textContent = 'Export Recipes';
    }
  });

  section.appendChild(h2);
  section.appendChild(desc);
  section.appendChild(btn);
  return section;
}

function buildImportSection() {
  const section = document.createElement('section');
  section.className = 'border border-stone-200 rounded-xl p-6 space-y-3';

  const h2 = document.createElement('h2');
  h2.className = 'text-lg font-semibold text-stone-800';
  h2.textContent = 'Import Recipes';

  const desc = document.createElement('p');
  desc.className = 'text-stone-600 text-sm';
  desc.textContent = 'Select a JSON file conforming to the recipe schema. Existing recipes with the same name will be skipped.';

  const fileInput = document.createElement('input');
  fileInput.type = 'file';
  fileInput.accept = '.json';
  fileInput.setAttribute('data-import-file', '');
  fileInput.className = 'block text-sm text-stone-600';

  const submitBtn = document.createElement('button');
  submitBtn.setAttribute('data-import-submit', '');
  submitBtn.className = 'bg-amber-500 hover:bg-amber-600 text-stone-900 font-semibold px-4 py-2 rounded-xl text-sm';
  submitBtn.textContent = 'Import';

  const status = document.createElement('div');
  status.setAttribute('data-import-status', '');
  status.className = 'text-sm';

  submitBtn.addEventListener('click', async () => {
    const file = fileInput.files && fileInput.files[0];
    if (!file) {
      setStatus(status, 'Please select a JSON file.', 'error');
      return;
    }
    submitBtn.disabled = true;
    submitBtn.textContent = 'Importing…';
    setStatus(status, '', '');
    try {
      const text = await readFileText(file);
      let parsed;
      try {
        parsed = JSON.parse(text);
      } catch {
        setStatus(status, 'The selected file is not valid JSON.', 'error');
        return;
      }
      if (!Array.isArray(parsed)) {
        setStatus(status, 'The selected file must be a JSON array of recipes.', 'error');
        return;
      }
      const result = await importRecipes(parsed, getToken());
      setStatus(status, `${result.imported} recipes imported, ${result.skipped} skipped.`, 'success');
    } catch (e) {
      setStatus(status, e.message || 'Import failed.', 'error');
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Import';
    }
  });

  section.appendChild(h2);
  section.appendChild(desc);
  section.appendChild(fileInput);
  section.appendChild(submitBtn);
  section.appendChild(status);
  return section;
}

function triggerDownload(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

function readFileText(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target.result);
    reader.onerror = () => reject(new Error('Failed to read file'));
    reader.readAsText(file);
  });
}

function setStatus(el, message, type) {
  el.textContent = message;
  el.className = 'text-sm ' + (type === 'error' ? 'text-red-600' : type === 'success' ? 'text-green-700' : '');
}

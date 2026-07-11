// RelatedCocktailPicker is an accessible combobox for choosing related cocktails.
// Type to filter cocktail names (case-insensitive substring), Arrow Up/Down to
// move the highlighted option, Enter to add it, Escape to close. Selected
// cocktails render as removable chips. Call el.getSelectedIds() to read the
// current selection (submitted as related_ids when the recipe form is saved).
export function RelatedCocktailPicker({ names = [], selectedIds = [], currentId = null } = {}) {
  const byId = new Map(names.map((n) => [n.id, n]));
  let selected = selectedIds.map((id) => byId.get(id)).filter(Boolean);
  let options = [];
  let active = -1;

  const el = document.createElement('div');
  el.className = 'space-y-2';

  const heading = document.createElement('label');
  heading.className = 'block text-sm font-medium text-stone-700';
  heading.textContent = 'Related cocktails';
  el.appendChild(heading);

  const chips = document.createElement('div');
  chips.className = 'flex flex-wrap gap-2';
  el.appendChild(chips);

  const inputWrap = document.createElement('div');
  inputWrap.className = 'relative';
  const input = document.createElement('input');
  input.type = 'text';
  input.setAttribute('role', 'combobox');
  input.setAttribute('aria-expanded', 'false');
  input.setAttribute('aria-autocomplete', 'list');
  input.setAttribute('aria-controls', 'rcp-listbox');
  input.placeholder = 'Search cocktails by name…';
  input.className =
    'w-full border border-stone-300 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-400 focus:border-transparent';
  inputWrap.appendChild(input);

  const list = document.createElement('ul');
  list.id = 'rcp-listbox';
  list.setAttribute('role', 'listbox');
  list.className =
    'absolute z-10 mt-1 w-full bg-white border border-stone-200 rounded-xl shadow-sm max-h-56 overflow-auto empty:hidden';
  inputWrap.appendChild(list);
  el.appendChild(inputWrap);

  const selectedIdSet = () => new Set(selected.map((s) => s.id));

  el.getSelectedIds = () => selected.map((s) => s.id);

  function renderChips() {
    chips.innerHTML = '';
    selected.forEach((s) => {
      const chip = document.createElement('span');
      chip.setAttribute('data-chip', s.id);
      chip.className =
        'inline-flex items-center gap-1 bg-amber-100 text-stone-800 rounded-full pl-3 pr-1 py-1 text-sm';
      chip.append(s.name);
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.setAttribute('aria-label', `Remove ${s.name}`);
      btn.className =
        'ml-1 w-5 h-5 rounded-full text-stone-600 hover:bg-amber-200 focus:outline-none focus:ring-2 focus:ring-amber-400';
      btn.textContent = '×';
      btn.addEventListener('click', () => {
        selected = selected.filter((x) => x.id !== s.id);
        renderChips();
        renderOptions(input.value);
      });
      chip.appendChild(btn);
      chips.appendChild(chip);
    });
  }

  function renderOptions(q) {
    const query = q.trim().toLowerCase();
    const sel = selectedIdSet();
    options =
      query === ''
        ? []
        : names
            .filter((n) => n.id !== currentId && !sel.has(n.id) && n.name.toLowerCase().includes(query))
            .sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));
    active = -1;
    list.innerHTML = '';
    input.setAttribute('aria-expanded', options.length ? 'true' : 'false');

    options.forEach((o) => {
      const li = document.createElement('li');
      li.id = `rcp-opt-${o.id}`;
      li.setAttribute('role', 'option');
      li.setAttribute('aria-selected', 'false');
      li.className = 'px-3 py-2 cursor-pointer hover:bg-amber-50';
      li.textContent = o.name;
      // mousedown (not click) so it fires before the input blur closes the list.
      li.addEventListener('mousedown', (e) => {
        e.preventDefault();
        add(o);
      });
      list.appendChild(li);
    });

    if (query !== '' && options.length === 0) {
      const li = document.createElement('li');
      li.className = 'px-3 py-2 text-stone-400 text-sm';
      li.textContent = 'No matches';
      list.appendChild(li);
    }
    updateActive();
  }

  function updateActive() {
    [...list.querySelectorAll('[role="option"]')].forEach((li, i) => {
      li.setAttribute('aria-selected', i === active ? 'true' : 'false');
      li.classList.toggle('bg-amber-100', i === active);
    });
    if (active >= 0 && options[active]) {
      input.setAttribute('aria-activedescendant', `rcp-opt-${options[active].id}`);
    } else {
      input.removeAttribute('aria-activedescendant');
    }
  }

  function add(o) {
    if (!selected.some((s) => s.id === o.id)) selected.push(o);
    input.value = '';
    renderChips();
    renderOptions('');
    input.focus();
  }

  input.addEventListener('input', () => renderOptions(input.value));
  input.addEventListener('keydown', (e) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (options.length) {
        active = (active + 1) % options.length;
        updateActive();
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (options.length) {
        active = (active - 1 + options.length) % options.length;
        updateActive();
      }
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (active >= 0 && options[active]) add(options[active]);
    } else if (e.key === 'Escape') {
      renderOptions('');
    }
  });
  input.addEventListener('blur', () => setTimeout(() => renderOptions(''), 100));

  renderChips();
  return el;
}

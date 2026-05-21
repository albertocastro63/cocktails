export function SortButtonGroup({ currentDir, onSort }) {
  const group = document.createElement('div');
  group.setAttribute('role', 'group');
  group.setAttribute('aria-label', 'Sort recipes');
  group.className = 'inline-flex rounded-lg overflow-hidden border border-stone-200';

  const buttons = [
    { dir: 'asc', label: 'A→Z' },
    { dir: 'desc', label: 'Z→A' },
  ];

  buttons.forEach(({ dir, label }) => {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.textContent = label;
    btn.dataset.dir = dir;
    btn.setAttribute('aria-pressed', String(currentDir === dir));

    const isActive = currentDir === dir;
    btn.className = isActive
      ? 'bg-amber-100 text-amber-800 border border-amber-300 font-semibold px-3 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500'
      : 'bg-white text-stone-600 border border-stone-200 hover:bg-stone-50 px-3 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500';

    btn.addEventListener('click', () => onSort(dir));
    group.appendChild(btn);
  });

  return group;
}

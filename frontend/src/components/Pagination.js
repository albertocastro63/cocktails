// Pagination renders Previous / Next controls plus a "Page X of Y" indicator.
// It is presentational: it reports the requested page through onPage and holds
// no state of its own, so the parent owns the current page. Returns null-ish
// (an empty fragment container) when there is only a single page.
export function Pagination({ page, totalPages, onPage }) {
  const nav = document.createElement('nav');
  nav.setAttribute('aria-label', 'Recipe pages');
  nav.className = 'flex items-center justify-center gap-4 mt-6';

  // A single page needs no controls; hand back an empty, non-rendering node.
  if (!totalPages || totalPages <= 1) {
    nav.dataset.empty = 'true';
    return nav;
  }

  const makeButton = (label, targetPage, enabled) => {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.textContent = label;
    btn.dataset.page = String(targetPage);
    btn.disabled = !enabled;
    btn.setAttribute('aria-disabled', String(!enabled));
    btn.className = enabled
      ? 'bg-white text-stone-700 border border-stone-200 hover:bg-stone-50 px-3 py-1 text-sm rounded-lg focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500'
      : 'bg-stone-50 text-stone-300 border border-stone-200 px-3 py-1 text-sm rounded-lg cursor-not-allowed';
    if (enabled) {
      btn.addEventListener('click', () => onPage(targetPage));
    }
    return btn;
  };

  const prev = makeButton('← Previous', page - 1, page > 1);
  prev.dataset.dir = 'prev';

  const indicator = document.createElement('span');
  indicator.className = 'text-stone-600 text-sm';
  indicator.dataset.testid = 'page-indicator';
  indicator.textContent = `Page ${page} of ${totalPages}`;

  const next = makeButton('Next →', page + 1, page < totalPages);
  next.dataset.dir = 'next';

  nav.append(prev, indicator, next);
  return nav;
}

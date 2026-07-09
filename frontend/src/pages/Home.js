import { getRandomRecipe } from '../api/client.js';
import { EmptyState } from '../components/EmptyState.js';
import { IngredientList } from '../components/IngredientList.js';
import { PropertyTable } from '../components/PropertyTable.js';
import { renderMarkdown } from '../utils/markdown.js';

export function Home() {
  const el = document.createElement('div');

  const hero = document.createElement('div');
  hero.className = 'bg-gradient-to-br from-stone-900 to-stone-800 text-white py-4 md:py-16 px-4';
  const heroInner = document.createElement('div');
  // On phones: text on the left, CTA on the right (one row). On md+: stacked
  // (block), exactly as today.
  heroInner.className = 'max-w-2xl mx-auto flex items-center justify-between md:block';
  const heroText = document.createElement('div');
  const heroHeading = document.createElement('h1');
  heroHeading.className = 'text-xl md:text-4xl font-bold mb-0.5 md:mb-3';
  heroHeading.textContent = 'Cocktail Recipes';
  const heroSub = document.createElement('p');
  heroSub.className = 'text-amber-400 text-sm md:text-lg mb-0 md:mb-6';
  heroSub.textContent = 'Discover your next favorite drink';
  heroText.appendChild(heroHeading);
  heroText.appendChild(heroSub);
  const heroCTA = document.createElement('a');
  heroCTA.href = '#/recipes';
  // Smaller on phones (~25% less padding + smaller font); restores to today's
  // size at md+. shrink-0/whitespace-nowrap keep it intact in the phone row.
  heroCTA.className = 'inline-block shrink-0 whitespace-nowrap ml-4 md:ml-0 bg-amber-500 text-stone-900 font-semibold px-4 py-1.5 text-sm md:px-6 md:py-2 md:text-base rounded-xl hover:bg-amber-600 transition-colors';
  heroCTA.textContent = 'All Recipes';
  heroInner.appendChild(heroText);
  heroInner.appendChild(heroCTA);
  hero.appendChild(heroInner);
  el.appendChild(hero);

  const body = document.createElement('div');
  body.className = 'max-w-2xl mx-auto px-4 py-8';

  const sectionLabel = document.createElement('h2');
  sectionLabel.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mb-4';
  sectionLabel.textContent = 'Random Recipe';
  body.appendChild(sectionLabel);

  const content = document.createElement('div');
  const loadingP = document.createElement('p');
  loadingP.className = 'text-stone-500 animate-pulse py-4 text-center';
  loadingP.textContent = 'Loading…';
  content.appendChild(loadingP);
  body.appendChild(content);
  el.appendChild(body);

  getRandomRecipe().then((recipe) => {
    content.innerHTML = '';
    if (recipe === null) {
      content.appendChild(EmptyState({ message: 'No recipes yet. Add one!' }));
      return;
    }

    const title = document.createElement('h2');
    title.className = 'text-2xl font-bold text-stone-900 mb-4';
    title.textContent = recipe.name;
    content.appendChild(title);

    if (recipe.ingredients && recipe.ingredients.length > 0) {
      const h3 = document.createElement('h3');
      h3.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h3.textContent = 'Ingredients';
      content.appendChild(h3);
      content.appendChild(IngredientList({ ingredients: recipe.ingredients }));
    }

    if (recipe.steps && recipe.steps.length > 0) {
      const h3 = document.createElement('h3');
      h3.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h3.textContent = 'Steps';
      content.appendChild(h3);
      const ol = document.createElement('ol');
      ol.className = 'list-decimal list-inside space-y-1 text-stone-700';
      recipe.steps.forEach((step) => {
        const li = document.createElement('li');
        li.textContent = step;
        ol.appendChild(li);
      });
      content.appendChild(ol);
    }

    if (recipe.properties && Object.keys(recipe.properties).length > 0) {
      const h3 = document.createElement('h3');
      h3.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h3.textContent = 'Properties';
      content.appendChild(h3);
      content.appendChild(PropertyTable({ properties: recipe.properties }));
    }

    if (recipe.notes) {
      const h3 = document.createElement('h3');
      h3.className = 'text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2';
      h3.textContent = 'Notes';
      content.appendChild(h3);
      const div = document.createElement('div');
      div.className = 'prose prose-stone max-w-none overflow-x-auto';
      div.innerHTML = renderMarkdown(recipe.notes);
      content.appendChild(div);
    }
  }).catch(() => {
    content.innerHTML = '';
    const errP = document.createElement('p');
    errP.className = 'text-stone-500 py-4 text-center';
    errP.textContent = 'Failed to load recipe.';
    content.appendChild(errP);
  });

  return el;
}

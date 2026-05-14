import { getRandomRecipe } from '../api/client.js';
import { EmptyState } from '../components/EmptyState.js';
import { IngredientList } from '../components/IngredientList.js';
import { PropertyTable } from '../components/PropertyTable.js';
import { renderMarkdown } from '../utils/markdown.js';

export function Home() {
  const el = document.createElement('div');

  const hero = document.createElement('div');
  hero.className = 'bg-gradient-to-br from-stone-900 to-stone-800 text-white py-16 px-4';
  const heroInner = document.createElement('div');
  heroInner.className = 'max-w-2xl mx-auto';
  const heroHeading = document.createElement('h1');
  heroHeading.className = 'text-4xl font-bold mb-3';
  heroHeading.textContent = 'Cocktail Recipes';
  const heroSub = document.createElement('p');
  heroSub.className = 'text-amber-400 text-lg mb-6';
  heroSub.textContent = 'Discover your next favorite drink';
  const heroCTA = document.createElement('a');
  heroCTA.href = '#/recipes';
  heroCTA.className = 'inline-block bg-amber-500 text-stone-900 font-semibold px-6 py-2 rounded-xl hover:bg-amber-600 transition-colors';
  heroCTA.textContent = 'All Recipes';
  heroInner.appendChild(heroHeading);
  heroInner.appendChild(heroSub);
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
      div.className = 'text-stone-700';
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

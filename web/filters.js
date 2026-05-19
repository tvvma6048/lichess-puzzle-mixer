import { fetchPuzzleCount } from "./api.js";
import { fillThemeSelect } from "./theme-catalog.js";

export { ALL_THEME_IDS as THEME_OPTIONS } from "./theme-catalog.js";

const MAX_THEME_GROUPS = 6;
const MAX_THEMES_PER_GROUP = 8;

export function initFilters(onStartTraining) {
  const panel = document.getElementById("filter-panel");
  const form = document.getElementById("filter-form");
  const groupsRoot = document.getElementById("theme-groups");
  const countEl = document.getElementById("match-count");
  const errorEl = document.getElementById("filter-error");

  initThemeGroups(groupsRoot);
  panel.hidden = false;

  const startBtn = document.getElementById("btn-start-training");
  if (startBtn && onStartTraining) {
    startBtn.addEventListener("click", () => onStartTraining(form));
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    errorEl.hidden = true;
    errorEl.textContent = "";
    countEl.textContent = "Counting…";

    const filter = readFilterFromForm(form);
    const themeLabel = formatThemeGroups(filter.theme_groups);

    try {
      const { count } = await fetchPuzzleCount(filter);
      countEl.textContent = `${count.toLocaleString()} puzzles match [${themeLabel}]`;

      if (count === 0 && filter.theme_groups.length > 0) {
        countEl.textContent += " — try fewer groups or themes";
      }
    } catch (err) {
      errorEl.hidden = false;
      errorEl.textContent = err.message;
      countEl.textContent = "";
    }
  });
}

export function formatThemeGroups(groups) {
  if (!groups?.length) {
    return "any theme";
  }
  return groups
    .map((g) => (g.length === 1 ? g[0] : `(${g.join(" or ")})`))
    .join(" and ");
}

function initThemeGroups(root) {
  if (!root) return;

  root.addEventListener("change", (e) => {
    if (!e.target.matches(".theme-select")) return;
    syncThemeGroupsUI(root);
  });

  root.addEventListener("click", (e) => {
    const removeBtn = e.target.closest("[data-remove-group]");
    if (removeBtn) {
      const groupEl = removeBtn.closest(".theme-group");
      if (groupEl && root.querySelectorAll(".theme-group").length > 1) {
        groupEl.remove();
      }
      return;
    }
  });

  document.getElementById("btn-add-theme-group")?.addEventListener("click", () => {
    const groups = readThemeGroupsFromDOM(root);
    if (groups.length >= MAX_THEME_GROUPS) return;
    appendThemeGroup(root, []);
    syncThemeGroupsUI(root);
  });

  if (root.querySelectorAll(".theme-group").length === 0) {
    appendThemeGroup(root, []);
  }
}

function readThemeGroupsFromDOM(root) {
  const groups = [];
  for (const groupEl of root.querySelectorAll(".theme-group")) {
    const themes = [...groupEl.querySelectorAll(".theme-select")]
      .map((s) => s.value)
      .filter(Boolean);
    if (themes.length > 0) {
      groups.push([...new Set(themes)]);
    }
  }
  return groups;
}

function syncThemeGroupsUI(root) {
  if (root.querySelectorAll(".theme-group").length === 0) {
    appendThemeGroup(root, []);
    return;
  }

  for (const groupEl of root.querySelectorAll(".theme-group")) {
    const values = [...groupEl.querySelectorAll(".theme-select")]
      .map((s) => s.value)
      .filter(Boolean);
    rebuildGroupSelects(groupEl, [...new Set(values)]);
  }

  const groupCount = root.querySelectorAll(".theme-group").length;
  for (const btn of root.querySelectorAll("[data-remove-group]")) {
    btn.hidden = groupCount <= 1;
  }
}

function appendThemeGroup(root, selectedValues) {
  const groupEl = document.createElement("div");
  groupEl.className = "theme-group";
  groupEl.dataset.testid = "theme-group";

  const themesRow = document.createElement("div");
  themesRow.className = "theme-group-themes";
  groupEl.appendChild(themesRow);

  const removeBtn = document.createElement("button");
  removeBtn.type = "button";
  removeBtn.className = "btn-link theme-group-remove";
  removeBtn.dataset.removeGroup = "";
  removeBtn.dataset.testid = "btn-remove-theme-group";
  removeBtn.textContent = "Remove group";
  removeBtn.title = "Remove this AND group";
  groupEl.appendChild(removeBtn);

  root.appendChild(groupEl);
  rebuildGroupSelects(groupEl, selectedValues);
}

function rebuildGroupSelects(groupEl, selectedValues) {
  const row = groupEl.querySelector(".theme-group-themes");
  if (!row) return;

  const unique = [...new Set(selectedValues)].slice(0, MAX_THEMES_PER_GROUP);
  row.replaceChildren();

  const slots = unique.length < MAX_THEMES_PER_GROUP ? unique.length + 1 : unique.length;
  for (let i = 0; i < slots; i++) {
    if (i > 0) {
      const or = document.createElement("span");
      or.className = "theme-or";
      or.textContent = "or";
      row.appendChild(or);
    }
    row.appendChild(createThemeSelect(unique[i] ?? ""));
  }
}

function createThemeSelect(selectedValue) {
  const select = document.createElement("select");
  select.className = "theme-select";
  select.dataset.testid = "theme-dropdown";
  fillThemeSelect(select, { selectedValue });
  return select;
}

export function readFilterFromForm(form) {
  const root = document.getElementById("theme-groups");
  return {
    theme_groups: readThemeGroupsFromDOM(root),
    rating_min: Number(form.elements.rating_min.value),
    rating_max: Number(form.elements.rating_max.value),
    popularity_min: Number(form.elements.popularity_min.value),
    side_to_move: form.elements.side_to_move.value,
    length_min: Number(form.elements.length_min.value),
    length_max: Number(form.elements.length_max.value),
    opening_family: null,
  };
}

export function saveFilterForPlay(filter) {
  sessionStorage.setItem("activeFilter", JSON.stringify(filter));
}

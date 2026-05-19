/**
 * Lichess puzzle theme categories (see lila PuzzleTheme.categorized).
 * Theme ids match the Lichess puzzle database tags.
 */
export const THEME_CATALOG = [
  {
    label: "Phases",
    themes: [
      "bishopEndgame",
      "endgame",
      "knightEndgame",
      "middlegame",
      "opening",
      "pawnEndgame",
      "queenEndgame",
      "queenRookEndgame",
      "rookEndgame",
    ],
  },
  {
    label: "Motifs",
    themes: [
      "advancedPawn",
      "attackingF2F7",
      "capturingDefender",
      "discoveredAttack",
      "doubleCheck",
      "exposedKing",
      "fork",
      "hangingPiece",
      "kingsideAttack",
      "pin",
      "queensideAttack",
      "sacrifice",
      "skewer",
      "trappedPiece",
    ],
  },
  {
    label: "Advanced",
    themes: [
      "attraction",
      "clearance",
      "collinearMove",
      "defensiveMove",
      "deflection",
      "discoveredCheck",
      "interference",
      "intermezzo",
      "quietMove",
      "xRayAttack",
      "zugzwang",
    ],
  },
  {
    label: "Checkmates",
    themes: ["mate", "mateIn1", "mateIn2", "mateIn3", "mateIn4", "mateIn5"],
  },
  {
    label: "Mate patterns",
    themes: [
      "anastasiaMate",
      "arabianMate",
      "backRankMate",
      "balestraMate",
      "blindSwineMate",
      "bodenMate",
      "cornerMate",
      "doubleBishopMate",
      "dovetailMate",
      "epauletteMate",
      "hookMate",
      "killBoxMate",
      "morphysMate",
      "operaMate",
      "pillsburysMate",
      "smotheredMate",
      "swallowstailMate",
      "triangleMate",
      "vukovicMate",
    ],
  },
  {
    label: "Special moves",
    themes: ["castling", "enPassant", "promotion", "underPromotion"],
  },
  {
    label: "Goals",
    themes: ["advantage", "crushing", "equality"],
  },
  {
    label: "Lengths",
    themes: ["long", "oneMove", "short", "veryLong"],
  },
  {
    label: "Origin",
    themes: ["master", "masterVsMaster", "superGM"],
  },
].map((group) => ({
  ...group,
  themes: [...group.themes].sort((a, b) => a.localeCompare(b)),
}));

/** Flat list of all catalog theme ids (sorted). */
export const ALL_THEME_IDS = THEME_CATALOG.flatMap((g) => g.themes).sort((a, b) =>
  a.localeCompare(b),
);

/** @deprecated Use ALL_THEME_IDS — kept for tests/imports that expect THEME_OPTIONS */
export const THEME_OPTIONS = ALL_THEME_IDS;

export function themeDisplayName(id) {
  return id
    .replace(/([A-Z])/g, " $1")
    .replace(/^./, (c) => c.toUpperCase())
    .trim();
}

/** Populate a &lt;select&gt; with Lichess-grouped theme options. */
export function fillThemeSelect(select, { placeholder = "(theme)", selectedValue = "" } = {}) {
  select.replaceChildren();

  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = placeholder;
  select.appendChild(empty);

  for (const { label, themes } of THEME_CATALOG) {
    const group = document.createElement("optgroup");
    group.label = label;
    for (const id of themes) {
      const opt = document.createElement("option");
      opt.value = id;
      opt.textContent = themeDisplayName(id);
      group.appendChild(opt);
    }
    select.appendChild(group);
  }

  select.value = selectedValue;
}

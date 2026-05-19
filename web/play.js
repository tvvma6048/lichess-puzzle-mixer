import { fetchNextPuzzle } from "./api.js";
import { PuzzleBoard } from "./board.js";
import { renderMoveHistory } from "./move-history.js";

const FILTER_KEY = "activeFilter";

export function initPlay() {
  const filter = loadFilter();
  if (!filter) {
    window.location.hash = "";
    return;
  }

  const boardEl = document.getElementById("board");
  const statusEl = document.getElementById("play-status");
  const ratingEl = document.getElementById("play-rating");
  const themesEl = document.getElementById("play-themes");
  const feedbackEl = document.getElementById("play-feedback");
  const movesEl = document.getElementById("play-moves");
  const nextBtn = document.getElementById("btn-next");
  const hintBtn = document.getElementById("btn-hint");
  const solutionBtn = document.getElementById("btn-solution");
  const restartBtn = document.getElementById("btn-restart");
  const prevBtn = document.getElementById("btn-prev");
  const forwardBtn = document.getElementById("btn-forward");
  const lastBtn = document.getElementById("btn-last");
  const backBtn = document.getElementById("btn-back-filters");

  initPlayMetaToggle();

  const setHelpEnabled = (enabled) => {
    if (hintBtn) hintBtn.disabled = !enabled;
    if (solutionBtn) solutionBtn.disabled = !enabled;
  };

  const updateHistoryUi = (state) => {
    renderMoveHistory(
      movesEl,
      state.history,
      state.cursor,
      state.displayStartFen || state.startFen,
    );
    if (restartBtn) restartBtn.disabled = !state.canRestart;
    if (prevBtn) prevBtn.disabled = !state.canStepBack;
    if (forwardBtn) forwardBtn.disabled = !state.canStepForward;
    if (lastBtn) lastBtn.disabled = !state.canStepForward;
  };

  let board = new PuzzleBoard(boardEl, {
    onStatus: (text) => {
      statusEl.textContent = text;
    },
    onHelpChange: setHelpEnabled,
    onHistoryChange: updateHistoryUi,
    onSolved: () => {
      feedbackEl.textContent = "Success";
      feedbackEl.className = "play-feedback is-success";
      nextBtn.disabled = false;
    },
    onRevealed: () => {
      feedbackEl.textContent = "Solution shown";
      feedbackEl.className = "play-feedback";
      nextBtn.disabled = false;
    },
  });

  movesEl?.addEventListener("click", (e) => {
    const btn = e.target.closest("[data-history-index]");
    if (!btn) return;
    board.goToHistoryIndex(Number(btn.dataset.historyIndex));
  });

  backBtn.addEventListener("click", () => {
    board.destroy();
    window.location.hash = "";
  });

  nextBtn.addEventListener("click", () => loadNext());
  document.getElementById("btn-skip")?.addEventListener("click", () => loadNext());
  hintBtn?.addEventListener("click", () => board.showHint());
  solutionBtn?.addEventListener("click", () => board.showSolution());
  restartBtn?.addEventListener("click", () => board.restart());
  prevBtn?.addEventListener("click", () => board.stepBack());
  forwardBtn?.addEventListener("click", () => board.stepForward());
  lastBtn?.addEventListener("click", () => board.stepToEnd());

  async function loadNext() {
    nextBtn.disabled = true;
    setHelpEnabled(false);
    feedbackEl.textContent = "";
    feedbackEl.className = "play-feedback";
    ratingEl.textContent = "…";
    themesEl.textContent = "…";
    setLichessLinks(null);
    statusEl.textContent = "Loading puzzle…";
    updateHistoryUi({
      history: [],
      cursor: 0,
      startFen: "",
      canRestart: false,
      canStepBack: false,
      canStepForward: false,
    });

    try {
      const puzzle = await fetchNextPuzzle(filter);
      ratingEl.textContent = String(puzzle.rating);
      themesEl.textContent = (puzzle.themes || []).join(", ") || "—";
      setLichessLinks(puzzle);
      await board.loadPuzzle(puzzle);
    } catch (err) {
      statusEl.textContent = "Error";
      feedbackEl.textContent = err.message;
      feedbackEl.className = "play-feedback is-error";
    }
  }

  loadNext();
}

function lichessPuzzleUrl(puzzleId) {
  return `https://lichess.org/training/${encodeURIComponent(puzzleId)}`;
}

function setLichessLinks(puzzle) {
  const puzzleLink = document.getElementById("play-lichess-puzzle");
  const gameLink = document.getElementById("play-lichess-game");
  if (!puzzleLink || !gameLink) return;

  if (puzzle?.puzzle_id) {
    puzzleLink.href = lichessPuzzleUrl(puzzle.puzzle_id);
    puzzleLink.title = `Lichess puzzle ${puzzle.puzzle_id}`;
    puzzleLink.hidden = false;
  } else {
    puzzleLink.hidden = true;
  }

  if (puzzle?.game_url) {
    gameLink.href = puzzle.game_url;
    gameLink.hidden = false;
  } else {
    gameLink.hidden = true;
  }
}

function initPlayMetaToggle() {
  const btn = document.getElementById("btn-toggle-play-meta");
  const section = document.getElementById("play-meta");
  if (!btn || !section) return;

  btn.addEventListener("click", () => {
    const show = section.classList.contains("play-meta--collapsed");
    section.classList.toggle("play-meta--collapsed", !show);
    btn.setAttribute("aria-expanded", String(show));
    btn.textContent = show ? "Hide rating & themes" : "Show rating & themes";
  });
}

function loadFilter() {
  try {
    const raw = sessionStorage.getItem(FILTER_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

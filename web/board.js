import { fetchGamePreamble } from "./api.js";
import { buildPreambleFromPgn } from "./preamble.js";
import { Chessground } from "./vendor/chessground.bundle.js";
import { Chess } from "./vendor/chess.js";

const SETUP_DELAY_MS = 400;
const OPPONENT_REPLY_DELAY_MS = 250;
const SOLUTION_STEP_MS = 450;

export class PuzzleBoard {
  #container;
  #cg = null;
  #chess = null;
  #puzzle = null;
  #moveIndex = 0;
  #phase = "idle";
  #hintLevel = 0;
  #opponentTimer = null;
  #startFen = "";
  #history = [];
  #historyCursor = 0;
  #historyStart = 0;
  #preambleLoadId = 0;
  #revealed = false;
  #onSolved = null;
  #onRevealed = null;
  #onStatus = null;
  #onHelpChange = null;
  #onHistoryChange = null;

  constructor(container, { onSolved, onRevealed, onStatus, onHelpChange, onHistoryChange } = {}) {
    this.#container = container;
    this.#onSolved = onSolved;
    this.#onRevealed = onRevealed;
    this.#onStatus = onStatus;
    this.#onHelpChange = onHelpChange;
    this.#onHistoryChange = onHistoryChange;
  }

  getHistoryState() {
    const displayStartFen =
      this.#history[0]?.preamble ? this.#history[0].fen : this.#startFen;
    return {
      history: this.#history,
      cursor: this.#historyCursor,
      startFen: this.#startFen,
      displayStartFen,
      canRestart: this.#historyCursor > this.#historyStart,
      canStepBack: this.#historyCursor > 0,
      canStepForward: this.#historyCursor < this.#history.length - 1,
    };
  }

  restart() {
    return this.#goToHistoryIndex(this.#historyStart);
  }

  stepBack() {
    if (this.#historyCursor <= 0) return false;
    return this.#goToHistoryIndex(this.#historyCursor - 1);
  }

  stepForward() {
    if (this.#historyCursor >= this.#history.length - 1) return false;
    return this.#goToHistoryIndex(this.#historyCursor + 1);
  }

  stepToEnd() {
    if (this.#history.length === 0) return false;
    return this.#goToHistoryIndex(this.#history.length - 1);
  }

  goToHistoryIndex(index) {
    if (index < 0 || index >= this.#history.length) return false;
    return this.#goToHistoryIndex(index);
  }

  destroy() {
    this.#clearOpponentTimer();
    this.#cg?.destroy();
    this.#cg = null;
    this.#container.innerHTML = "";
    this.#container.classList.remove("is-solved", "is-revealed");
  }

  canUseHelp() {
    return this.#phase === "player";
  }

  showHint() {
    if (!this.canUseHelp()) return false;

    const uci = this.#puzzle.moves[this.#moveIndex];
    if (!uci) return false;

    const from = uci.slice(0, 2);
    const to = uci.slice(2, 4);
    const custom = new Map();
    let arrow = null;

    if (this.#hintLevel < 1) {
      this.#hintLevel = 1;
      custom.set(from, "hint-from");
      this.#setStatus("Hint: move this piece — hint again for square");
    } else {
      this.#hintLevel = 2;
      custom.set(from, "hint-from");
      const capture = Boolean(this.#chess.get(to));
      custom.set(to, capture ? "move-dest oc" : "move-dest");
      arrow = { from, to };
      this.#setStatus("Hint: move here");
    }

    this.#applyHintOverlay(custom, arrow);
    return true;
  }

  async showSolution() {
    if (!this.canUseHelp()) return false;

    this.#clearHint();
    this.#clearOpponentTimer();
    this.#revealed = true;
    this.#phase = "revealed";
    this.#disablePlayer();
    this.#notifyHelpChange();

    while (this.#moveIndex < this.#puzzle.moves.length) {
      await delay(SOLUTION_STEP_MS);
      const uci = this.#puzzle.moves[this.#moveIndex];
      this.#applyUci(uci);
      this.#moveIndex++;
      this.#recordPosition(uci);
      this.#syncBoard({ lastMove: uciSquares(uci) });
      this.#notifyHistoryChange();
    }

    this.#container.classList.add("is-revealed");
    this.#setStatus("Solution");
    this.#onRevealed?.(this.#puzzle);
    return true;
  }

  async loadPuzzle(puzzle) {
    this.#preambleLoadId++;
    this.destroy();
    this.#puzzle = puzzle;
    this.#moveIndex = 0;
    this.#hintLevel = 0;
    this.#phase = "loading";
    this.#revealed = false;
    this.#startFen = puzzle.fen;
    this.#history = [
      { fen: puzzle.fen, moveIndex: 0, lastMove: undefined, san: undefined, uci: undefined },
    ];
    this.#historyCursor = 0;
    this.#historyStart = 0;

    const wrap = document.createElement("div");
    wrap.className = "cg-wrap";
    this.#container.appendChild(wrap);

    this.#chess = new Chess(puzzle.fen);
    const solver = solverFromFen(puzzle.fen);

    this.#cg = Chessground(wrap, {
      fen: puzzle.fen,
      orientation: colorToCG(solver),
      turnColor: colorToCG(this.#chess.turn()),
      check: this.#chess.inCheck(),
      coordinates: true,
      movable: {
        color: undefined,
        dests: new Map(),
        events: {
          after: (orig, dest) => this.#onPlayerMove(orig, dest),
        },
      },
      premovable: { enabled: true },
      draggable: { enabled: false, showGhost: true },
      selectable: { enabled: true },
      highlight: { lastMove: true, check: true },
      animation: { enabled: true, duration: 200 },
      drawable: { enabled: false, visible: true, eraseOnClick: false },
    });

    this.#setStatus(`${colorName(opponentFromFen(puzzle.fen))} to move — setup`);
    await delay(SETUP_DELAY_MS);
    await this.#playSetupMove();
    void this.#loadGamePreamble();
  }

  async #playSetupMove() {
    const setup = this.#puzzle.moves[0];
    if (!setup) {
      this.#phase = "player";
      this.#historyStart = 0;
      this.#enablePlayer();
      this.#setStatus(`You play ${colorName(this.#chess.turn())}`);
      this.#notifyHelpChange();
      this.#notifyHistoryChange();
      return;
    }
    this.#phase = "setup";
    this.#applyUci(setup);
    this.#moveIndex = 1;
    this.#recordPosition(setup);
    this.#historyStart = this.#historyCursor;
    this.#syncBoard({ lastMove: uciSquares(setup) });
    this.#enablePlayer();
    this.#phase = "player";
    this.#setStatus(`You play ${colorName(this.#chess.turn())}`);
    this.#notifyHelpChange();
    this.#notifyHistoryChange();
  }

  #onPlayerMove(orig, dest) {
    if (this.#phase !== "player") {
      this.#syncBoard();
      return;
    }

    const uci = toUci(orig, dest);

    if (!this.#isCorrectMove(uci)) {
      this.#flashWrong(dest);
      this.#setStatus("Wrong — try again");
      this.#syncBoard();
      return;
    }

    this.#clearHint();
    this.#truncateHistoryForward();
    this.#applyMoveVerbose(uciToMove(uci));
    const lastMove = [orig, dest];
    this.#moveIndex++;
    this.#recordPosition(uci);

    if (this.#moveIndex >= this.#puzzle.moves.length) {
      this.#finishSolved(lastMove);
      return;
    }

    this.#phase = "opponent";
    this.#disablePlayer();
    this.#syncBoard({ lastMove });

    this.#clearOpponentTimer();
    this.#opponentTimer = window.setTimeout(() => {
      this.#opponentTimer = null;
      const reply = this.#puzzle.moves[this.#moveIndex];
      this.#applyUci(reply);
      this.#moveIndex++;
      this.#recordPosition(reply);
      this.#syncBoard({ lastMove: uciSquares(reply) });

      if (this.#moveIndex >= this.#puzzle.moves.length) {
        this.#finishSolved(uciSquares(reply));
        return;
      }

      this.#enablePlayer();
      this.#phase = "player";
      this.#hintLevel = 0;
      this.#setStatus(`You play ${colorName(this.#chess.turn())}`);
      this.#notifyHelpChange();
      this.#notifyHistoryChange();
    }, OPPONENT_REPLY_DELAY_MS);
  }

  #isCorrectMove(uci) {
    if (uci === this.#puzzle.moves[this.#moveIndex]) return true;
    if (this.#puzzle.themes?.includes("mateIn1")) {
      const trial = new Chess(this.#chess.fen());
      const result = trial.move(uciToMove(uci));
      if (result && trial.isCheckmate()) return true;
    }
    return false;
  }

  #finishSolved(lastMove) {
    this.#clearHint();
    this.#clearOpponentTimer();
    this.#phase = "solved";
    this.#disablePlayer();
    this.#syncBoard({ lastMove });
    this.#container.classList.add("is-solved");
    this.#setStatus("Puzzle solved!");
    this.#notifyHelpChange();
    this.#notifyHistoryChange();
    this.#onSolved?.(this.#puzzle);
  }

  #clearHint() {
    this.#hintLevel = 0;
    this.#applyHintOverlay(new Map(), null);
  }

  #applyHintOverlay(custom, arrow) {
    this.#cg?.set({
      highlight: { lastMove: true, check: true, custom },
      drawable: {
        visible: true,
        shapes: arrow ? [{ orig: arrow.from, dest: arrow.to, brush: "green" }] : [],
      },
    });
    this.#cg?.state.dom.redraw();
  }

  #clearOpponentTimer() {
    if (this.#opponentTimer !== null) {
      window.clearTimeout(this.#opponentTimer);
      this.#opponentTimer = null;
    }
  }

  #notifyHelpChange() {
    this.#onHelpChange?.(this.canUseHelp());
  }

  #notifyHistoryChange() {
    this.#onHistoryChange?.(this.getHistoryState());
  }

  async #loadGamePreamble() {
    const loadId = ++this.#preambleLoadId;
    const gameUrl = this.#puzzle?.game_url;
    if (!gameUrl) return;

    try {
      const data = await fetchGamePreamble(gameUrl, this.#startFen);
      if (loadId !== this.#preambleLoadId) return;
      if (!data?.ok || !data.pgn) return;

      const entries = buildPreambleFromPgn(data.pgn, this.#startFen, data.ply || 0);
      if (loadId !== this.#preambleLoadId || entries.length <= 1) return;

      this.#prependPreamble(entries);
    } catch {
      // Preamble is optional; never interrupt puzzle play.
    }
  }

  #prependPreamble(entries) {
    const extra = entries.slice(0, -1);
    if (extra.length === 0) return;

    const shift = extra.length;
    const atLiveEnd = this.#historyCursor === this.#history.length - 1;

    this.#history = extra.concat(this.#history);
    this.#historyStart += shift;
    this.#historyCursor = atLiveEnd
      ? this.#history.length - 1
      : this.#historyCursor + shift;

    this.#notifyHistoryChange();
  }

  #recordPosition(uci) {
    const hist = this.#chess.history({ verbose: true });
    const last = hist[hist.length - 1];
    this.#history.push({
      fen: this.#chess.fen(),
      moveIndex: this.#moveIndex,
      lastMove: uci ? uciSquares(uci) : undefined,
      san: last?.san,
      uci,
    });
    this.#historyCursor = this.#history.length - 1;
  }

  #truncateHistoryForward() {
    if (this.#historyCursor < this.#history.length - 1) {
      this.#history = this.#history.slice(0, this.#historyCursor + 1);
    }
  }

  #goToHistoryIndex(index) {
    this.#clearOpponentTimer();
    this.#clearHint();
    this.#container.classList.remove("is-solved", "is-revealed");

    const entry = this.#history[index];
    this.#historyCursor = index;
    this.#chess.load(entry.fen);
    this.#moveIndex = entry.moveIndex;
    this.#hintLevel = 0;

    const atLiveEnd = index === this.#history.length - 1;

    if (atLiveEnd) {
      if (this.#revealed) {
        this.#phase = "revealed";
        this.#disablePlayer();
        this.#container.classList.add("is-revealed");
        this.#setStatus("Solution");
      } else if (this.#moveIndex >= this.#puzzle.moves.length) {
        this.#phase = "solved";
        this.#disablePlayer();
        this.#container.classList.add("is-solved");
        this.#setStatus("Puzzle solved!");
      } else {
        this.#phase = "player";
        this.#enablePlayer();
        this.#setStatus(`You play ${colorName(this.#chess.turn())}`);
      }
    } else if (index === this.#historyStart) {
      this.#revealed = false;
      this.#phase = "player";
      this.#enablePlayer();
      this.#setStatus(`You play ${colorName(this.#chess.turn())}`);
    } else {
      this.#revealed = false;
      this.#phase = "review";
      this.#disablePlayer();
      this.#setStatus("Reviewing moves");
    }

    this.#syncBoard({ lastMove: entry.lastMove });
    this.#notifyHelpChange();
    this.#notifyHistoryChange();
    return true;
  }

  #flashWrong(square) {
    this.#cg?.set({
      highlight: { custom: new Map([[square, "wrong-move"]]) },
    });
    window.setTimeout(() => {
      this.#cg?.set({ highlight: { custom: new Map() } });
    }, 900);
  }

  #enablePlayer() {
    const color = colorToCG(this.#chess.turn());
    this.#cg?.set({
      turnColor: color,
      movable: {
        color,
        dests: toDests(this.#chess),
        free: false,
        showDests: true,
        events: {
          after: (orig, dest) => this.#onPlayerMove(orig, dest),
        },
      },
    });
  }

  #disablePlayer() {
    this.#cg?.set({
      movable: { color: undefined, dests: new Map() },
    });
  }

  #applyUci(uci) {
    this.#chess.move(uciToMove(uci));
  }

  #applyMoveVerbose(move) {
    this.#chess.move(move);
  }

  #syncBoard({ lastMove } = {}) {
    const turn = colorToCG(this.#chess.turn());
    this.#cg?.set({
      fen: this.#chess.fen(),
      turnColor: turn,
      check: this.#chess.inCheck(),
      lastMove: lastMove || undefined,
      movable:
        this.#phase === "player"
          ? {
              color: turn,
              dests: toDests(this.#chess),
              free: false,
              showDests: true,
              events: {
                after: (orig, dest) => this.#onPlayerMove(orig, dest),
              },
            }
          : { color: undefined, dests: new Map() },
    });
  }

  #setStatus(text) {
    this.#onStatus?.(text);
  }
}

function toDests(chess) {
  const dests = new Map();
  for (const m of chess.moves({ verbose: true })) {
    if (!dests.has(m.from)) dests.set(m.from, []);
    dests.get(m.from).push(m.to);
  }
  return dests;
}

function uciToMove(uci) {
  return {
    from: uci.slice(0, 2),
    to: uci.slice(2, 4),
    promotion: uci.length > 4 ? uci[4] : undefined,
  };
}

function toUci(orig, dest) {
  return orig + dest;
}

function uciSquares(uci) {
  return [uci.slice(0, 2), uci.slice(2, 4)];
}

function solverFromFen(fen) {
  return opponentFromFen(fen);
}

function opponentFromFen(fen) {
  return fen.split(" ")[1] === "w" ? "b" : "w";
}

function colorToCG(c) {
  return c === "w" ? "white" : "black";
}

function colorName(c) {
  return c === "w" ? "White" : "Black";
}

function delay(ms) {
  return new Promise((r) => window.setTimeout(r, ms));
}

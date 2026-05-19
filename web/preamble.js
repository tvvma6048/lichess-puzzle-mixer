import { Chess } from "./vendor/chess.js";

/** Compare positions ignoring halfmove/fullmove counters. */
export function fenKey(fen) {
  return fen.split(" ").slice(0, 4).join(" ");
}

/**
 * Build history entries from a Lichess game PGN up to (and including) targetFen.
 * @returns {Array<{fen:string,san?:string,uci?:string,lastMove?:string[],preamble:boolean,moveIndex:number}>}
 */
function stripPgnComments(pgn) {
  return pgn.replace(/\{[^}]*\}/g, " ");
}

function replayToEntries(replay, verbose, count) {
  const entries = [
    {
      fen: replay.fen(),
      san: undefined,
      uci: undefined,
      lastMove: undefined,
      preamble: true,
      moveIndex: 0,
    },
  ];

  for (let i = 0; i < count; i++) {
    const m = verbose[i];
    const uci = m.from + m.to + (m.promotion || "");
    replay.move(m);
    entries.push({
      fen: replay.fen(),
      san: m.san,
      uci,
      lastMove: [m.from, m.to],
      preamble: true,
      moveIndex: 0,
    });
  }

  return entries;
}

export function buildPreambleFromPgn(pgn, targetFen, plyHint = 0) {
  try {
    const target = fenKey(targetFen);
    const probe = new Chess();
    probe.loadPgn(stripPgnComments(pgn), { sloppy: true });

    const verbose = probe.history({ verbose: true });
    if (verbose.length === 0) {
      return [];
    }

    const headers = probe.header();
    const headerFen = headers?.FEN;
    const replay = headerFen ? new Chess(headerFen) : new Chess();

    if (fenKey(replay.fen()) === target) {
      return [];
    }

    const maxPly = verbose.length;
    const plyCap = plyHint > 0 ? Math.min(plyHint, maxPly) : maxPly;

    for (let i = 0; i < plyCap; i++) {
      const trial = headerFen ? new Chess(headerFen) : new Chess();
      const partial = replayToEntries(trial, verbose, i + 1);
      if (fenKey(partial[partial.length - 1].fen) === target) {
        return partial;
      }
    }

    for (let i = plyCap; i <= maxPly; i++) {
      const trial = headerFen ? new Chess(headerFen) : new Chess();
      const partial = replayToEntries(trial, verbose, i);
      if (fenKey(partial[partial.length - 1].fen) === target) {
        return partial;
      }
    }

    if (plyHint > 0) {
      const trial = headerFen ? new Chess(headerFen) : new Chess();
      return replayToEntries(trial, verbose, Math.min(plyHint, maxPly));
    }

    return [];
  } catch {
    return [];
  }
}

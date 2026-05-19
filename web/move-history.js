const FIGURINE = { K: "♔", Q: "♕", R: "♖", B: "♗", N: "♘" };

/** Convert SAN to Lichess-style figurine notation (piece symbol + squares). */
export function sanToFigurine(san) {
  if (!san) return "";
  if (san.startsWith("O-O")) return san;
  return san.replace(/^([KQRBN])/, (_, p) => FIGURINE[p] ?? p);
}

/**
 * @param {Array<{ san?: string, uci?: string }>} history Positions; index 0 has no san.
 * @param {number} cursor Active position index in history.
 * @param {string} startFen Initial puzzle FEN (before setup).
 */
export function buildMoveRows(history, cursor, startFen) {
  const parts = startFen.split(" ");
  let turn = parts[1];
  let moveNum = parseInt(parts[5], 10) || 1;
  const rows = [];
  let row = { num: moveNum, white: null, black: null };

  for (let i = 1; i < history.length; i++) {
    const entry = { ...history[i], index: i };
    if (turn === "w") {
      if (row.white) {
        rows.push(row);
        moveNum++;
        row = { num: moveNum, white: null, black: null };
      }
      row.white = entry;
    } else {
      row.black = entry;
      rows.push(row);
      moveNum++;
      row = { num: moveNum, white: null, black: null };
    }
    turn = turn === "w" ? "b" : "w";
  }

  if (row.white || row.black) rows.push(row);
  return { rows, cursor };
}

export function renderMoveHistory(container, history, cursor, startFen) {
  if (!container) return;

  if (history.length <= 1) {
    container.replaceChildren();
    container.hidden = true;
    return;
  }

  container.hidden = false;
  const { rows } = buildMoveRows(history, cursor, startFen);
  const frag = document.createDocumentFragment();

  for (const row of rows) {
    const tr = document.createElement("tr");
    tr.className = "move-row";

    const num = document.createElement("th");
    num.className = "move-num";
    num.textContent = `${row.num}.`;
    tr.appendChild(num);

    for (const side of ["white", "black"]) {
      const td = document.createElement("td");
      const entry = row[side];
      if (!entry) {
        td.className = "move-cell move-cell--empty";
        tr.appendChild(td);
        continue;
      }

      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "move-cell";
      btn.dataset.historyIndex = String(entry.index);
      btn.textContent = sanToFigurine(entry.san || entry.uci || "…");
      if (entry.preamble) btn.classList.add("is-preamble");
      if (entry.index === cursor) btn.classList.add("is-active");
      tr.appendChild(td);
      td.appendChild(btn);
    }

    frag.appendChild(tr);
  }

  container.replaceChildren(frag);

  const active = container.querySelector(".is-active");
  active?.scrollIntoView({ block: "nearest" });
}

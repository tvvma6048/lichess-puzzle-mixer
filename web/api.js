export async function fetchStatus() {
  const res = await fetch("/api/status");
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return res.json();
}

export async function fetchDatabase() {
  const res = await fetch("/api/database");
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function deleteDatabase() {
  const res = await fetch("/api/database", { method: "DELETE" });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function importSampleDatabase() {
  const res = await fetch("/api/database/import-sample", { method: "POST" });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function importDatabaseFile(file) {
  const form = new FormData();
  form.append("file", file);
  const res = await fetch("/api/database/import", { method: "POST", body: form });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

/** Returns { ok, pgn, ply } or { ok: false }. Never throws. */
export async function fetchGamePreamble(gameUrl, fen) {
  try {
    const params = new URLSearchParams({ game_url: gameUrl, fen });
    const res = await fetch(`/api/puzzle/preamble?${params}`);
    if (!res.ok) {
      return { ok: false };
    }
    return await res.json();
  } catch {
    return { ok: false };
  }
}

export async function downloadLichessDatabase() {
  const res = await fetch("/api/database/download-lichess", { method: "POST" });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function fetchPuzzleCount(filter) {
  const res = await fetch("/api/puzzle/count", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(filter),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function fetchNextPuzzle(filter) {
  const res = await fetch("/api/puzzle/next", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(filter),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}


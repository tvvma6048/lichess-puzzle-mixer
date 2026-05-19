import {
  deleteDatabase,
  fetchDatabase,
  importSampleDatabase,
  importDatabaseFile,
  downloadLichessDatabase,
} from "./api.js";

export function initDatabasePanel({ onReadyChange } = {}) {
  const statusText = document.getElementById("db-status-text");
  const puzzleCount = document.getElementById("db-puzzle-count");
  const importedAt = document.getElementById("db-imported-at");
  const fileSize = document.getElementById("db-file-size");
  const dbPath = document.getElementById("db-path");
  const progressEl = document.getElementById("db-progress");
  const messageEl = document.getElementById("db-message");
  const topStatus = document.getElementById("status");
  const panelBody = document.getElementById("database-panel-body");
  const toggleBtn = document.getElementById("btn-toggle-database");

  const importSampleBtn = document.getElementById("btn-import-sample");
  const uploadInput = document.getElementById("input-import-csv");
  const downloadBtn = document.getElementById("btn-download-lichess");
  const deleteBtn = document.getElementById("btn-delete-db");

  let pollTimer = null;
  let lastReady = null;
  let detailsExpanded = false;

  const setDetailsCollapsed = (collapsed) => {
    if (!panelBody) return;
    panelBody.classList.toggle("database-panel-body--collapsed", collapsed);
    if (toggleBtn) {
      toggleBtn.setAttribute("aria-expanded", String(!collapsed));
      toggleBtn.textContent = collapsed ? "Show database" : "Hide database";
    }
  };

  const syncPanelVisibility = (ready) => {
    if (!ready) {
      if (toggleBtn) toggleBtn.hidden = true;
      setDetailsCollapsed(false);
      detailsExpanded = false;
      return;
    }
    if (toggleBtn) toggleBtn.hidden = false;
    if (lastReady !== true) {
      detailsExpanded = false;
    }
    setDetailsCollapsed(!detailsExpanded);
  };

  toggleBtn?.addEventListener("click", () => {
    const collapsed = panelBody?.classList.contains("database-panel-body--collapsed");
    detailsExpanded = Boolean(collapsed);
    setDetailsCollapsed(!detailsExpanded);
  });

  const setBusy = (busy) => {
    for (const el of [importSampleBtn, downloadBtn, deleteBtn, uploadInput]) {
      if (el) el.disabled = busy;
    }
  };

  const formatBytes = (n) => {
    if (!n) return "—";
    const units = ["B", "KB", "MB", "GB"];
    let v = n;
    let i = 0;
    while (v >= 1024 && i < units.length - 1) {
      v /= 1024;
      i++;
    }
    return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
  };

  const formatDate = (iso) => {
    if (!iso) return "—";
    try {
      return new Date(iso).toLocaleString();
    } catch {
      return iso;
    }
  };

  const render = (info) => {
    const ready = Boolean(info.db_ready);
    statusText.textContent = ready ? "Ready" : info.import_running ? "Importing…" : "Empty";
    puzzleCount.textContent = ready ? info.puzzle_count.toLocaleString() : "0";
    importedAt.textContent = formatDate(info.imported_at);
    fileSize.textContent = formatBytes(info.db_size_bytes);
    dbPath.textContent = info.db_path || "—";

    if (info.import_running) {
      progressEl.hidden = false;
      progressEl.textContent = info.import_message || "Working…";
      setBusy(true);
    } else {
      progressEl.hidden = true;
      progressEl.textContent = "";
      setBusy(false);
    }

    if (topStatus) {
      if (ready) {
        topStatus.className = "status ok";
        topStatus.innerHTML = `
          <strong>Ready</strong>
          <br />
          <span class="meta">${info.puzzle_count.toLocaleString()} puzzles in database</span>
        `;
      } else if (info.import_running) {
        topStatus.className = "status loading";
        topStatus.textContent = info.import_message || "Import in progress…";
      } else {
        topStatus.className = "status loading";
        topStatus.innerHTML = `
          <strong>No puzzles imported</strong>
          <br />
          <span class="meta">Import a sample or upload a Lichess CSV below</span>
        `;
      }
    }

    syncPanelVisibility(ready);
    lastReady = ready;
    onReadyChange?.(ready);
  };

  const refresh = async () => {
    const info = await fetchDatabase();
    render(info);
    return info;
  };

  const showMessage = (text, isError = false) => {
    messageEl.textContent = text;
    messageEl.className = isError ? "db-message is-error" : "db-message is-ok";
  };

  const stopPoll = () => {
    if (pollTimer !== null) {
      window.clearInterval(pollTimer);
      pollTimer = null;
    }
  };

  const startPoll = () => {
    stopPoll();
    pollTimer = window.setInterval(async () => {
      try {
        const info = await refresh();
        if (!info.import_running) {
          stopPoll();
          if (info.import_error) {
            showMessage(info.import_error, true);
          } else if (info.import_stage === "done") {
            showMessage(`Import finished — ${info.import_rows.toLocaleString()} puzzles`);
          }
        }
      } catch (err) {
        stopPoll();
        showMessage(err.message, true);
        setBusy(false);
      }
    }, 1000);
  };

  const runAction = async (fn) => {
    messageEl.textContent = "";
    messageEl.className = "db-message";
    setBusy(true);
    try {
      const result = await fn();
      await refresh();
      return result;
    } catch (err) {
      showMessage(err.message, true);
      setBusy(false);
      throw err;
    }
  };

  importSampleBtn?.addEventListener("click", () =>
    runAction(async () => {
      const result = await importSampleDatabase();
      showMessage(result.message);
      return result;
    }),
  );

  uploadInput?.addEventListener("change", async () => {
    const file = uploadInput.files?.[0];
    uploadInput.value = "";
    if (!file) return;
    await runAction(async () => {
      const result = await importDatabaseFile(file);
      showMessage(result.message);
      return result;
    });
  });

  downloadBtn?.addEventListener("click", async () => {
    const ok = window.confirm(
      "Download the full Lichess puzzle database (~300MB compressed)?\n\nThis can take a long time and replaces your current database.",
    );
    if (!ok) return;
    messageEl.textContent = "";
    setBusy(true);
    try {
      await downloadLichessDatabase();
      startPoll();
      await refresh();
    } catch (err) {
      showMessage(err.message, true);
      setBusy(false);
    }
  });

  deleteBtn?.addEventListener("click", async () => {
    const ok = window.confirm("Delete all puzzles from the database? This cannot be undone.");
    if (!ok) return;
    await runAction(async () => {
      const result = await deleteDatabase();
      showMessage(result.message || "Database cleared");
      return result;
    });
  });

  return {
    refresh,
    stopPoll,
  };
}

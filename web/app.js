import { initFilters, readFilterFromForm, saveFilterForPlay } from "./filters.js";
import { initPlay } from "./play.js";
import { initDatabasePanel } from "./database.js";

const statusEl = document.getElementById("status");
const viewSetup = document.getElementById("view-setup");
const viewPlay = document.getElementById("view-play");
const filterPanel = document.getElementById("filter-panel");

let playInitialized = false;
let filtersInitialized = false;
let databasePanel = null;

function showView() {
  const isPlay = location.hash === "#/play";
  viewSetup.hidden = isPlay;
  viewPlay.hidden = !isPlay;

  if (isPlay) {
    if (!playInitialized) {
      playInitialized = true;
      initPlay();
    }
  } else {
    playInitialized = false;
    void initSetup();
  }
}

async function initSetup() {
  statusEl.className = "status loading";
  statusEl.textContent = "Checking API…";

  if (!databasePanel) {
    databasePanel = initDatabasePanel({
      onReadyChange: (ready) => {
        if (ready && !filtersInitialized) {
          filtersInitialized = true;
          initFilters((form) => {
            saveFilterForPlay(readFilterFromForm(form));
            location.hash = "#/play";
            showView();
          });
        }
        if (filterPanel) filterPanel.hidden = !ready;
      },
    });
  }

  try {
    await databasePanel.refresh();
  } catch (err) {
    statusEl.className = "status error";
    statusEl.textContent = `API error: ${err.message}`;
    if (filterPanel) filterPanel.hidden = true;
  }
}

window.addEventListener("hashchange", showView);

if (location.hash === "#/play") {
  viewSetup.hidden = true;
  viewPlay.hidden = false;
  initPlay();
} else {
  void initSetup();
}

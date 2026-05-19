import { test, expect, type Page } from "@playwright/test";
import fs from "fs";
import path from "path";
import { execSync } from "child_process";

const outDir = path.join(__dirname, "..", "docs", "images");
const framesDir = path.join(outDir, "_gif-frames");

async function selectInGroup(page: Page, groupIndex: number, themes: string[]) {
  const group = page.getByTestId("theme-group").nth(groupIndex);
  for (let i = 0; i < themes.length; i++) {
    await group.getByTestId("theme-dropdown").nth(i).selectOption(themes[i]);
  }
}

test("capture README screenshots and demo gif", async ({ page }) => {
  fs.mkdirSync(outDir, { recursive: true });
  fs.rmSync(framesDir, { recursive: true, force: true });
  fs.mkdirSync(framesDir, { recursive: true });

  let frame = 0;
  const captureFrame = async () => {
    const name = `frame${String(frame++).padStart(2, "0")}.png`;
    await page.screenshot({ path: path.join(framesDir, name) });
  };

  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 15_000 });
  await expect(page.getByTestId("filter-panel")).toBeVisible();

  await captureFrame();

  await selectInGroup(page, 0, ["fork", "pin"]);
  await page.getByTestId("btn-add-theme-group").click();
  await selectInGroup(page, 1, ["short"]);
  await captureFrame();

  await page.getByTestId("filter-search").click();
  await expect(page.getByTestId("match-count")).toContainText(/\(fork or pin\) and short/i, {
    timeout: 10_000,
  });
  await captureFrame();

  await page.locator(".setup-main").screenshot({
    path: path.join(outDir, "filters.png"),
  });

  await page.getByTestId("btn-start-training").click();
  await expect(page.getByTestId("view-play")).toBeVisible();
  await expect(page.getByTestId("btn-hint")).toBeEnabled({ timeout: 20_000 });
  await expect(page.locator("cg-board")).toBeVisible({ timeout: 10_000 });
  await captureFrame();

  await page.locator(".play-layout").screenshot({
    path: path.join(outDir, "board.png"),
  });

  try {
    execSync(
      `ffmpeg -y -framerate 0.4 -i "${framesDir}/frame%02d.png" -vf "fps=2,scale=900:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" -loop 0 "${path.join(outDir, "demo.gif")}"`,
      { stdio: "pipe" },
    );
  } catch {
    test.info().annotations.push({
      type: "note",
      description: "ffmpeg not available; demo.gif skipped",
    });
  }

  fs.rmSync(framesDir, { recursive: true, force: true });
});

import { test, expect } from "@playwright/test";

test("home page loads static assets and API status UI", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle(/Lichess Puzzle Mixer/);
  await expect(page.locator("h1")).toHaveText("Lichess Puzzle Mixer");
  await expect(page.locator(".subtitle")).toContainText("choose filters");

  const status = page.getByTestId("api-status");
  await expect(status).toHaveClass(/ok/, { timeout: 10_000 });
  await expect(status).toContainText(/puzzles in database/i);
});

test("styles are applied", async ({ page }) => {
  await page.goto("/");
  const main = page.locator(".setup-main");
  const bg = await main.evaluate((el) => getComputedStyle(el).backgroundColor);
  expect(bg).not.toBe("rgba(0, 0, 0, 0)");
});

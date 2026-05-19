import { test, expect } from "@playwright/test";

test("play view loads board after start training", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  await page.getByTestId("btn-start-training").click();

  await expect(page.getByTestId("view-play")).toBeVisible();
  await expect(page.getByTestId("board")).toBeVisible();
  await expect(page.getByTestId("play-status")).not.toHaveText("Loading puzzle…", {
    timeout: 15_000,
  });
  await expect(page.locator("cg-board")).toBeVisible({ timeout: 10_000 });
  await expect(page.getByTestId("play-meta")).toHaveClass(/play-meta--collapsed/);

  const puzzleLink = page.getByTestId("play-lichess-puzzle");
  await expect(puzzleLink).toBeVisible();
  await expect(puzzleLink).toHaveAttribute("href", /lichess\.org\/training\//);
});

test("play rating and themes expand on toggle", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });
  await page.getByTestId("btn-start-training").click();
  await expect(page.getByTestId("btn-hint")).toBeEnabled({ timeout: 15_000 });

  await expect(page.getByTestId("play-meta")).toHaveClass(/play-meta--collapsed/);
  await page.getByTestId("btn-toggle-play-meta").click();
  await expect(page.getByTestId("play-meta")).not.toHaveClass(/play-meta--collapsed/);
  await expect(page.getByTestId("play-rating")).not.toHaveText("…");
});

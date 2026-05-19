import { test, expect } from "@playwright/test";

test("move history and navigation controls", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });
  await page.getByTestId("btn-start-training").click();
  await expect(page.getByTestId("btn-hint")).toBeEnabled({ timeout: 15_000 });

  await expect(page.getByTestId("play-moves")).toBeVisible();
  await expect(page.getByTestId("btn-restart")).toBeDisabled();

  await page.getByTestId("btn-hint").click();
  await page.getByTestId("btn-solution").click();
  await expect(page.getByTestId("play-status")).toHaveText("Solution", { timeout: 15_000 });
  await expect(page.getByTestId("play-moves").locator(".move-cell")).not.toHaveCount(0);

  await expect(page.getByTestId("btn-prev")).toBeEnabled();
  await page.getByTestId("btn-prev").click();
  await expect(page.getByTestId("play-status")).not.toHaveText("Solution");

  await page.getByTestId("btn-restart").click();
  await expect(page.getByTestId("btn-hint")).toBeEnabled();
  await expect(page.getByTestId("play-status")).toContainText(/you play/i);
});

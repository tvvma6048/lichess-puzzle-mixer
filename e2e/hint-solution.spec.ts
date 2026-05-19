import { test, expect } from "@playwright/test";

test("hint and view solution work during a puzzle", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  await page.getByTestId("btn-start-training").click();
  await expect(page.getByTestId("view-play")).toBeVisible();
  await expect(page.getByTestId("btn-hint")).toBeEnabled({ timeout: 15_000 });

  await page.getByTestId("btn-hint").click();
  await expect(page.getByTestId("play-status")).toContainText(/move this piece/i);
  await expect(page.locator("cg-board square.hint-from")).toBeVisible();

  await page.getByTestId("btn-hint").click();
  await expect(page.getByTestId("play-status")).toContainText(/move here/i);
  await expect(page.locator("cg-board square.move-dest, cg-board square.oc.move-dest")).toBeVisible();

  await page.getByTestId("btn-solution").click();
  await expect(page.getByTestId("play-status")).toHaveText("Solution", { timeout: 15_000 });
  await expect(page.getByTestId("play-feedback")).toHaveText("Solution shown");
  await expect(page.getByTestId("btn-next")).toBeEnabled();
  await expect(page.getByTestId("btn-hint")).toBeDisabled();
});

import { test, expect } from "@playwright/test";

test("database panel collapses when ready, expands on toggle", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("database-panel")).toBeVisible();
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  const toggle = page.getByTestId("btn-toggle-database");
  const body = page.getByTestId("database-panel-body");

  await expect(toggle).toBeVisible();
  await expect(toggle).toHaveText("Show database");
  await expect(body).toHaveClass(/database-panel-body--collapsed/);
  await expect(page.getByTestId("db-stats")).toBeHidden();

  await toggle.click();
  await expect(toggle).toHaveText("Hide database");
  await expect(body).not.toHaveClass(/database-panel-body--collapsed/);
  await expect(page.getByTestId("db-path")).not.toHaveText("—");
  await expect(page.getByTestId("db-puzzle-count")).not.toHaveText("0");
  await expect(page.getByTestId("btn-import-sample")).toBeEnabled();
  await expect(page.getByTestId("btn-delete-db")).toBeEnabled();

  await toggle.click();
  await expect(body).toHaveClass(/database-panel-body--collapsed/);
});

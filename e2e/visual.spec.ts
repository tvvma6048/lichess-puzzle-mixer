import { test, expect } from "@playwright/test";

/**
 * Visual regression: compares screenshots to committed baselines in
 * e2e/visual.spec.ts-snapshots/. Run once after intentional UI changes:
 *   cd e2e && npx playwright test visual --update-snapshots
 */
test("home page success state matches baseline", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, {
    timeout: 10_000,
  });

  // Screenshot theme filters only — database panel collapse/expand changes setup-main height on CI.
  await expect(page.getByTestId("filter-panel")).toBeVisible();
  await expect(page.getByTestId("theme-groups")).toBeVisible();

  await expect(page.getByTestId("filter-panel")).toHaveScreenshot("home-success.png", {
    maxDiffPixelRatio: 0.02,
  });
});

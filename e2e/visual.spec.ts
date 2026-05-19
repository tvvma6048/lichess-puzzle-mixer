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

  await expect(page.locator(".setup-main")).toHaveScreenshot("home-success.png", {
    maxDiffPixelRatio: 0.02,
  });
});

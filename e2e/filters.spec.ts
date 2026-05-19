import { test, expect, type Page } from "@playwright/test";

async function selectInGroup(page: Page, groupIndex: number, themes: string[]) {
  const group = page.getByTestId("theme-group").nth(groupIndex);
  for (let i = 0; i < themes.length; i++) {
    await group.getByTestId("theme-dropdown").nth(i).selectOption(themes[i]);
  }
}

async function addThemeGroup(page: Page) {
  await page.getByTestId("btn-add-theme-group").click();
}

test("theme groups are always visible", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  await expect(page.getByTestId("theme-filters")).toBeVisible();
  await expect(page.getByTestId("theme-group")).toHaveCount(1);
  await expect(page.getByTestId("theme-dropdown")).toHaveCount(1);
});

test("selecting a theme adds another dropdown in the group", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  const group = page.getByTestId("theme-group").first();
  await group.getByTestId("theme-dropdown").first().selectOption("fork");
  await expect(group.getByTestId("theme-dropdown")).toHaveCount(2);
  await expect(group.getByTestId("theme-dropdown").nth(1)).toHaveValue("");
});

test("filter preview shows fork match count", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  await selectInGroup(page, 0, ["fork"]);
  await page.getByTestId("filter-search").click();

  await expect(page.getByTestId("match-count")).toContainText(/puzzles match/i);
  await expect(page.getByTestId("match-count")).toContainText(/fork/i);

  const count = parseMatchCount(await page.getByTestId("match-count").textContent());
  expect(count).toBeGreaterThan(0);
  expect(count).toBeLessThan(500);
});

test("fork AND pin as separate groups", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  await selectInGroup(page, 0, ["fork"]);
  await addThemeGroup(page);
  await selectInGroup(page, 1, ["pin"]);
  await page.getByTestId("filter-search").click();

  await expect(page.getByTestId("match-count")).toContainText(/fork and pin/i);
  await expect(page.getByTestId("match-count")).toContainText(/1 puzzles match/);
});

test("(fork or pin) and short", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("api-status")).toHaveClass(/ok/, { timeout: 10_000 });

  await selectInGroup(page, 0, ["fork", "pin"]);
  await addThemeGroup(page);
  await selectInGroup(page, 1, ["short"]);
  await page.getByTestId("filter-search").click();

  await expect(page.getByTestId("match-count")).toContainText(/\(fork or pin\) and short/i);

  expect(parseMatchCount(await page.getByTestId("match-count").textContent())).toBeGreaterThan(0);
  expect(parseMatchCount(await page.getByTestId("match-count").textContent())).toBeLessThan(500);
});

function parseMatchCount(text: string | null): number {
  const m = text?.replace(/,/g, "").match(/^(\d+)/);
  return parseInt(m?.[1] ?? "0", 10);
}

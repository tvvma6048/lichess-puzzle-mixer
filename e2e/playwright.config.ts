import { defineConfig } from "@playwright/test";
import path from "path";

const root = path.join(__dirname, "..");
const binary = path.join(root, "bin", "lichess-puzzle-mixer");
const dataDir = path.join(root, ".e2e-data");

export default defineConfig({
  testDir: ".",
  testIgnore: process.env.CAPTURE_README_ASSETS
    ? []
    : ["**/capture-readme-assets.spec.ts"],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL: "http://127.0.0.1:7799",
    trace: "on-first-retry",
    viewport: { width: 900, height: 700 },
  },
  expect: {
    toHaveScreenshot: {
      maxDiffPixelRatio: 0.02,
    },
  },
  snapshotPathTemplate: "{testDir}/{testFilePath}-snapshots/{arg}{ext}",
  webServer: {
    command: `${binary} --no-browser --data-dir ${dataDir} --port 7799`,
    url: "http://127.0.0.1:7799/api/status",
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});

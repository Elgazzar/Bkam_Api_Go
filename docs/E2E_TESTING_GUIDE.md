# 🧪 End-to-End Testing Guide (`tester-army/e2e`)

This document provides a comprehensive reference for the end-to-end (E2E) testing framework integrated across the **API Gateway** (`api/`) and the **Scraper Service** (`serverless-scraper/`).

---

## 📌 1. What is `tester-army/e2e`?

[`tester-army/e2e`](https://github.com/tester-army/e2e) is an open-source, next-generation end-to-end testing framework developed by TesterArmy. It combines:

1. **Deterministic Testing**: Fast, reliable assertions using standard locators, HTTP `fetch`, and Playwright browser engines (`@e2e-dev/web`).
2. **Standard Schema Matching**: Built-in `expect(payload).toMatchSchema(schema)` supporting Zod, Valibot, and ArkType to validate API contracts at runtime.
3. **Automated Server Lifecycle**: Automatically spins up and shuts down local servers (`app.command`), or seamlessly attaches to existing servers already running on the port (`reuseExisting: true`).
4. **Agentic Capabilities (Opt-in)**: Allows tests to describe high-level goals in natural language (`agent.act()`, `agent.assert()`) powered by AI models, accompanied by a 0-cost replay cache (`.e2e/cache/`) that records passing actions and replays them deterministically without spending tokens on subsequent runs.
5. **Local-First & Zero-Cost Default**: Tests that do not contain `agent` steps require no model keys, subscriptions, or external network calls.

---

## ⚙️ 2. How Do the Tests Run?

### Does it run the actual API?
**YES, 100% actual API.**

When you run `npm run test:e2e` inside `api/`:
1. **Server Orchestration**: The runner checks [`api/e2e.config.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/e2e.config.ts). If the Fastify server is not already active on port 3000, it spawns `node node_modules/tsx/dist/cli.mjs src/app.ts` as a managed background child process.
2. **Health Readiness**: The runner polls `http://127.0.0.1:3000/v1/health` until the Fastify server responds.
3. **Live HTTP Execution**: Tests in [`api/tests/e2e/api.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/api.e2e.ts) send real HTTP `fetch` requests over loopback TCP (`http://127.0.0.1:3000`).
4. **Live Browser Navigation**: Tests in [`api/tests/e2e/browser.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/browser.e2e.ts) launch an actual Chromium browser (via Playwright) and physically navigate to `http://127.0.0.1:3000/v1/health` to verify page rendering.
5. **Database / Emulator Interaction**:
   - Routes that read from memory or cache (`/v1/health`, `/v1/catalog`, `/v1/rates`, and unauthenticated requests) respond in milliseconds.
   - For database-dependent routes (`/v1/gold`, `/v1/gas`), tests are configured with offline resilience: if the Firebase emulator is offline, requests gracefully fail-safe with HTTP 503 or timeout handling rather than freezing. When the Firebase emulator is running (e.g., via `run-functional-tests.ps1`), tests execute against live seeded data.

---

### Does it run the actual Scraper?
**The browser engine and DOM extraction are 100% REAL; upstream bank network calls are intercepted for reliability.**

When you run `npm run test:e2e` inside `serverless-scraper/`:
1. **Real Browser Engine**: It launches a genuine Chromium browser engine via `@e2e-dev/web`. The page renders real HTML, executes DOM rendering, computes layout styles, and evaluates JavaScript.
2. **Network Interception (`browser.route`)**: Real Egyptian bank websites (such as NBE, CIB, CBE, Banque Misr) frequently deploy Cloudflare WAFs, CAPTCHAs, or experience downtime. Hitting live bank sites directly on every test run causes rate limits and test flakiness.
3. **Deterministic Verification**: In [`serverless-scraper/tests/e2e/scraper-browser.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/tests/e2e/scraper-browser.e2e.ts), `browser.route(...)` intercepts bank URLs and serves realistic HTML tables and feeds. This proves that the scraper's selectors, locators, and numeric parsing work correctly against expected markup without depending on third-party bank uptime.

---

## 📂 3. Currently Configured Tests

### In `api/`
Configuration: [`api/e2e.config.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/e2e.config.ts)

| Test File | Test Case | What It Verifies |
| :--- | :--- | :--- |
| [`api.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/api.e2e.ts) | `GET /v1/health` | HTTP 200, validates payload against `HealthResponseSchema` (`status` is `"ok"` or `"degraded"`, valid ISO timestamp). |
| [`api.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/api.e2e.ts) | `GET /v1/rates` (No Key) | HTTP 401 Unauthorized, validates structured error schema (`code === "UNAUTHORIZED"`). |
| [`api.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/api.e2e.ts) | `GET /v1/rates` (With Key) | Sends `x-api-key: local-dev-key`, verifies valid snapshot acceptance (200 or 503 during cache warm-up). |
| [`api.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/api.e2e.ts) | `GET /v1/catalog` | HTTP 200, verifies `banks` and `currencies` catalog arrays. |
| [`api.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/api.e2e.ts) | `GET /v1/gold` | Authenticated request to gold rates with offline emulator resilience. |
| [`api.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/api.e2e.ts) | `GET /v1/gas` | Authenticated request to domestic fuel rates with offline emulator resilience. |
| [`browser.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/api/tests/e2e/browser.e2e.ts) | Browser Health Check | Headless Chromium navigates to `http://127.0.0.1:3000/v1/health`, verifies DOM content. |

### In `serverless-scraper/`
Configuration: [`serverless-scraper/e2e.config.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/e2e.config.ts)

| Test File | Test Case | What It Verifies |
| :--- | :--- | :--- |
| [`scraper-browser.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/tests/e2e/scraper-browser.e2e.ts) | Table Rate Extraction | Navigates to a simulated exchange rates table; extracts `USD` buy/sell rates using scoped CSS locators; verifies numeric parsing and total row counts. |
| [`scraper-browser.e2e.ts`](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/serverless-scraper/tests/e2e/scraper-browser.e2e.ts) | Dynamic DOM Retrieval | Verifies accessible role queries (`screen.getByRole("status")`, `screen.getByRole("button")`) for client-side rendered rate widgets. |

---

## ✍️ 4. How to Add a New Test

All `e2e` tests must be placed under a `tests/e2e/` directory and end with the extension `.e2e.ts` (e.g. `tests/e2e/<feature>.e2e.ts`).

### Example A: Adding a New API Endpoint Test (in `api/`)

Create or update a file in `api/tests/e2e/`:

```ts
// api/tests/e2e/stocks.e2e.ts
// Import test runner and assertions from e2e
import { test, expect } from "e2e";
// Import Zod for contract schema verification
import { z } from "zod";

// Define the expected contract schema
const StockResponseSchema = z.object({
  stocks: z.array(
    z.object({
      symbol: z.string(),
      price: z.number().positive(),
    })
  ).optional(),
});

test("GET /v1/stocks returns stocks list with valid key", async ({ app }) => {
  // Resolve API key
  const apiKey = process.env.API_KEY ?? "local-dev-key";

  // Make the HTTP request against the live server
  const response = await fetch(new URL("/v1/stocks", app.baseUrl), {
    headers: { "x-api-key": apiKey },
  });

  // Verify HTTP status code
  expect([200, 503]).toContain(response.status);

  if (response.status === 200) {
    const json = await response.json();
    // Validate schema
    expect(json).toMatchSchema(StockResponseSchema);
  }
});
```

---

### Example B: Adding a New Scraper Browser Test (in `serverless-scraper/`)

Create or update a file in `serverless-scraper/tests/e2e/`:

```ts
// serverless-scraper/tests/e2e/gold-scraper.e2e.ts
// Import test with web engine fixture
import { test } from "@e2e-dev/web";
// Import expect assertions
import { expect } from "e2e";

test("extracts 21K gold prices from supplier page", async ({ browser }) => {
  // 1. Mock the supplier page markup
  const mockHtml = `
    <html>
      <body>
        <div class="gold-card" data-karat="21">
          <span class="label">Gold 21K</span>
          <span class="buy-price">3150.00</span>
          <span class="sell-price">3170.00</span>
        </div>
      </body>
    </html>
  `;

  // 2. Intercept the route to return our mock markup
  await browser.route("http://localhost:8080/gold-supplier", (route) => {
    route.fulfill({ status: 200, contentType: "text/html", body: mockHtml });
  });

  // 3. Navigate the real Chromium browser
  await browser.goto("http://localhost:8080/gold-supplier");

  // 4. Extract data using locators
  const buyCell = browser.locator(".gold-card[data-karat='21'] .buy-price");
  await expect(buyCell).toHaveText("3150.00");

  const buyPriceNum = parseFloat((await buyCell.textContent()) ?? "0");
  expect(buyPriceNum).toBe(3150.00);
});
```

---

### Example C: Adding an AI Agentic Test (Optional)

If an AI subscription or model API key (e.g. `OPENAI_API_KEY`, `AI_GATEWAY_API_KEY`) is configured, you can drive tests via natural language goals:

```ts
// api/tests/e2e/agent.e2e.ts
import { test } from "@e2e-dev/web";
import { expect } from "e2e";

test("agent verifies status indicator on page", async ({ app, agent, screen }) => {
  await app.open("/v1/health");

  // Natural language assertion driven by the agent
  await agent.assert("the response shows an ok or degraded status indicator");
  await expect(screen.getByText("status")).toBeVisible();
});
```

---

## 🚀 5. How to Run the Tests

### Run API E2E Tests
```powershell
cd api
npm run test:e2e
```
*Discovers and runs all `api/tests/e2e/*.e2e.ts` tests against the Fastify server.*

### Run Scraper E2E Tests
```powershell
cd serverless-scraper
npm run test:e2e
```
*Discovers and runs all `serverless-scraper/tests/e2e/*.e2e.ts` tests using Chromium.*

### Useful CLI Flags
You can append flags directly to `npx e2e run`:
- **Run a specific test file**:
  ```powershell
  npx e2e run tests/e2e/api.e2e.ts
  ```
- **Run in headed browser mode** (watch Chromium open on your screen):
  ```powershell
  npx e2e run --headed
  ```
- **Inspect traces on failure**:
  When a test fails, `e2e` generates a detailed step trace under `.e2e/results/<test-name>/trace.md` showing DOM snapshots, network requests, and timings.

---

## 🔄 6. Running E2E Tests Against the Firestore Emulator and Scraper

To run **full functional end-to-end testing** where the Scraper seeds data into the local Firestore Emulator and the API serves it to `tester-army/e2e` and Vitest:

### Unified Test Runner: `run-functional-tests.ps1`

Run the all-in-one automation script from the repository root:

```powershell
# Run the complete functional test suite using imported emulator seed data
powershell -ExecutionPolicy Bypass -File .\run-functional-tests.ps1 -SkipLiveScrape -AutoCleanup

# Or run with live bank scraping to seed freshly scraped data
powershell -ExecutionPolicy Bypass -File .\run-functional-tests.ps1 -AutoCleanup
```

### What This Pipeline Does:
1. **Bootstraps Firestore Emulator**: Starts `127.0.0.1:8085` and imports `./emulator-data` (28 banks, gold, gas, stocks, and historical archives).
2. **Executes Scraper E2E Suite**: Runs `serverless-scraper` browser tests via `tester-army/e2e` verifying Chromium table extractions.
3. **Persists Seed Data**: Exports the updated Firestore state to disk via `npx firebase emulators:export`.
4. **Boots Live Fastify API**: Starts `api/src/app.ts` attached to `FIRESTORE_EMULATOR_HOST=127.0.0.1:8085` and verifies health readiness.
5. **Executes API E2E Suite (`tester-army/e2e`)**: Runs 10 tests in `api/tests/e2e/` (health, rates, gold, gas, catalog, browser) against the live API and Firestore Emulator in under 3 seconds.
6. **Executes Vitest Functional Suite**: Runs 21 integration/cache tests in `api/tests/functional/` and generates an interactive HTML test report.
7. **Clean Teardown**: When `-AutoCleanup` is provided, automatically terminates background processes and frees all port bindings.

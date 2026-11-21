# Automation Hub frontend

React, TypeScript and MUI, built with [Vite](https://vite.dev). Use yarn (the repo has a `yarn.lock`).

## Commands

Run them in this directory.

| Command | What it does |
|---|---|
| `yarn dev` | Starts the Vite dev server on [http://localhost:3000](http://localhost:3000) and proxies `/api`, `/auth`, `/upload` and `/data` to the Go master (`HUB_BACKEND`, default `http://localhost:8002`). |
| `yarn build` | Production build into `build/`. `BUILD_PATH=../endpoints/web/data/ yarn build` writes the build where the Go master embeds it. |
| `yarn typecheck` | Runs `tsc --noEmit`. |
| `yarn lint` | Runs ESLint with the flat config `eslint.config.js`. |
| `yarn test` | Runs the Vitest suite once (jsdom, Testing Library). |
| `yarn test:coverage` | Runs the suite with v8 coverage (text and `coverage/coverage-summary.json`). |

`yarn start` is an alias of `yarn dev`, and `yarn preview` serves a finished build.

## Size limits

`max-lines` 200, `max-lines-per-function` 40, `complexity` 10 and `max-params` 3 are errors for all of
`src`. `src/shared-theme/**` and `src/theme/**` are exempt: they come from a template, and their other
lint findings are warnings.

Tests are the one exception: `max-lines` and `max-lines-per-function` are off for `*.test.ts(x)` and
`src/testing/**`, because a `describe` block holds many cases and the kit builds fixtures. Every other rule
(`complexity`, `max-params`, types, hooks) still applies to them.

## Test kit

`src/testing/` is the one kit the tests share: `fakeApi` (an `ApiClient` that answers by `"METHOD /url"`
and records the calls), `FakeSse` and `fakeSources` (the hub event stream and the streams hooks open
themselves), `fakeAuthClient`, the hub-state seed `SeedHub` and `renderWithProviders` / `wrapperOf`
(the providers and router of the app over fakes). A page keeps only its route tree and fixtures
(`*Harness.tsx`, `*Fixtures.ts`). The kit is excluded from the coverage report.

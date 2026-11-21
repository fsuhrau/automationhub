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

`max-lines` 200, `max-lines-per-function` 40, `complexity` 10 and `max-params` 3 are errors for the
globs in the `refactored` list at the top of `eslint.config.js`, and warnings everywhere else. A package
that refactors a folder adds its glob (for example `'src/api/**'`) to that list.
`src/shared-theme/**` and `src/theme/**` are exempt.

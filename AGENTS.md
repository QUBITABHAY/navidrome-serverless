# Repository Guidelines

## Project structure

- `api/index.go` is the Vercel entry point. HTTP routes live in `server/`, and CLI commands live in `cmd/`. The `scan-r2` command uses `core/storage/r2/` to sync the music library.
- `db/postgres/` holds the PostgreSQL schema and queries. After editing either, run `sqlc generate` from the repository root to refresh the Go bindings in `persistence/postgres/pgdb/`.
- `ui/src/` contains the React player. Edit static assets in `ui/public/`. The Go binary embeds the compiled bundle from `ui/build/`.

## Build and local development

Use the Go version in `go.mod` and the Node.js version in `.nvmrc`. Run these commands from the repository root.

- `make setup` installs Go and UI dependencies, golangci-lint, and Git hooks.
- `make dev` starts the Go server and Vite with hot reload. Open `http://localhost:4533` for the player.
- `make build` builds the player and compiles the Navidrome binary.

## Coding style and naming

Indent Go with tabs and format imports with `goimports`. UI source uses two spaces, single quotes, and omits trailing semicolons. Use PascalCase for components such as `AboutDialog.jsx` and camelCase for hooks such as `useCurrentTheme.js`. Run `make lintall` to check Go with golangci-lint and UI code with Prettier and ESLint.

## Testing

Keep Go tests beside the package code in `*_test.go`. Packages use Go’s `testing` package and Ginkgo/Gomega suites. Shared mocks and media fixtures live in `tests/`. Place Vitest files beside their UI modules, for example `AboutDialog.test.jsx`.

Run `make test PKG=./persistence/postgres` for a focused check. Set `TEST_POSTGRES_URL` to a disposable database to include the live test, which writes library records. Use `make test-js` for UI tests or `make testall` for both stacks and translation checks. CI reports Go statement coverage without a fixed minimum.

## Commits and pull requests

Use scoped commit subjects such as `fix(streaming): handle expired URLs` and sign off each commit with `git commit --signoff`. Open or find the corresponding issue and name your branch after its title and number, for example `adding-docs/834`.

Follow the PR template. Explain what changed and why, link the issue with `Closes #123`, and give test steps with expected results. Include screenshots or a demo for visual changes.

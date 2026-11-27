# Serverless cleanup and pull request review

## What changed from upstream

The comparison uses upstream `9e3deb4330b346ab0e78aa4e48605d074a6bcee3` and fork head `d0532c3`. The initial import, `70c7986`, matched every upstream blob it copied. It omitted the base Compose example and `tests/navidrome-test.toml`. By `d0532c3`, the fork had added 88 files and modified 20. [Upstream snapshot](https://github.com/navidrome/navidrome/tree/9e3deb4330b346ab0e78aa4e48605d074a6bcee3) · [Fork changes since import](https://github.com/QUBITABHAY/navidrome-serverless/compare/70c7986...d0532c3)

- Vercel routes requests to `api/index.go`. The handler initializes the app once per instance, uses `/tmp` for writable state, and mounts the native, Subsonic and public routers plus the embedded player.
- With `DATABASE_URL` set, the routers use PostgreSQL through pgx and sqlc. The shared router constructors still use scanner, artwork, transcoding and plugin types. SQLite also serves local development and the handler’s fallback.
- The R2 adapter supplies object storage and signed playback URLs. A daily GitHub Actions job calls the scan webhook and falls back to the tagged Go CLI. Both entry points use the same sync implementation.

## Cleanup plan and result

1. Remove the standalone release targets and their callers together. This removes container publishing, cross-compilation, Windows installers, Linux packages, Compose, Kubernetes and OS service examples. The developer container has its own Dockerfile.
2. Retire upstream-only POEditor jobs and generated API cache markers. Keep translation validation, source translations and the embedded player. Disable production source maps so rebuilding locally does not embed them in the Go binary.
3. Run CI on `main` and its PRs. Build both entry points and test the handler and datastore against disposable PostgreSQL. Restore the upstream test fixture, repair three incompatible UI development dependencies, and use a Go 1.27-compatible linter.

## Open pull requests reviewed on 2026-11-27

All 20 PRs change dependency manifests and lockfiles. Nineteen had a successful Vercel check. Vercel embeds the checked-in player, so a preview alone does not exercise a new UI dependency. The UI checks below used isolated copies of the repaired toolchain.

| Disposition | Evidence and next step | PR |
| --- | --- | --- |
| Block | `deluan/rest` v1 makes `Repository` generic. The handler fails to compile at `model/datastore.go:20`. Migrate the datastore and native API callers together. | [#28](https://github.com/QUBITABHAY/navidrome-serverless/pull/28) |
| Hold for migration | The refresh plugin requires ESLint 9 or 10. The player uses ESLint 8. | [#22](https://github.com/QUBITABHAY/navidrome-serverless/pull/22) |
| Hold for migration | JSON Forms React requires core 3.8.0. Core and the Material renderers are still on 2.5.2. | [#30](https://github.com/QUBITABHAY/navidrome-serverless/pull/30) |
| Hold for migration | The two v5 providers install a second `ra-core` that requires React 18/19. The React Admin upgrade also requires React 18/19, while the player uses React 17 and the v3 Redux APIs. | [#34](https://github.com/QUBITABHAY/navidrome-serverless/pull/34), [#35](https://github.com/QUBITABHAY/navidrome-serverless/pull/35), [#39](https://github.com/QUBITABHAY/navidrome-serverless/pull/39) |
| Hold for drag-and-drop verification | The backend installs `dnd-core` 16 alongside the provider’s v14 core. Upgrade the pair or verify a real song drag and drop before merging this separately. | [#38](https://github.com/QUBITABHAY/navidrome-serverless/pull/38) |
| Rebase onto the repaired UI toolchain | The proposed Prettier passed the source formatting check. The Redux Saga update passed all 761 player tests. Rebuild the embedded player for the Saga runtime change. | [#31](https://github.com/QUBITABHAY/navidrome-serverless/pull/31), [#32](https://github.com/QUBITABHAY/navidrome-serverless/pull/32) |
| Consolidate | Both lockfiles resolve Vitest and its coverage provider to 5.0.3. That pair passed all 761 tests with coverage enabled. One update can cover both packages. | [#33](https://github.com/QUBITABHAY/navidrome-serverless/pull/33), [#36](https://github.com/QUBITABHAY/navidrome-serverless/pull/36) |
| Ready for CI | Focused tests passed separately for Gomega, Ginkgo, doublestar, httprate and x/term, including the relevant CLI, playlist, configuration or server suites. | [#24](https://github.com/QUBITABHAY/navidrome-serverless/pull/24), [#40](https://github.com/QUBITABHAY/navidrome-serverless/pull/40), [#41](https://github.com/QUBITABHAY/navidrome-serverless/pull/41), [#42](https://github.com/QUBITABHAY/navidrome-serverless/pull/42), [#43](https://github.com/QUBITABHAY/navidrome-serverless/pull/43) |
| Consolidate | Each AWS SDK diff passed the API, R2 and datastore tests and produced a signed URL with the expected object path and expiry. The S3 and config updates already include the proposed core version, and config includes the credentials update. Recheck the combined dependency graph. | [#44](https://github.com/QUBITABHAY/navidrome-serverless/pull/44), [#45](https://github.com/QUBITABHAY/navidrome-serverless/pull/45), [#46](https://github.com/QUBITABHAY/navidrome-serverless/pull/46), [#47](https://github.com/QUBITABHAY/navidrome-serverless/pull/47) |

## Verification

The tagged Go suite passed across 83 packages. The handler and CLI compiled, and the handler and datastore tests passed against a disposable PostgreSQL 16 instance. A clean npm install, 761 player tests, lint, formatting, and the production build passed. The rebuilt player contained zero source maps. A separate `make clean` check retained all 36 tracked player files. Actionlint, migration checks, translation validation, and Go lint also passed.

Verification used local builds and isolated test data. Production Vercel deployment and playback from a live R2 bucket were outside this run.

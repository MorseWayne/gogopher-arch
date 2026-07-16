# Repository guidance

## Invariants

- Treat every published directory under `content/learning/releases/` as immutable. Create a new release ID instead of editing or deleting historical releases.
- Keep database migrations forward-only. Do not run down migrations or delete Learning audit data such as Attempt, Submission, Evidence, Snapshot, ReviewItem, or outbox records.
- Keep the base Compose topology loopback-only and expose only the Web entrypoint; development ports belong in `docker-compose.dev.yml`.

## Validation

- Run the smallest relevant checks first.
- For Go changes, run `go test ./...` and `go vet ./...` when the change scope warrants the full suite.
- For Web changes, run `npm test --prefix web -- --run` and `npm run build --prefix web`.
- For Compose changes, run `./scripts/check-compose-exposure.sh`.
- For Learning release changes, run the content validator and release verifier documented in `README.md`.
- Reserve `npm run e2e:compose --prefix web` for cross-service, persistence, release, or end-to-end workflow changes.

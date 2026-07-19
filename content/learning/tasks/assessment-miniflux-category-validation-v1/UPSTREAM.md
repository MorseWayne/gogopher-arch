# Upstream provenance

- Project: Miniflux v2
- Release: v2.3.2
- Commit: `51f2e0d8199ea8fa305081f6e175bba64b0ef94b`
- Source: <https://github.com/miniflux/v2/tree/51f2e0d8199ea8fa305081f6e175bba64b0ef94b>
- License: Apache-2.0; see `LICENSE`

`internal/model/category.go` retains the upstream SPDX notice and structure. `internal/validator/category.go` is a training modification: the concrete PostgreSQL storage dependency is replaced by the smallest consumer-owned interface so the patch can be tested without a database. The title normalization behavior is the learner change and is not represented as an upstream contribution.

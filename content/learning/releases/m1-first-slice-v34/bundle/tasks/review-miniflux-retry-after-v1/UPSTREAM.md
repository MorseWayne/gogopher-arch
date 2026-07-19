# Upstream provenance

- Project: Miniflux v2
- Release: v2.3.2
- Commit: `51f2e0d8199ea8fa305081f6e175bba64b0ef94b`
- Source: <https://github.com/miniflux/v2/tree/51f2e0d8199ea8fa305081f6e175bba64b0ef94b>
- License: Apache-2.0; see `LICENSE`

The fetcher type, 429 check, integer/HTTP-date parsing shape, package paths and SPDX notice come from the fixed upstream source. This training modification injects the clock and caller cap, and extracts the handler call into a small seam so it can be verified without PostgreSQL or public feeds.

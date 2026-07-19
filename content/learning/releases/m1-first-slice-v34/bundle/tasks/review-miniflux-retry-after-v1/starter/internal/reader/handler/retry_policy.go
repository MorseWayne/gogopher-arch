// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
// Extracted training seam for the fixed RefreshFeed call site; see UPSTREAM.md.

package handler

import "time"

type RateLimitResponse interface {
	IsRateLimited() bool
	ParseRetryDelay(now time.Time, maximum time.Duration) time.Duration
}

func RateLimitDelay(response RateLimitResponse, now time.Time, maximum time.Duration) time.Duration {
	// TODO: only a rate-limited response may influence the next refresh.
	return 0
}

// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
// Modified for the GoGopher Arch training seam; see UPSTREAM.md.

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"net/http"
	"time"
)

type ResponseHandler struct {
	httpResponse *http.Response
}

func NewResponseHandler(httpResponse *http.Response) *ResponseHandler {
	return &ResponseHandler{httpResponse: httpResponse}
}

func (r *ResponseHandler) ParseRetryDelay(now time.Time, maximum time.Duration) time.Duration {
	// TODO: parse positive delta-seconds or HTTP-date relative to now and cap it.
	return 0
}

func (r *ResponseHandler) IsRateLimited() bool {
	return r.httpResponse != nil && r.httpResponse.StatusCode == http.StatusTooManyRequests
}

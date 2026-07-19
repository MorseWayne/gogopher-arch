// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
// Modified for the GoGopher Arch training seam; see UPSTREAM.md.

package validator // import "miniflux.app/v2/internal/validator"

import (
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
)

const maxCategoryTitleRunes = 100

type CategoryStore interface {
	CategoryTitleExists(userID int64, title string) bool
	AnotherCategoryExists(userID, categoryID int64, title string) bool
}

func ValidateCategoryCreation(store CategoryStore, userID int64, request *model.CategoryCreationRequest) *locale.LocalizedError {
	// TODO: normalize, validate, check duplicates, and retain the normalized title.
	return nil
}

func ValidateCategoryModification(store CategoryStore, userID, categoryID int64, request *model.CategoryModificationRequest) *locale.LocalizedError {
	// TODO: leave nil titles untouched; otherwise apply the same create policy.
	return nil
}

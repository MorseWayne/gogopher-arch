package validator

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

type smokeCategoryStore struct{ title string }

func (s *smokeCategoryStore) CategoryTitleExists(_ int64, title string) bool {
	s.title = title
	return false
}
func (s *smokeCategoryStore) AnotherCategoryExists(_, _ int64, title string) bool {
	s.title = title
	return false
}

func TestCategoryValidationSmoke(t *testing.T) {
	store := &smokeCategoryStore{}
	request := &model.CategoryCreationRequest{Title: "  Engineering  "}
	if err := ValidateCategoryCreation(store, 7, request); err != nil {
		t.Fatal(err)
	}
	if request.Title != "Engineering" || store.title != "Engineering" {
		t.Fatalf("request=%q store=%q", request.Title, store.title)
	}
}

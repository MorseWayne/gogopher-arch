package validator

import (
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

type contractCategoryStore struct {
	title           string
	createDuplicate bool
	updateDuplicate bool
	createCalls     int
	updateCalls     int
}

func (s *contractCategoryStore) CategoryTitleExists(_ int64, title string) bool {
	s.title = title
	s.createCalls++
	return s.createDuplicate
}
func (s *contractCategoryStore) AnotherCategoryExists(_, _ int64, title string) bool {
	s.title = title
	s.updateCalls++
	return s.updateDuplicate
}

func TestCategoryNormalizationContract(t *testing.T) {
	tests := []struct {
		name, title, key, normalized string
		duplicate                    bool
	}{
		{"trim unicode whitespace", "\u3000News\u00a0", "", "News", false},
		{"blank", " \t\n", "error.title_required", "", false},
		{"one hundred runes", strings.Repeat("界", 100), "", strings.Repeat("界", 100), false},
		{"too long", strings.Repeat("界", 101), "error.title_too_long", "", false},
		{"normalized duplicate", " News ", "error.category_already_exists", "News", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &contractCategoryStore{createDuplicate: tc.duplicate}
			request := &model.CategoryCreationRequest{Title: tc.title, HideGlobally: true}
			err := ValidateCategoryCreation(store, 9, request)
			if tc.key == "" {
				if err != nil || request.Title != tc.normalized || !request.HideGlobally {
					t.Fatalf("request=%#v err=%v", request, err)
				}
			} else if err == nil || err.Key != tc.key {
				t.Fatalf("err=%#v", err)
			}
			if tc.normalized != "" && store.createCalls > 0 && store.title != tc.normalized {
				t.Fatalf("duplicate lookup title=%q", store.title)
			}
			if (tc.key == "error.title_required" || tc.key == "error.title_too_long") && store.createCalls != 0 {
				t.Fatalf("invalid title queried store %d time(s)", store.createCalls)
			}
		})
	}
}

func TestCategoryUpdateContract(t *testing.T) {
	store := &contractCategoryStore{}
	request := &model.CategoryModificationRequest{}
	if err := ValidateCategoryModification(store, 3, 4, request); err != nil || store.updateCalls != 0 {
		t.Fatalf("nil update err=%v calls=%d", err, store.updateCalls)
	}

	title := "  Alerts  "
	request.Title = &title
	if err := ValidateCategoryModification(store, 3, 4, request); err != nil {
		t.Fatal(err)
	}
	if *request.Title != "Alerts" || store.title != "Alerts" {
		t.Fatalf("title=%q lookup=%q", *request.Title, store.title)
	}
	category := &model.Category{Title: "Old", HideGlobally: true}
	request.Patch(category)
	if category.Title != "Alerts" || !category.HideGlobally {
		t.Fatalf("category=%#v", category)
	}

	store.updateDuplicate = true
	duplicate := " Alerts "
	request.Title = &duplicate
	if err := ValidateCategoryModification(store, 3, 4, request); err == nil || err.Key != "error.category_already_exists" {
		t.Fatalf("duplicate err=%#v", err)
	}
}

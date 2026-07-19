package deliveryreview

import "testing"

func TestReviewRejectsRoot(t *testing.T) {
	if findings := Review(Plan{RuntimeUser: "root"}); len(findings) == 0 {
		t.Fatal("root runtime accepted")
	}
}

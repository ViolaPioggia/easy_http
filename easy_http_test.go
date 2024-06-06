package easy_http

import "testing"

func TestNew(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("unexpect error: %v", err)
	}
	if c == nil {
		t.Fatalf("expected a non-nil client")
	}
}

package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	l := NewRateLimiter(2, 50*time.Millisecond)
	for i, want := range []bool{true, true, false} {
		if ok, _ := l.Allow("a"); ok != want {
			t.Fatalf("hit %d: got %v, want %v", i+1, ok, want)
		}
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Fatal("keys must be independent")
	}
	time.Sleep(60 * time.Millisecond)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("window should have reset")
	}
}

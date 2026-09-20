package ingest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// The route triggers a job that hammers external APIs and rewrites the opportunities
// table. The token check is the only thing between the open internet and it, so the
// rejection cases are the ones worth pinning down.
//
// db is nil on purpose: a rejected request must never reach Run, and would panic here
// if it did.
func TestHandlerRejectsBadToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		header string
	}{
		{"no token", ""},
		{"wrong token", "nope"},
		{"prefix of real token", "secr"},
		{"real token with suffix", "secret-extra"},
		{"empty header", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/internal/ingest", Handler(nil, "secret"))

			req := httptest.NewRequest(http.MethodPost, "/internal/ingest", nil)
			if tc.header != "" {
				req.Header.Set("X-Ingest-Token", tc.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("got %d, want %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

// An empty configured token must never authorise a request, even one that also sends
// an empty header. main.go additionally refuses to register the route in that case.
func TestHandlerEmptyTokenRejectsEmptyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/internal/ingest", Handler(nil, ""))

	req := httptest.NewRequest(http.MethodPost, "/internal/ingest", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("empty token authorised an empty header: got %d", w.Code)
	}
}

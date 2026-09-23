package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Ping(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "." {
		t.Errorf("expected heartbeat response '.', got %q", rec.Body.String())
	}
}

func TestHandler_SubsonicPing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/rest/ping.view?u=test&p=test&v=1.16.1&c=test", nil)
	rec := httptest.NewRecorder()

	Handler(rec, req)

	// Since authentication may fail for non-existent test user, status should still be 200 with Subsonic XML/JSON error
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 (Subsonic standard), got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

package webui

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"goproxy/config"
	"goproxy/storage"
)

func TestHandleLoginSetsSessionCookieAttributes(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WebUIPasswordHash = fmt.Sprintf("%x", sha256.Sum256([]byte("secret")))
	s := &Server{cfg: cfg}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	s.handleLogin(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect, got %d", res.StatusCode)
	}

	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}
	cookie := cookies[0]
	if !cookie.HttpOnly {
		t.Fatal("expected HttpOnly session cookie")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
	}
	if cookie.Secure {
		t.Fatal("expected non-secure cookie on plain HTTP request")
	}
}

func TestHandleLoginSetsSecureCookieWhenForwardedHTTPS(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WebUIPasswordHash = fmt.Sprintf("%x", sha256.Sum256([]byte("secret")))
	s := &Server{cfg: cfg}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	s.handleLogin(rec, req)

	res := rec.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}
	if !cookies[0].Secure {
		t.Fatal("expected secure cookie when request is forwarded as HTTPS")
	}
}

func TestAPISubscriptionsRedactsSensitiveFields(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	defer store.Close()

	if _, err := store.AddSubscription("demo", "https://example.com/sub?token=secret", "/tmp/private.yaml", "auto", 60); err != nil {
		t.Fatalf("add subscription: %v", err)
	}

	s := &Server{storage: store}
	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions", nil)
	rec := httptest.NewRecorder()

	s.apiSubscriptions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var subs []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &subs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if _, ok := subs[0]["url"]; ok {
		t.Fatal("expected url to be redacted from response")
	}
	if _, ok := subs[0]["file_path"]; ok {
		t.Fatal("expected file_path to be redacted from response")
	}
}

func TestAPISubscriptionContributeDisabled(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/api/subscription/contribute", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.apiSubscriptionContribute(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAPISubscriptionAddRejectsOversizedBody(t *testing.T) {
	s := &Server{}
	oversized := `{"name":"demo","file_content":"` + strings.Repeat("a", int(subscriptionJSONBodyMaxBytes)) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscription/add", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.apiSubscriptionAdd(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

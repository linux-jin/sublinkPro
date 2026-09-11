package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	socks5service "sublink/services/socks5"
)

func performSocks5JSONRequest(t *testing.T, username, method string, payload any, handler func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequestWithContext(context.Background(), method, "/", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("username", username)
	handler(ctx)
	return recorder
}

func TestGetSocks5SettingsRequiresAdmin(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	recorder := performBackupRequest(t, "member", GetSocks5Settings)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestUpdateSocks5SettingsPersistsAndMasksPassword(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	t.Setenv("SUBLINK_API_ENCRYPTION_KEY", "socks5-test-key-0123456789abcdef0123456789")
	recorder := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"enabled":       false,
		"listenAddress": "127.0.0.1",
		"port":          1080,
		"username":      "proxy-user",
		"password":      "proxy-secret",
		"selection":     "best",
		"requireAuth":   true,
	}, UpdateSocks5Settings)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "proxy-secret") || strings.Contains(body, `"password"`) {
		t.Fatalf("response leaked password: %s", body)
	}
	if !strings.Contains(body, `"hasPassword":true`) {
		t.Fatalf("expected masked password metadata: %s", body)
	}
	loaded, err := socks5service.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Username != "proxy-user" || loaded.Password != "proxy-secret" || loaded.Enabled || !loaded.RequireAuth {
		t.Fatalf("unexpected stored config: %+v", loaded)
	}
	if _, err := socks5service.SaveConfig(socks5service.Config{
		Enabled:       false,
		ListenAddress: "127.0.0.1",
		Port:          1080,
		Username:      "proxy-user",
		Selection:     "best",
		RequireAuth:   false,
		ClearPassword: true,
	}); err != nil {
		t.Fatalf("clear SOCKS5 password: %v", err)
	}
	cleared, err := socks5service.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Password != "" || cleared.RequireAuth {
		t.Fatalf("expected cleared password and disabled auth, got %+v", cleared)
	}
}

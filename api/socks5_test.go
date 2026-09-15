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
		"enabled":                      false,
		"listenAddress":                "127.0.0.1",
		"port":                         1080,
		"username":                     "proxy-user",
		"password":                     "proxy-secret",
		"selection":                    "round_robin",
		"requireAuth":                  true,
		"maxAttempts":                  4,
		"dialTimeoutSeconds":           12,
		"failureCooldownSeconds":       45,
		"specificFallback":             true,
		"maxConnections":               100,
		"maxConnectionsPerClient":      8,
		"idleTimeoutSeconds":           90,
		"maxConnectionDurationSeconds": 3600,
		"healthCheckEnabled":           true,
		"healthCheckIntervalSeconds":   120,
		"healthCheckTimeoutSeconds":    8,
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
	if loaded.Selection != "round_robin" || loaded.MaxAttempts != 4 || loaded.DialTimeoutSeconds != 12 || loaded.FailureCooldownSeconds != 45 || !loaded.SpecificFallback {
		t.Fatalf("unexpected phase-two settings: %+v", loaded)
	}
	if loaded.MaxConnections != 100 || loaded.MaxConnectionsPerClient != 8 || loaded.IdleTimeoutSeconds != 90 || loaded.MaxConnectionDurationSeconds != 3600 {
		t.Fatalf("unexpected connection governance settings: %+v", loaded)
	}
	if !loaded.HealthCheckEnabled || loaded.HealthCheckIntervalSeconds != 120 || loaded.HealthCheckTimeoutSeconds != 8 {
		t.Fatalf("unexpected health check settings: %+v", loaded)
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

func TestUpdateSocks5SettingsPreservesPhaseTwoFieldsForLegacyRequest(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	if _, err := socks5service.SaveConfig(socks5service.Config{
		ListenAddress:              "127.0.0.1",
		Port:                       1080,
		Selection:                  "round_robin",
		RequireAuth:                false,
		MaxAttempts:                5,
		DialTimeoutSeconds:         20,
		FailureCooldownSeconds:     60,
		SpecificFallback:           true,
		HealthCheckEnabled:         true,
		HealthCheckIntervalSeconds: 180,
		HealthCheckTimeoutSeconds:  9,
	}); err != nil {
		t.Fatal(err)
	}
	recorder := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"enabled":       false,
		"listenAddress": "127.0.0.1",
		"port":          1081,
		"username":      "legacy",
		"selection":     "best",
		"requireAuth":   false,
	}, UpdateSocks5Settings)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	loaded, err := socks5service.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.MaxAttempts != 5 || loaded.DialTimeoutSeconds != 20 || loaded.FailureCooldownSeconds != 60 || !loaded.SpecificFallback {
		t.Fatalf("legacy request overwrote phase-two fields: %+v", loaded)
	}
	if !loaded.HealthCheckEnabled || loaded.HealthCheckIntervalSeconds != 180 || loaded.HealthCheckTimeoutSeconds != 9 {
		t.Fatalf("legacy request overwrote health check fields: %+v", loaded)
	}
}

func TestUpdateSocks5SettingsRejectsInvalidRoutingLimits(t *testing.T) {
	setupBackupAPITestDB(t)
	recorder := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"enabled":       false,
		"listenAddress": "127.0.0.1",
		"port":          1080,
		"selection":     "best",
		"requireAuth":   false,
		"maxAttempts":   6,
	}, UpdateSocks5Settings)
	if !strings.Contains(recorder.Body.String(), `"code":500`) || !strings.Contains(recorder.Body.String(), "max attempts") {
		t.Fatalf("invalid routing limits were accepted: %s", recorder.Body.String())
	}
}

func TestGetSocks5StatusAndConnectionsRequireAdmin(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	statusRecorder := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, GetSocks5Status)
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("status endpoint code = %d, body = %s", statusRecorder.Code, statusRecorder.Body.String())
	}
	if !strings.Contains(statusRecorder.Body.String(), `"activeConnections":0`) || !strings.Contains(statusRecorder.Body.String(), `"health"`) {
		t.Fatalf("status response missing gateway stats or health: %s", statusRecorder.Body.String())
	}
	connectionsRecorder := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, GetSocks5Connections)
	if connectionsRecorder.Code != http.StatusOK {
		t.Fatalf("connections endpoint code = %d, body = %s", connectionsRecorder.Code, connectionsRecorder.Body.String())
	}
	if !strings.Contains(connectionsRecorder.Body.String(), `"data":[]`) {
		t.Fatalf("connections response should be empty: %s", connectionsRecorder.Body.String())
	}
	for _, handler := range []func(*gin.Context){GetSocks5Status, GetSocks5Connections} {
		recorder := performSocks5JSONRequest(t, "member", http.MethodGet, nil, handler)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("non-admin status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
	}
}

func TestProbeSocks5HealthRequiresRunningAdminGateway(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	member := performSocks5JSONRequest(t, "member", http.MethodPost, nil, ProbeSocks5Health)
	if member.Code != http.StatusForbidden {
		t.Fatalf("member status = %d, want %d", member.Code, http.StatusForbidden)
	}
	admin := performSocks5JSONRequest(t, "admin", http.MethodPost, nil, ProbeSocks5Health)
	if admin.Code != http.StatusOK || !strings.Contains(admin.Body.String(), `"code":500`) || !strings.Contains(admin.Body.String(), "healthProbeUnavailable") {
		t.Fatalf("stopped gateway probe response = %s", admin.Body.String())
	}
}

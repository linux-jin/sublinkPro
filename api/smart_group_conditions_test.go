package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sublink/database"
	"sublink/models"
	"testing"
)

func TestSmartGroupConditionsAPIPreservesOmittedFieldsAndClearsExplicitEmpty(t *testing.T) {
	setupPreviewAPITestDB(t)
	router := gin.New()
	router.POST("/", CreateSmartGroup)
	router.PUT("/:id", UpdateSmartGroup)
	candidate := `{"logic":"and","conditions":[{"field":"protocol","operator":"equals","value":"vless"}]}`
	member := `{"logic":"and","conditions":[{"field":"quality_status","operator":"equals","value":"success"}]}`
	payload := map[string]any{"name": "conditions", "countries": "GB", "candidateConditions": candidate, "memberConditions": member}
	send := func(method, path string, body map[string]any) *httptest.ResponseRecorder {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(raw)))
		return rec
	}
	rec := send(http.MethodPost, "/", payload)
	if rec.Code != http.StatusOK {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	var response struct{ Data models.SmartGroup }
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	path := "/" + strconv.Itoa(response.Data.ID)
	legacy := map[string]any{"name": "renamed", "countries": "GB"}
	rec = send(http.MethodPut, path, legacy)
	if rec.Code != http.StatusOK {
		t.Fatalf("legacy PUT: %s", rec.Body.String())
	}
	saved, err := models.GetSmartGroup(response.Data.ID)
	if err != nil || saved.CandidateConditions != candidate || saved.MemberConditions != member {
		t.Fatalf("omitted fields lost: %+v %v", saved, err)
	}
	for _, bad := range []any{nil, `{"logic":"and","conditions":[{"field":"delay_status","operator":"equals","value":"success"}]}`, `{"logic":"or","conditions":[]}`} {
		legacy["candidateConditions"] = bad
		rec = send(http.MethodPut, path, legacy)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid field accepted: %v %d %s", bad, rec.Code, rec.Body.String())
		}
	}
	saved, err = models.GetSmartGroup(response.Data.ID)
	if err != nil || saved.CandidateConditions != candidate {
		t.Fatal("invalid update changed stored conditions")
	}
	legacy["candidateConditions"] = ""
	legacy["memberConditions"] = ""
	rec = send(http.MethodPut, path, legacy)
	if rec.Code != http.StatusOK {
		t.Fatalf("clear: %s", rec.Body.String())
	}
	saved, err = models.GetSmartGroup(response.Data.ID)
	if err != nil || saved.CandidateConditions != "" || saved.MemberConditions != "" {
		t.Fatalf("clear failed: %+v %v", saved, err)
	}
	var count int64
	if err := database.DB.Model(&models.SmartGroup{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatal("unexpected group writes")
	}
}

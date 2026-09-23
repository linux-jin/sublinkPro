package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"sublink/database"
	"sublink/models"
)

func TestSmartGroupPreviewAndMembers(t *testing.T) {
	setupPreviewAPITestDB(t)
	group := models.SmartGroup{Name: "preview-smart", Countries: "GB,DE", MaxDelay: 400, MaxAgeHours: 72}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, item := range []struct{ name, source, country string }{
		{"smart-gb", "source-one", "GB"}, {"smart-de", "source-two", "DE"}, {"smart-us", "source-three", "US"},
	} {
		node := models.Node{Name: item.name, LinkName: item.name, Link: "ss://" + item.name, Protocol: "ss", Group: item.source, LinkCountry: item.country,
			DelayStatus: "success", SpeedStatus: "success", DelayTime: 100, Speed: 3, LatencyCheckAt: now, SpeedCheckAt: now}
		if err := node.Add(); err != nil {
			t.Fatal(err)
		}
	}
	result, err := previewFormSubscription(PreviewRequest{SmartGroupIDs: []int{group.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("preview nodes = %d, want 2", len(result.Nodes))
	}
	router := gin.New()
	router.GET("/:id/members", SmartGroupMembers)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/"+strconv.Itoa(group.ID)+"/members", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("members status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Data struct {
			Count int `json:"count"`
			Nodes []struct {
				Group string `json:"group"`
			} `json:"nodes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Count != 2 || response.Data.Nodes[0].Group != "source-one" || response.Data.Nodes[1].Group != "source-two" {
		t.Fatalf("members response = %+v", response.Data)
	}
}

func TestDeleteSmartGroupRejectsSubscribedGroup(t *testing.T) {
	setupPreviewAPITestDB(t)
	group := models.SmartGroup{Name: "used", Countries: "GB"}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	sub := models.Subcription{Name: "reference", SmartGroupIDs: strconv.Itoa(group.ID)}
	if err := database.DB.Create(&sub).Error; err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.DELETE("/:id", DeleteSmartGroup)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/"+strconv.Itoa(group.ID), nil))
	if recorder.Code != http.StatusConflict {
		t.Fatalf("delete status = %d: %s", recorder.Code, recorder.Body.String())
	}
	if _, err := models.GetSmartGroup(group.ID); err != nil {
		t.Fatalf("referenced group removed: %v", err)
	}
}

func TestCreateAndUpdateSmartGroup(t *testing.T) {
	setupPreviewAPITestDB(t)
	router := gin.New()
	router.POST("/", CreateSmartGroup)
	router.PUT("/:id", UpdateSmartGroup)
	create := httptest.NewRecorder()
	router.ServeHTTP(create, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBufferString(`{"name":"Europe","countries":"gb,FR","maxDelay":500}`)))
	if create.Code != http.StatusOK {
		t.Fatalf("create status = %d: %s", create.Code, create.Body.String())
	}
	var response struct {
		Data models.SmartGroup `json:"data"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Countries != "GB,FR" || response.Data.MaxAgeHours != 72 {
		t.Fatalf("created group = %+v", response.Data)
	}
	update := httptest.NewRecorder()
	router.ServeHTTP(update, httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/"+strconv.Itoa(response.Data.ID), bytes.NewBufferString(`{"name":"Europe","countries":"DE","maxDelay":0,"minSpeed":0,"maxAgeHours":0}`)))
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d: %s", update.Code, update.Body.String())
	}
	saved, err := models.GetSmartGroup(response.Data.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Countries != "DE" || saved.MaxAgeHours != 0 || saved.MaxDelay != 0 {
		t.Fatalf("updated group = %+v", saved)
	}
}

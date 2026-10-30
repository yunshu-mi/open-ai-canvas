package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"yingce/backend/internal/model"
	"yingce/backend/internal/outbound/outboundtest"
)

func TestVintedModelCatalogURL(t *testing.T) {
	for _, baseURL := range []string{"https://vinted.cam", "https://vinted.cam/", "https://vinted.cam/v1", "https://VINTED.CAM/v1/"} {
		if got := strings.ToLower(vintedModelCatalogURL(baseURL)); got != "https://vinted.cam/v1/catalog" {
			t.Fatalf("%q -> %q", baseURL, got)
		}
	}
	for _, baseURL := range []string{"https://vinted.cam.example.com", "https://example.com/vinted.cam", "https://vinted.cam/custom", "https://example.com/v1"} {
		if got := vintedModelCatalogURL(baseURL); got != "" {
			t.Fatalf("unrelated channel %q matched: %q", baseURL, got)
		}
	}
}

func TestVintedModelCatalogRejectsStaleAndUnavailableModels(t *testing.T) {
	validUntil := time.Now().Add(time.Hour).Format(time.RFC3339Nano)
	models := `[{"id":"seedance-2.0","available":false},{"id":"seedance-2.0-fast","available":false},{"id":"seedance-2.5","name":"Seedance 2.5","available":true},{"id":"seedance-2.5","available":true},{"id":"multiref:sd-2-c4","available":true},{"id":"multiref:ali-wan3","available":true},{"id":"qsqs:wan3-720p-deal","available":true},{"id":"frameflow:seedance-2.0","available":true},{"id":"unknown","available":true}]`
	catalog, err := parseVintedModelCatalog([]byte(fmt.Sprintf(`{"status":"ready","revision":"rev","valid_until":%q,"models":%s}`, validUntil, models)))
	if err != nil || len(catalog) != 1 || catalog[0].ID != "seedance-2.5" || catalog[0].ModelType != "video" || catalog[0].DisplayName != "Seedance 2.5" {
		t.Fatalf("catalog=%#v err=%v", catalog, err)
	}
	for _, payload := range []string{
		`<html>gateway error</html>`,
		`{"status":"ready","revision":"rev","valid_until":"invalid"}`,
		`{"status":"ready","revision":"rev","valid_until":"2000-01-01T00:00:00Z"}`,
		fmt.Sprintf(`{"status":"stale","revision":"rev","valid_until":%q}`, validUntil),
		fmt.Sprintf(`{"status":"unavailable","revision":"rev","valid_until":%q}`, validUntil),
		fmt.Sprintf(`{"status":"ready","valid_until":%q}`, validUntil),
	} {
		if _, err := parseVintedModelCatalog([]byte(payload)); err == nil {
			t.Fatalf("invalid catalog accepted: %s", payload)
		}
	}
}

func TestVintedSeedanceModelDiscovery(t *testing.T) {
	validUntil := time.Now().Add(time.Hour).Format(time.RFC3339Nano)
	payload := fmt.Sprintf(`{"status":"ready","revision":"rev","valid_until":%q,"models":[{"id":"seedance-2.0","name":"Seedance 2.0","available":true},{"id":"seedance-2.0-fast","name":"Seedance 2.0 Fast","available":true},{"id":"seedance-2.5","name":"Seedance 2.5","available":true},{"id":"seedance-2.5","available":true},{"id":"unknown","available":true}]}`, validUntil)
	catalog, err := parseVintedModelCatalog([]byte(payload))
	if err != nil || len(catalog) != 3 {
		t.Fatalf("catalog=%#v err=%v", catalog, err)
	}
	for i, id := range []string{"seedance-2.0", "seedance-2.0-fast", "seedance-2.5"} {
		if catalog[i].ID != id || catalog[i].ModelType != "video" || catalog[i].DisplayName == "" {
			t.Fatalf("catalog item=%#v", catalog[i])
		}
	}
}

func TestVintedAdminChannelModelDiscoveryAndImport(t *testing.T) {
	outboundtest.DNS(t, [4]byte{127, 0, 0, 1}, "vinted.cam")
	t.Setenv("CANVAS_ALLOWED_PRIVATE_UPSTREAM_HOSTS", "vinted.cam")
	for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy"} {
		t.Setenv(name, "")
	}
	statusCode, requests := http.StatusOK, 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodGet || r.URL.Path != "/v1/catalog" || r.URL.RawQuery != "" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer stored-key" || r.Header.Get("X-Catalog-Test") != "stored-header" {
			t.Error("stored channel authentication or headers missing")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ready", "revision": "rev", "valid_until": time.Now().Add(time.Hour).Format(time.RFC3339Nano),
			"models": []any{
				map[string]any{"id": "seedance-2.0", "available": true},
				map[string]any{"id": "seedance-2.0-fast", "available": true},
				map[string]any{"id": "seedance-2.5", "available": true},
				map[string]any{"id": "multiref:sd-2-c4", "available": true},
				map[string]any{"id": "frameflow:seedance-2.0", "available": true},
			},
		})
	}))
	defer upstream.Close()
	svc, db := newChannelModelTestService(t)
	admin := &model.User{ID: "admin", Role: model.UserRoleAdmin}
	channel := model.ModelChannel{ID: "vinted-catalog", Scope: model.ChannelScopeSystem, Enabled: true, BaseURL: strings.Replace(upstream.URL, "127.0.0.1", "vinted.cam", 1), APIKey: "stored-key", APIFormat: "openai", ModelsJSON: `[]`, HeadersJSON: `[{"name":"X-Catalog-Test","value":"stored-header"}]`}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PreviewAdminChannelModels(context.Background(), &model.User{ID: "user", Role: model.UserRoleUser}, channel.ID); err == nil || requests != 0 {
		t.Fatal("non-admin reached stored channel credentials")
	}
	want := []string{"seedance-2.0", "seedance-2.0-fast", "seedance-2.5"}
	preview, err := svc.PreviewAdminChannelModels(context.Background(), admin, channel.ID)
	if err != nil || !reflect.DeepEqual(preview, want) {
		t.Fatalf("preview=%v err=%v", preview, err)
	}
	var count int64
	if err := db.Model(&model.ChannelModel{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatal("preview wrote models")
	}
	imported, err := svc.ImportAdminChannelModels(context.Background(), admin, channel.ID, []string{want[2]})
	if err != nil || imported.Added != 1 {
		t.Fatalf("import=%#v err=%v", imported, err)
	}
	pulled, err := svc.FetchAdminChannelModels(context.Background(), admin, channel.ID)
	if err != nil || pulled.Added != 2 || !reflect.DeepEqual(pulled.Models, want) {
		t.Fatalf("pull=%#v err=%v", pulled, err)
	}
	pulled, err = svc.FetchAdminChannelModels(context.Background(), admin, channel.ID)
	if err != nil || pulled.Added != 0 {
		t.Fatalf("duplicate pull=%#v err=%v", pulled, err)
	}
	var stored []model.ChannelModel
	if err := db.Where("channel_id = ?", channel.ID).Find(&stored).Error; err != nil || len(stored) != 3 {
		t.Fatalf("stored=%#v err=%v", stored, err)
	}
	for _, item := range stored {
		if item.Enabled || item.PriceConfigured {
			t.Fatal("discovery bypassed model pricing and activation")
		}
	}
	for _, code := range []int{http.StatusNotFound, http.StatusUnauthorized} {
		statusCode = code
		_, err = svc.PreviewAdminChannelModels(context.Background(), admin, channel.ID)
		if err == nil || strings.Contains(err.Error(), "/models") || strings.Contains(err.Error(), "stored-key") {
			t.Fatalf("catalog error=%v", err)
		}
	}
}

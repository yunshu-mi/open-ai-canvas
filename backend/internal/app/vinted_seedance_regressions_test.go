package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"yingce/backend/internal/protocol"
)

func vintedSeedanceTestModel(modelID string) map[string]any {
	return map[string]any{
		"id": modelID, "available": true,
		"input_modes":     []string{"text", "references", "frames"},
		"x-canvas-inputs": map[string]any{"image": "images", "video": "videos", "audio": "audios"},
		"parameter_schema": map[string]any{"properties": map[string]any{
			"duration": map[string]any{"default": 4}, "ratio": map[string]any{"default": "16:9"}, "resolution": map[string]any{"default": "720p"},
		}},
	}
}

func vintedSeedanceTestCatalog(model map[string]any, revision string) map[string]any {
	return map[string]any{"status": "ready", "revision": revision, "valid_until": time.Now().Add(time.Hour).Format(time.RFC3339Nano), "models": []any{model}}
}

func TestVintedSeedanceSignedURLRefresh(t *testing.T) {
	allowLoopbackProviderTest(t)
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok || adapter.Metadata().Name != "200t.video.tv" {
		t.Fatal("200t.video.tv package was not installed")
	}
	var imageBytes bytes.Buffer
	if err := png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 301, 301))); err != nil {
		t.Fatal(err)
	}
	for _, modelID := range []string{"seedance-2.0", "seedance-2.0-fast", "seedance-2.5"} {
		t.Run(modelID, func(t *testing.T) {
			var quoted map[string]any
			var submitted []byte
			var key string
			creates, polls, uploads, downloads, signatures := 0, 0, 0, 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/file" {
					downloads++
					if r.Header.Get("Authorization") != "" || r.Header.Get("X-Private") != "" {
						t.Error("signed download leaked channel credentials")
					}
					if downloads == 1 {
						w.WriteHeader(http.StatusForbidden)
						_, _ = io.WriteString(w, `{"error":"expired"}`)
						return
					}
					// 当前接口不保证旧的 X-Video-Quality 响应头。
					w.Header().Set("Content-Type", "video/mp4")
					_, _ = io.WriteString(w, vintedTestMP4)
					return
				}
				if r.Header.Get("Authorization") != "Bearer test-key" {
					t.Error("API authentication missing")
				}
				switch r.URL.Path {
				case "/v1/catalog":
					_ = json.NewEncoder(w).Encode(vintedSeedanceTestCatalog(vintedSeedanceTestModel(modelID), "rev-current"))
				case "/v1/files/content":
					uploads++
					data, _ := io.ReadAll(r.Body)
					kind := r.URL.Query().Get("kind")
					mime := map[string]string{"image": "image/png", "video": "video/mp4", "audio": "audio/wav"}[kind]
					if len(data) == 0 || r.URL.Query().Get("name") == "" || r.ContentLength != int64(len(data)) || r.Header.Get("Content-Type") != mime || !protocolMediaFormatMatches(mime, data) {
						t.Errorf("invalid raw upload: kind=%s mime=%s", kind, r.Header.Get("Content-Type"))
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"asset_id": fmt.Sprintf("asset-%d", uploads), "kind": kind})
				case "/v1/videos/quote":
					_ = json.NewDecoder(r.Body).Decode(&quoted)
					if len(quoted) != 5 || quoted["catalog_revision"] != "rev-current" || quoted["input_mode"] != "references" {
						t.Errorf("invalid quote body: %#v", quoted)
					}
					parameters := quoted["parameters"].(map[string]any)
					_, hasAudio := parameters["save_audio"]
					if hasAudio {
						t.Errorf("save_audio mapping = %#v", parameters)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"quote_id": "quote-1", "credits": 30, "catalog_revision": "rev-current"})
				case "/v1/videos":
					creates++
					body, _ := io.ReadAll(r.Body)
					var actual map[string]any
					_ = json.Unmarshal(body, &actual)
					if !reflect.DeepEqual(quoted, actual) || r.Header.Get("Expected-Quote-Id") != "quote-1" {
						t.Error("submission differs from quote")
					}
					if creates == 1 {
						submitted, key = body, r.Header.Get("Idempotency-Key")
						w.WriteHeader(http.StatusServiceUnavailable)
						_, _ = io.WriteString(w, `{"error":{"message":"temporary"}}`)
						return
					}
					if key == "" || key != r.Header.Get("Idempotency-Key") || !bytes.Equal(submitted, body) {
						t.Error("retry changed operation identity or body")
					}
					w.WriteHeader(http.StatusAccepted)
					_, _ = io.WriteString(w, `{"id":"db_9007199254740993","status":"queued"}`)
				case "/v1/videos/db_9007199254740993":
					polls++
					if polls == 1 {
						_, _ = io.WriteString(w, `{"id":"db_9007199254740993","status":"succeeded","has_file":false}`)
					} else {
						_, _ = io.WriteString(w, `{"id":"db_9007199254740993","status":"succeeded","has_file":true}`)
					}
				case "/v1/videos/db_9007199254740993/signed_url":
					signatures++
					_, _ = io.WriteString(w, `{"url":"/file?signature=test"}`)
				default:
					t.Errorf("unexpected path: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			input := canvasGenerationInput{Mode: "video", Prompt: "镜头缓慢向前", Config: providerConfig{Model: modelID, BaseURL: server.URL, APIKey: "test-key", InterfaceType: "200t-video-tv", VideoSeconds: "10", Size: "16:9", VQuality: "720p", VideoGenerateAudio: "false"}, ReferenceImages: []providerMedia{{DataURL: dataURL("image/png", imageBytes.Bytes()), Role: "reference_image"}}, ReferenceAudios: []providerMedia{{DataURL: dataURL("audio/wav", []byte("RIFF0000WAVEdata"))}}, Metadata: map[string]interface{}{"videoEditOperation": "reference_to_video"}}
			input.Config.Headers = []OutboundHeader{{Name: "X-Private", Value: "private-value"}, {Name: "Expected-Quote-Id", Value: "stale-quote"}}
			input.ReferenceVideos = []providerMedia{{DataURL: dataURL("video/mp4", []byte(vintedTestMP4))}}
			result, err := runProtocolAdapterTaskWithPolicy(context.Background(), input, adapter, videoPollPolicy{Interval: time.Millisecond, Sleep: func(context.Context, time.Duration) error { return nil }})
			if err != nil || result == nil || creates != 2 || polls != 2 || uploads != 2+len(input.ReferenceVideos) || downloads != 2 || signatures != 2 {
				t.Fatalf("result=%#v err=%v creates=%d polls=%d uploads=%d downloads=%d", result, err, creates, polls, uploads, downloads)
			}
		})
	}
}

func TestVintedSeedanceFramesAndTextMappings(t *testing.T) {
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok {
		t.Fatal("200t.video.tv plugin not installed")
	}
	preparer := adapter.(protocol.PreparationAdapter)
	quoteStep := adapter.Metadata().Lifecycle.Prepare[4]
	request := protocol.GenerationRequest{Model: "seedance-2.0", Prompt: "镜头移动", Duration: 8, Resolution: "720p", AspectRatio: "16:9", Images: []protocol.MediaReference{{Role: "last_frame", Order: 0}, {Role: "first_frame", Order: 1}}, Extra: map[string]any{"preparation": map[string]any{
		"catalog": map[string]any{"revision": "rev", "model": vintedSeedanceTestModel("seedance-2.0")},
		"images":  []any{map[string]any{"asset_id": "last", "kind": "image", "role": "last_frame"}, map[string]any{"asset_id": "first", "kind": "image", "role": "first_frame"}},
	}}}
	spec, err := preparer.BuildPreparation(context.Background(), quoteStep, request)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(spec.Body)
	var body map[string]any
	_ = json.Unmarshal(encoded, &body)
	images := body["input"].(map[string]any)["images"].([]any)
	if body["input_mode"] != "frames" || images[0].(map[string]any)["asset_id"] != "first" || images[1].(map[string]any)["asset_id"] != "last" || len(images[0].(map[string]any)) != 2 {
		t.Fatalf("frames body = %s", encoded)
	}
	request.Images = []protocol.MediaReference{{Role: "last_frame"}}
	if _, err := adapter.BuildCreate(context.Background(), protocol.RequestContext{Request: request}); err == nil {
		t.Fatal("tail-only frames accepted")
	}
	request.Model, request.Images = "seedance-2.5", nil
	request.Extra["preparation"] = map[string]any{"catalog": map[string]any{"revision": "rev", "model": vintedSeedanceTestModel("seedance-2.5")}}
	spec, err = preparer.BuildPreparation(context.Background(), quoteStep, request)
	if err != nil {
		t.Fatal(err)
	}
	body = spec.Body.(map[string]any)
	if body["input_mode"] != "text" || len(body["input"].(map[string]any)) != 1 {
		t.Fatalf("text body = %#v", body)
	}
	for _, response := range []string{
		`{"quote_id":"q","catalog_revision":"rev"}`,
		`{"quote_id":"q","credits":"30","catalog_revision":"rev"}`,
		`{"quote_id":"q","credits":-1,"catalog_revision":"rev"}`,
		`{"quote_id":"q","credits":30,"catalog_revision":"different"}`,
	} {
		if _, _, err := preparer.ParsePreparation(context.Background(), quoteStep, request, []byte(response)); err == nil {
			t.Fatalf("invalid quote accepted: %s", response)
		}
	}
	request.Model, request.Videos = "unsupported", nil
	if _, err := adapter.BuildCreate(context.Background(), protocol.RequestContext{Request: request}); err == nil {
		t.Fatal("unsupported model accepted")
	}
}

func TestVintedSeedancePriceChangeDoesNotRequoteOrReplaceOperation(t *testing.T) {
	allowLoopbackProviderTest(t)
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok {
		t.Fatal("200t.video.tv plugin not installed")
	}
	quotes, creates := 0, 0
	var firstBody []byte
	var firstKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/catalog":
			_ = json.NewEncoder(w).Encode(vintedSeedanceTestCatalog(vintedSeedanceTestModel("seedance-2.5"), "rev"))
		case "/v1/videos/quote":
			quotes++
			_, _ = io.WriteString(w, `{"quote_id":"q1","credits":30,"catalog_revision":"rev"}`)
		case "/v1/videos":
			creates++
			body, _ := io.ReadAll(r.Body)
			if creates == 1 {
				firstBody, firstKey = body, r.Header.Get("Idempotency-Key")
			}
			if !bytes.Equal(firstBody, body) || firstKey != r.Header.Get("Idempotency-Key") || r.Header.Get("Expected-Quote-Id") != "q1" {
				t.Error("price rejection changed operation")
			}
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `{"error":{"code":"pricing_changed","message":"price changed"}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	_, err = runProtocolAdapterTaskWithPolicy(context.Background(), canvasGenerationInput{Mode: "video", Prompt: "test", Config: providerConfig{Model: "seedance-2.5", BaseURL: server.URL, VideoSeconds: "10", Size: "16:9", VQuality: "720p"}}, adapter, videoPollPolicy{Sleep: func(context.Context, time.Duration) error { return nil }})
	if err == nil || quotes != 1 || creates < 1 || creates > 3 || firstKey == "" {
		t.Fatalf("err=%v quotes=%d creates=%d", err, quotes, creates)
	}
}

func TestVintedSeedanceDoesNotSubmitUnavailableModel(t *testing.T) {
	allowLoopbackProviderTest(t)
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok {
		t.Fatal("200t.video.tv plugin not installed")
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/v1/catalog" {
			t.Error("unavailable model triggered a write")
		}
		unavailable := vintedSeedanceTestModel("seedance-2.0")
		unavailable["available"] = false
		_ = json.NewEncoder(w).Encode(vintedSeedanceTestCatalog(unavailable, "rev"))
	}))
	defer server.Close()
	_, err = runProtocolAdapterTask(context.Background(), canvasGenerationInput{Mode: "video", Prompt: "test", Config: providerConfig{Model: "seedance-2.0", BaseURL: server.URL, VideoSeconds: "10", Size: "16:9", VQuality: "720p"}}, adapter)
	if err == nil || requests != 1 {
		t.Fatalf("err=%v requests=%d", err, requests)
	}
}

func TestVintedSeedanceRecoveredAndUnknownStates(t *testing.T) {
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok {
		t.Fatal("200t.video.tv plugin not installed")
	}
	for _, tc := range []struct {
		body   string
		status protocol.Status
	}{
		{`{"id":"db_1","status":"running","provider_state":"unknown"}`, protocol.StatusProcessing},
		{`{"id":"db_1","status":"failed","refunded":true,"has_file":true,"recovered_artifact":{"available":true,"settlement_unchanged":true}}`, protocol.StatusSucceeded},
		{`{"id":"db_1","status":"failed","refunded":true,"has_file":true}`, protocol.StatusFailed},
		{`{"id":"db_other","status":"succeeded","has_file":true}`, protocol.StatusFailed},
	} {
		result, err := adapter.ParsePoll(context.Background(), protocol.PollContext{TaskID: "db_1"}, []byte(tc.body))
		if err != nil || result.Status != tc.status {
			t.Fatalf("body=%s result=%#v err=%v", tc.body, result, err)
		}
	}
}

func TestVintedSeedanceDynamicCatalogMapping(t *testing.T) {
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok {
		t.Fatal("200t.video.tv plugin not installed")
	}
	preparer := adapter.(protocol.PreparationAdapter)
	quote := adapter.Metadata().Lifecycle.Prepare[4]
	for _, modelID := range []string{"seedance-2.0", "seedance-2.0-fast", "seedance-2.5"} {
		t.Run(modelID, func(t *testing.T) {
			model := vintedSeedanceTestModel(modelID)
			model["x-canvas-inputs"] = map[string]any{"image": "references", "audio": "references"}
			model["parameter_schema"] = map[string]any{"properties": map[string]any{"duration": map[string]any{"default": 15}, "quality": map[string]any{"default": "720p"}, "extra_declared": map[string]any{"default": "official"}, "no_default": map[string]any{}}}
			request := protocol.GenerationRequest{Model: modelID, Prompt: "test", Resolution: "480p", Extra: map[string]any{"preparation": map[string]any{
				"catalog": map[string]any{"revision": "rev", "model": model},
				"images":  []any{map[string]any{"asset_id": "image-1", "kind": "image", "role": "reference_image"}},
				"audios":  []any{map[string]any{"asset_id": "audio-1", "kind": "audio"}},
			}}, ProviderOptions: map[string]map[string]any{"200t-video-tv": {"parameters": map[string]any{"extra_declared": "override", "undeclared": "must-not-leak"}}}}
			spec, err := preparer.BuildPreparation(context.Background(), quote, request)
			if err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(spec.Body)
			var body map[string]any
			_ = json.Unmarshal(encoded, &body)
			input := body["input"].(map[string]any)
			refs := input["references"].([]any)
			if len(input) != 2 || len(refs) != 2 || refs[0].(map[string]any)["asset_id"] != "image-1" || refs[1].(map[string]any)["asset_id"] != "audio-1" || len(refs[0].(map[string]any)) != 2 {
				t.Fatalf("input=%s", encoded)
			}
			want := map[string]any{"duration": float64(15), "quality": "480p", "extra_declared": "override"}
			if !reflect.DeepEqual(body["parameters"], want) {
				t.Fatalf("parameters=%#v", body["parameters"])
			}
			request.Duration = 6
			request.ProviderOptions["200t-video-tv"]["parameters"].(map[string]any)["duration"] = 30
			spec, err = preparer.BuildPreparation(context.Background(), quote, request)
			if err != nil {
				t.Fatal(err)
			}
			encoded, _ = json.Marshal(spec.Body)
			_ = json.Unmarshal(encoded, &body)
			if body["parameters"].(map[string]any)["duration"] != float64(6) {
				t.Fatal("advanced parameters overrode the unified duration used by billing")
			}
		})
	}
}

func TestVintedSeedanceCatalogRejectsExpiredSnapshotAndMissingMediaMapping(t *testing.T) {
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok {
		t.Fatal("200t.video.tv plugin not installed")
	}
	preparer := adapter.(protocol.PreparationAdapter)
	step := adapter.Metadata().Lifecycle.Prepare[0]
	for _, tc := range []string{"expired", "missing-mapping", "unsupported-mode"} {
		model := vintedSeedanceTestModel("seedance-2.0-fast")
		payload := vintedSeedanceTestCatalog(model, "rev")
		switch tc {
		case "expired":
			payload["valid_until"] = time.Now().Add(-time.Minute).Format(time.RFC3339Nano)
		case "missing-mapping":
			model["x-canvas-inputs"] = map[string]any{}
		case "unsupported-mode":
			model["input_modes"] = []string{"text"}
		}
		encoded, _ := json.Marshal(payload)
		if _, _, err := preparer.ParsePreparation(context.Background(), step, protocol.GenerationRequest{Model: "seedance-2.0-fast", Images: []protocol.MediaReference{{Role: "reference_image"}}}, encoded); err == nil {
			t.Fatalf("accepted %s", tc)
		}
	}
}

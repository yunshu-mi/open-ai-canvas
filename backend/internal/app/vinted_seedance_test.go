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

func TestVintedSeedanceLifecycle(t *testing.T) {
	allowLoopbackProviderTest(t)
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := center.registry.Resolve("200t-video-tv")
	if !ok {
		t.Fatal("200t.video.tv plugin not installed")
	}
	var picture bytes.Buffer
	if err := png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 16, 16))); err != nil {
		t.Fatal(err)
	}
	for _, modelID := range []string{"seedance-2.0", "seedance-2.0-fast", "seedance-2.5"} {
		for _, mode := range []string{"text", "references", "frames"} {
			t.Run(modelID+"/"+mode, func(t *testing.T) {
				var quoted map[string]any
				var submitted []byte
				var key string
				creates, polls, uploads, downloads := 0, 0, 0, 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Path == "/file" {
						downloads++
						if r.Header.Get("Authorization") != "" || r.Header.Get("X-Private") != "" || r.URL.Query().Get("sig") != "test" {
							t.Error("signed download credentials or signature incorrect")
						}
						w.Header().Set("Content-Type", "video/mp4")
						_, _ = io.WriteString(w, vintedTestMP4)
						return
					}
					if r.Header.Get("Authorization") != "Bearer test-key" {
						t.Error("API authentication missing")
					}
					switch r.URL.Path {
					case "/v1/catalog":
						catalogModel := vintedSeedanceTestModel(modelID)
						catalogModel["input_modes"] = []string{"text", "references", "frames"}
						_ = json.NewEncoder(w).Encode(vintedSeedanceTestCatalog(catalogModel, "rev-seedance"))
					case "/v1/files/content":
						uploads++
						data, _ := io.ReadAll(r.Body)
						kind := r.URL.Query().Get("kind")
						mime := map[string]string{"image": "image/png", "video": "video/mp4", "audio": "audio/wav"}[kind]
						if r.Method != "POST" || r.ContentLength != int64(len(data)) || r.URL.Query().Get("name") == "" || r.Header.Get("Content-Type") != mime || !protocolMediaFormatMatches(mime, data) {
							t.Errorf("invalid raw upload: kind=%s length=%d mime=%s", kind, r.ContentLength, r.Header.Get("Content-Type"))
						}
						w.WriteHeader(http.StatusCreated)
						_ = json.NewEncoder(w).Encode(map[string]any{"asset_id": fmt.Sprintf("asset-%d", uploads), "kind": kind})
					case "/v1/videos/quote":
						_ = json.NewDecoder(r.Body).Decode(&quoted)
						if len(quoted) != 5 || quoted["model_id"] != modelID || quoted["input_mode"] != mode || quoted["catalog_revision"] != "rev-seedance" {
							t.Errorf("quote=%#v", quoted)
						}
						parameters := quoted["parameters"].(map[string]any)
						ratio := "9:16"
						if modelID == "seedance-2.5" && mode == "frames" {
							ratio = "adaptive"
						}
						if !reflect.DeepEqual(parameters, map[string]any{"duration": float64(10), "ratio": ratio, "resolution": "720p"}) {
							t.Errorf("parameters=%#v", parameters)
						}
						input := quoted["input"].(map[string]any)
						if mode == "text" && len(input) != 1 {
							t.Errorf("text request contains media: %#v", input)
						}
						if mode == "frames" {
							images := input["images"].([]any)
							if len(images) != 2 || images[0].(map[string]any)["asset_id"] != "asset-2" || images[1].(map[string]any)["asset_id"] != "asset-1" {
								t.Errorf("frames not sorted first/last: %#v", images)
							}
						}
						_ = json.NewEncoder(w).Encode(map[string]any{"quote_id": "q", "credits": 15, "catalog_revision": "rev-seedance"})
					case "/v1/videos":
						creates++
						body, _ := io.ReadAll(r.Body)
						var actual map[string]any
						_ = json.Unmarshal(body, &actual)
						if !reflect.DeepEqual(quoted, actual) || r.Header.Get("Expected-Quote-Id") != "q" {
							t.Error("submission differs from quote")
						}
						if creates == 1 {
							submitted, key = body, r.Header.Get("Idempotency-Key")
							w.WriteHeader(http.StatusServiceUnavailable)
							return
						}
						if key == "" || key != r.Header.Get("Idempotency-Key") || !bytes.Equal(body, submitted) {
							t.Error("retry changed body or idempotency key")
						}
						w.WriteHeader(http.StatusAccepted)
						_, _ = io.WriteString(w, `{"id":"db_seedance","status":"queued"}`)
					case "/v1/videos/db_seedance":
						polls++
						_ = json.NewEncoder(w).Encode(map[string]any{"id": "db_seedance", "status": "succeeded", "has_file": polls > 1})
					case "/v1/videos/db_seedance/signed_url":
						_, _ = io.WriteString(w, `{"url":"/file?sig=test"}`)
					default:
						t.Errorf("unexpected path %s", r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()
				input := canvasGenerationInput{Mode: "video", Prompt: "镜头推进", Config: providerConfig{Model: modelID, BaseURL: server.URL, APIKey: "test-key", InterfaceType: "200t-video-tv", VideoSeconds: "10", Size: "9:16", VQuality: "720p", Headers: []OutboundHeader{{Name: "X-Private", Value: "private"}}}, Metadata: map[string]any{"videoOperation": "text_to_video"}}
				expectedUploads := 0
				if mode == "references" {
					expectedUploads = 3
					input.ReferenceImages = []providerMedia{{DataURL: dataURL("image/png", picture.Bytes()), Role: "reference_image"}}
					input.ReferenceVideos = []providerMedia{{DataURL: dataURL("video/mp4", []byte(vintedTestMP4))}}
					input.ReferenceAudios = []providerMedia{{DataURL: dataURL("audio/wav", []byte("RIFF0000WAVEdata"))}}
					input.Metadata["videoOperation"] = "reference_to_video"
				}
				if mode == "frames" {
					expectedUploads = 2
					input.ReferenceImages = []providerMedia{{DataURL: dataURL("image/png", picture.Bytes()), Role: "last_frame"}, {DataURL: dataURL("image/png", picture.Bytes()), Role: "first_frame"}}
					input.Metadata["videoOperation"] = "image_to_video"
					if modelID == "seedance-2.5" {
						input.Config.Size = "adaptive"
					}
				}
				profile := DefaultModelCapabilityConfigForModel("200t-video-tv", modelID).Video
				if err := validateVideoTask(profile, input); err != nil {
					t.Fatal(err)
				}
				result, err := runProtocolAdapterTaskWithPolicy(context.Background(), input, adapter, videoPollPolicy{Interval: time.Millisecond, Sleep: func(context.Context, time.Duration) error { return nil }})
				if err != nil || result == nil || uploads != expectedUploads || creates != 2 || polls != 2 || downloads != 1 {
					t.Fatalf("result=%#v err=%v uploads=%d creates=%d polls=%d downloads=%d", result, err, uploads, creates, polls, downloads)
				}
			})
		}
	}
}

func TestVintedSeedanceRejectsInvalidRequestsAndQuotes(t *testing.T) {
	center, err := newPluginRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	adapter, _ := center.registry.Resolve("200t-video-tv")
	valid := protocol.GenerationRequest{Model: "seedance-2.5", Prompt: "test", Duration: 30, AspectRatio: "9:16", Resolution: "720p"}
	for _, change := range []func(*protocol.GenerationRequest){
		func(r *protocol.GenerationRequest) { r.Model = "multiref:sd-2-c4" },
		func(r *protocol.GenerationRequest) { r.Model = "seedance-2.0" },
		func(r *protocol.GenerationRequest) { r.Resolution = "1080p" },
		func(r *protocol.GenerationRequest) { r.AspectRatio = "adaptive" },
		func(r *protocol.GenerationRequest) { r.Audios = []protocol.MediaReference{{Kind: "audio"}} },
		func(r *protocol.GenerationRequest) { r.Images = []protocol.MediaReference{{Role: "last_frame"}} },
		func(r *protocol.GenerationRequest) { r.Images = make([]protocol.MediaReference, 10) },
	} {
		request := valid
		change(&request)
		if _, err := adapter.BuildCreate(context.Background(), protocol.RequestContext{Request: request}); err == nil {
			t.Fatalf("accepted invalid request %#v", request)
		}
	}
	preparer := adapter.(protocol.PreparationAdapter)
	quote := adapter.Metadata().Lifecycle.Prepare[4]
	valid.Extra = map[string]any{"preparation": map[string]any{"catalog": map[string]any{"revision": "rev"}}}
	for _, response := range []string{
		`{"quote_id":"q","credits":0,"catalog_revision":"rev"}`,
		`{"quote_id":"q","credits":-1,"catalog_revision":"rev"}`,
		`{"quote_id":"q","credits":1.5,"catalog_revision":"rev"}`,
		`{"credits":15,"catalog_revision":"rev"}`,
		`{"quote_id":"q","credits":15,"catalog_revision":"stale"}`,
	} {
		if _, _, err := preparer.ParsePreparation(context.Background(), quote, valid, []byte(response)); err == nil {
			t.Fatalf("accepted invalid quote %s", response)
		}
	}
}

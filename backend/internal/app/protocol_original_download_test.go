package app

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const vintedTestMP4 = "\x00\x00\x00\x18ftypisom\x00\x00\x02\x00isomiso2"

func TestProtocolOriginalDownloadResumesSameFile(t *testing.T) {
	allowLoopbackProviderTest(t)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Authorization") != "" {
			t.Error("signed download must be anonymous")
		}
		if requests == 1 {
			connection, buffer, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_, _ = fmt.Fprintf(buffer, "HTTP/1.1 200 OK\r\nContent-Type: video/mp4\r\nX-Video-Quality: original\r\nETag: original-file\r\nContent-Length: %d\r\n\r\n%s", len(vintedTestMP4), vintedTestMP4[:12])
			_ = buffer.Flush()
			_ = connection.Close()
			return
		}
		if r.Header.Get("Range") != "bytes=12-" || r.Header.Get("If-Range") != "original-file" {
			t.Error("resume headers missing")
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("X-Video-Quality", "original")
		w.Header().Set("ETag", "original-file")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 12-%d/%d", len(vintedTestMP4)-1, len(vintedTestMP4)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte(vintedTestMP4[12:]))
	}))
	defer server.Close()
	state := &protocolDownloadState{}
	config := providerConfig{BaseURL: server.URL, APIKey: "secret"}
	if _, _, err := downloadProtocolOriginal(context.Background(), config, server.URL+"/file", state); err == nil {
		t.Fatal("truncated download accepted")
	}
	data, _, err := downloadProtocolOriginal(context.Background(), config, server.URL+"/file?new-signature=1", state)
	if err != nil || !bytes.Equal(data, []byte(vintedTestMP4)) || requests != 2 {
		t.Fatalf("data=%q err=%v requests=%d", data, err, requests)
	}
}

func TestProtocolOriginalRejectsPreviewHeaderAndFakeMP4(t *testing.T) {
	allowLoopbackProviderTest(t)
	for _, tc := range []struct{ name, quality, body string }{
		{"preview", "preview", vintedTestMP4}, {"missing_quality", "", vintedTestMP4}, {"html_disguised_as_mp4", "original", "<html>not a video</html>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "video/mp4")
				w.Header().Set("X-Video-Quality", tc.quality)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			if _, _, err := downloadProtocolOriginal(context.Background(), providerConfig{BaseURL: server.URL}, server.URL+"/file", &protocolDownloadState{}); err == nil {
				t.Fatal("unverified original was accepted")
			}
		})
	}
}

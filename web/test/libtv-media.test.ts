import { describe, expect, test } from "bun:test";

import { buildLibTVImagePreviewUrl, buildLibTVVideoPreviewUrl, buildLibTVVideoSourceUrl } from "../src/lib/canvas/libtv-media";

describe("LibTV media helpers", () => {
    test("uses a LibTV CDN thumbnail without changing the original media contract", () => {
        const original = "https://libtv-res.liblib.art/path/example.png";
        expect(buildLibTVImagePreviewUrl(original)).toBe("https://libtv-res.liblib.art/path/example.png?x-oss-process=image%2Fresize%2Cw_960");
        expect(buildLibTVImagePreviewUrl("https://example.com/path/example.png")).toBe("https://example.com/path/example.png");
    });

    test("builds a 400px OSS snapshot for LibTV videos only", () => {
        const original = "https://libtv-res.liblib.art/path/example.mp4";
        expect(buildLibTVVideoPreviewUrl(original)).toBe("https://libtv-res.liblib.art/path/example.mp4?x-oss-process=video%2Fsnapshot%2Ct_0%2Cf_jpg%2Cw_400%2Cm_fast%2Car_auto");
        expect(buildLibTVVideoPreviewUrl("https://example.com/path/example.mp4")).toBe("");
        expect(buildLibTVVideoPreviewUrl("not-a-url")).toBe("");
    });

    test("removes a persisted OSS snapshot transform before video playback", () => {
        const snapshot = "https://libtv-res.liblib.art/path/example.mp4?x-oss-process=video%2Fsnapshot%2Ct_0%2Cf_jpg%2Cw_400";
        expect(buildLibTVVideoSourceUrl(snapshot)).toBe("https://libtv-res.liblib.art/path/example.mp4");
        expect(buildLibTVVideoSourceUrl("https://example.com/path/example.mp4?x-oss-process=video%2Fsnapshot")).toBe("https://example.com/path/example.mp4?x-oss-process=video%2Fsnapshot");
    });
});

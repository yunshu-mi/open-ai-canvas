import { expect, test } from "bun:test";
import { defaultModelCapabilityConfig } from "../src/lib/model-capabilities";
import { modelCompatibilityError } from "../src/lib/model-selection";
import { createModelChannel, defaultConfig } from "../src/stores/use-config-store";

test("200t.video.tv 三个模型的参数与纯文字选择", () => {
    for (const model of ["seedance-2.0", "seedance-2.0-fast", "seedance-2.5"]) {
        const capabilityConfig = defaultModelCapabilityConfig("200t-video-tv", model);
        const video = capabilityConfig.video!;
        expect(video.duration).toEqual({ selection: "range", min: 4, max: model === "seedance-2.5" ? 30 : 15, step: 1, default: 10 });
        expect(video.resolutions).toEqual(["480p", "720p"]);
        expect(video.references).toMatchObject({ promptMaxChars: 10000, maxImages: 9, maxVideos: 3, maxAudios: 3, maxImageBytes: 20 * 1024 * 1024, maxVideoBytes: 100 * 1024 * 1024, maxAudioBytes: 100 * 1024 * 1024 });
        expect(video.ratios.includes("adaptive")).toBe(model === "seedance-2.5");
        expect(video.generateAudio.supported).toBe(false);
        expect(video.watermark.supported).toBe(false);
        expect(video.operations).not.toContain("audio_to_video");
        const channel = createModelChannel({
            id: "vinted",
            models: [model],
            interfaceType: "200t-video-tv",
            modelCosts: [{ model, capability: "video", protocol: "200t-video-tv", billingMode: "fixed_request", unitPriceMicrocredits: 0, capabilityConfig }],
        });
        expect(modelCompatibilityError({ ...defaultConfig, channels: [channel] }, `vinted::${model}`, { capability: "video", input: { textCount: 1, imageCount: 0, videoCount: 0, audioCount: 0, characterCount: 0 } })).toBe("");
    }
});

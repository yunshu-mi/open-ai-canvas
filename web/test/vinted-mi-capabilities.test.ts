import { expect, test } from "bun:test";
import { defaultModelCapabilityConfig } from "../src/lib/model-capabilities";
import { modelCompatibilityError } from "../src/lib/model-selection";
import { createModelChannel, defaultConfig } from "../src/stores/use-config-store";

test("200t.vinted.mi keeps three different video contracts", () => {
    const c4 = defaultModelCapabilityConfig("200t-vinted-mi", "multiref:sd-2-c4").video!;
    const wan = defaultModelCapabilityConfig("200t-vinted-mi", "multiref:ali-wan3").video!;
    const deal = defaultModelCapabilityConfig("200t-vinted-mi", "qsqs:wan3-720p-deal").video!;
    expect(c4.duration).toEqual({ selection: "range", min: 4, max: 15, step: 1, default: 4 });
    expect(c4.resolutions).toEqual(["720p"]);
    expect(c4.references.maxImages).toBe(9);
    expect(wan.duration.max).toBe(30);
    expect(wan.resolutions).toEqual(["480p", "720p", "1080p"]);
    expect(wan.generateAudio).toEqual({ supported: true, default: true });
    expect(wan.references.maxVideos).toBe(5);
    expect(deal.duration).toEqual({ selection: "enum", values: [5, 10, 15, 30], default: 30 });
    expect(deal.ratios).toEqual(["16:9", "9:16"]);
    expect(deal.references.maxVideos).toBe(0);
    expect(deal.references.maxAudios).toBe(5);
    expect(deal.references.promptMaxChars).toBe(10000);
});

test("三个模型的纯文字请求可以通过创作页模型选择", () => {
    for (const model of ["multiref:sd-2-c4", "multiref:ali-wan3", "qsqs:wan3-720p-deal"]) {
        const capabilityConfig = defaultModelCapabilityConfig("200t-vinted-mi", model);
        expect(capabilityConfig.video!.operations).toContain("text_to_video");
        expect(capabilityConfig.video!.defaultOperation).toBe("text_to_video");
        const channel = createModelChannel({ id: "vinted", models: [model], interfaceType: "200t-vinted-mi", modelCosts: [{ model, capability: "video", protocol: "200t-vinted-mi", billingMode: "fixed_request", unitPriceMicrocredits: 0, capabilityConfig }] });
        const config = { ...defaultConfig, channels: [channel] };
        expect(modelCompatibilityError(config, `vinted::${model}`, { capability: "video", input: { textCount: 1, imageCount: 0, videoCount: 0, audioCount: 0, characterCount: 0 } })).toBe("");
    }
});

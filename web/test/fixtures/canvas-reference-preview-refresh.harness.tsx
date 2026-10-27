import { afterAll, expect, mock, spyOn, test } from "bun:test";
import { Window } from "happy-dom";

const dom = new Window();
Object.assign(globalThis, { window: dom, document: dom.document, IS_REACT_ACT_ENVIRONMENT: true });
const calls: boolean[] = [];
let now = 100_000;
let fail = false;
const clock = spyOn(Date, "now").mockImplementation(() => now);
mock.module("../../src/services/api/resources", () => ({
    resourceIdFromStorageKey: (key?: string) => (key?.startsWith("resource:") ? key.slice(9) : ""),
    resolveResourceAccessURL: (url: string) => url,
    getResourceAccess: async (_key: string, _purpose: string, _variant: string, _name: string, options: { forceRefresh: boolean }) => {
        calls.push(options.forceRefresh);
        if (fail) throw new Error("temporary failure");
        return { url: `https://example.test/preview-${calls.length}`, refreshAt: new Date(now + 1_000).toISOString() };
    },
}));
mock.module("../../src/lib/canvas/canvas-drawing-storage", () => ({ loadCanvasDrawingPreview: async () => null }));
mock.module("../../src/services/image-storage", () => ({ resolveImageUrl: async () => "" }));
const { act } = await import("react");
const { createRoot } = await import("react-dom/client");
const { useResolvedCanvasResourceReferences, reportResourcePreviewFailure } = await import("../../src/components/canvas/use-resolved-canvas-resource-references");
const references = [{ id: "image", kind: "image" as const, name: "image", storageKey: "resource:image", previewUrl: "https://example.test/fallback" }];
function Preview() {
    const resolved = useResolvedCanvasResourceReferences(references);
    return <img src={resolved[0].previewUrl} />;
}
afterAll(() => {
    clock.mockRestore();
    mock.restore();
    dom.happyDOM.abort();
});

test("重复引用合并请求，到期自动刷新，错误强刷有冷却，失败保留回退地址", async () => {
    const host = document.createElement("div");
    const root = createRoot(host);
    try {
        await act(async () =>
            root.render(
                <>
                    <Preview />
                    <Preview />
                </>,
            ),
        );
        expect(calls).toEqual([false]);
        expect(host.querySelector("img")?.src).toBe("https://example.test/preview-1");
        now += 1_001;
        await act(async () => {
            await Bun.sleep(1_100);
        });
        expect(calls).toEqual([false, false]);
        await act(async () => reportResourcePreviewFailure("https://example.test/preview-2"));
        expect(calls).toEqual([false, false, true]);
        expect(host.querySelector("img")?.src).toBe("https://example.test/preview-3");
        await act(async () => reportResourcePreviewFailure("https://example.test/preview-3"));
        expect(calls).toHaveLength(3);
        now += 30_001;
        fail = true;
        await act(async () => {
            await Bun.sleep(1_100);
        });
        expect(host.querySelector("img")?.src).toBe("https://example.test/fallback");
    } finally {
        await act(async () => root.unmount());
    }
});

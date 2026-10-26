import { afterEach, beforeEach, expect, spyOn, test } from "bun:test";

import { mountAuthVideo, preloadAuthVideo } from "../src/services/auth-video";
import { useAppearanceStore } from "../src/stores/use-appearance-store";

const source = "https://i.yourimageshare.com/ebwcGZCYli.mp4";
const originalAppearance = useAppearanceStore.getState().appearance;
const originals = new Map<string, PropertyDescriptor | undefined>();
const videos: FakeVideo[] = [];
const timers = new Map<number, { callback: () => void; delay: number }>();
let reducedMotion = false;
let timerID = 0;
let dispose: (() => void) | undefined;
let timeoutSpy: ReturnType<typeof spyOn>;
let clearTimeoutSpy: ReturnType<typeof spyOn>;

class FakeVideo extends EventTarget {
    src = "";
    poster = "";
    error: object | null = null;
    muted = false;
    defaultMuted = false;
    loop = false;
    playsInline = false;
    preload = "";
    readyState = 0;
    currentTime = 0;
    duration = 10;
    bufferEnd = 0;
    loadCount = 0;
    playCount = 0;
    pauseCount = 0;
    rejectsPlayback = false;
    removed = false;
    className = "";
    get buffered() {
        return { length: this.bufferEnd ? 1 : 0, start: () => 0, end: () => this.bufferEnd };
    }
    load() { this.loadCount++; }
    play() { this.playCount++; return this.rejectsPlayback ? Promise.reject(new Error("autoplay denied")) : Promise.resolve(); }
    pause() { this.pauseCount++; }
    remove() { this.removed = true; }
    removeAttribute(name: string) { if (name === "src") this.src = ""; }
    setAttribute() {}
}

function setGlobal(key: string, value: unknown) {
    originals.set(key, Object.getOwnPropertyDescriptor(globalThis, key));
    Object.defineProperty(globalThis, key, { configurable: true, writable: true, value });
}

function mount(poster?: string, url = source) {
    let attached: FakeVideo | undefined;
    dispose = mountAuthVideo({ append: (video: FakeVideo) => { attached = video; } } as unknown as HTMLElement, url, poster);
    return attached!;
}

function fireTimer(delay: number) {
    const [id, timer] = [...timers].find(([, timer]) => timer.delay === delay)!;
    timers.delete(id);
    timer.callback();
}

beforeEach(() => {
    videos.length = 0;
    timers.clear();
    reducedMotion = false;
    setGlobal("window", { matchMedia: () => ({ matches: reducedMotion }) });
    setGlobal("document", { createElement: () => { const video = new FakeVideo(); videos.push(video); return video; } });
    timeoutSpy = spyOn(globalThis, "setTimeout").mockImplementation(((callback: () => void, delay: number) => {
        const id = ++timerID;
        timers.set(id, { callback, delay });
        return id;
    }) as typeof setTimeout);
    clearTimeoutSpy = spyOn(globalThis, "clearTimeout").mockImplementation(((id: number) => { timers.delete(id); }) as typeof clearTimeout);
    useAppearanceStore.setState({ appearance: { ...originalAppearance, authVideoUrl: source, authVideoAutoplay: true } });
});

afterEach(() => {
    dispose?.();
    dispose = undefined;
    useAppearanceStore.setState({ appearance: { ...originalAppearance, authVideoAutoplay: false } });
    preloadAuthVideo();
    useAppearanceStore.setState({ appearance: originalAppearance });
    timeoutSpy.mockRestore();
    clearTimeoutSpy.mockRestore();
    for (const [key, descriptor] of originals) {
        if (descriptor) Object.defineProperty(globalThis, key, descriptor);
        else Reflect.deleteProperty(globalThis, key);
    }
    originals.clear();
});

test("repeated login intent loads once and navigation reuses the same paused video and exact URL", () => {
    preloadAuthVideo();
    preloadAuthVideo();
    expect(videos).toHaveLength(1);
    const prepared = videos[0];
    expect(prepared.src).toBe(source);
    expect(prepared.preload).toBe("auto");
    expect(prepared.playCount).toBe(0);
    expect(prepared.muted && prepared.defaultMuted && prepared.playsInline && prepared.loop).toBe(true);
    expect(mount()).toBe(prepared);
    expect(prepared.loadCount).toBe(1);
    expect(timers.size).toBe(0);
});

test("the first frame stays paused until buffered data is available, then starts only once", () => {
    const video = mount("/cover.webp");
    expect(video.poster).toBe("/cover.webp");
    video.readyState = 2;
    video.bufferEnd = 0.1;
    video.dispatchEvent(new Event("loadeddata"));
    expect(video.playCount).toBe(0);
    video.bufferEnd = 1.2;
    video.dispatchEvent(new Event("progress"));
    video.dispatchEvent(new Event("canplaythrough"));
    expect(video.playCount).toBe(1);
    expect(timers.size).toBe(0);
});

test("a short video can start when its whole duration is buffered", () => {
    const video = mount();
    video.duration = 0.5;
    video.readyState = 2;
    video.bufferEnd = 0.5;
    video.dispatchEvent(new Event("loadeddata"));
    expect(video.playCount).toBe(1);
});

test("slow loading does not start without a frame and waits at most 1.5 seconds after it arrives", () => {
    const video = mount();
    expect(timers.size).toBe(0);
    video.readyState = 2;
    video.dispatchEvent(new Event("loadeddata"));
    fireTimer(1_500);
    expect(video.playCount).toBe(1);
});

test("unmount cancels pending playback and releases the decoder", () => {
    const video = mount();
    video.readyState = 2;
    video.dispatchEvent(new Event("loadeddata"));
    dispose!();
    dispose = undefined;
    video.dispatchEvent(new Event("canplaythrough"));
    expect(video.playCount).toBe(0);
    expect(video.pauseCount).toBe(1);
    expect(video.src).toBe("");
    expect(video.removed).toBe(true);
    expect(timers.size).toBe(0);
});

test("unused preloads expire and a changed appearance never reuses the old media", () => {
    preloadAuthVideo();
    const old = videos[0];
    const video = mount(undefined, "/new-brand.mp4");
    expect(video).not.toBe(old);
    expect(old.src).toBe("");
    dispose!();
    dispose = undefined;
    preloadAuthVideo();
    const unused = videos.at(-1)!;
    fireTimer(30_000);
    expect(unused.src).toBe("");
    expect(unused.removed).toBe(true);
});

test("reduced motion and disabled autoplay skip preloading and clear stale buffers", () => {
    reducedMotion = true;
    preloadAuthVideo();
    expect(videos).toHaveLength(0);
    reducedMotion = false;
    preloadAuthVideo();
    useAppearanceStore.setState({ appearance: { ...originalAppearance, authVideoAutoplay: false } });
    preloadAuthVideo();
    expect(videos).toHaveLength(1);
    expect(videos[0].src).toBe("");
    expect(timers.size).toBe(0);
});

test("an autoplay rejection keeps the frame without retrying or an unhandled rejection", async () => {
    const video = mount();
    video.rejectsPlayback = true;
    video.readyState = 4;
    video.dispatchEvent(new Event("canplaythrough"));
    await Promise.resolve();
    video.dispatchEvent(new Event("progress"));
    expect(video.playCount).toBe(1);
    expect(video.src).toBe(source);
});

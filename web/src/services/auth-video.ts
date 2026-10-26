import { useAppearanceStore } from "@/stores/use-appearance-store";

let prepared: { video: HTMLVideoElement; source: string; expiry: ReturnType<typeof setTimeout> } | undefined;

function createVideo(source: string) {
    const video = document.createElement("video");
    video.muted = true;
    video.defaultMuted = true;
    video.loop = true;
    video.playsInline = true;
    video.preload = "auto";
    video.src = source;
    video.load();
    return video;
}

function releaseVideo(video: HTMLVideoElement) {
    video.pause();
    video.remove();
    video.removeAttribute("src");
    video.load();
}

function clearPreparedVideo() {
    if (!prepared) return;
    clearTimeout(prepared.expiry);
    releaseVideo(prepared.video);
    prepared = undefined;
}

/** 只在登录意图出现时加载；暂停的元素保留首帧和缓冲，跳转后直接复用。 */
export function preloadAuthVideo() {
    if (typeof document === "undefined") return;
    const { authVideoUrl, authVideoAutoplay } = useAppearanceStore.getState().appearance;
    if (!authVideoUrl || !authVideoAutoplay || window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
        clearPreparedVideo();
        return;
    }
    if (prepared?.source === authVideoUrl && !prepared.video.error) return;
    clearPreparedVideo();
    prepared = { video: createVideo(authVideoUrl), source: authVideoUrl, expiry: setTimeout(clearPreparedVideo, 30_000) };
}

export function mountAuthVideo(host: HTMLElement, source: string, poster?: string) {
    let video: HTMLVideoElement;
    if (prepared?.source === source && !prepared.video.error) {
        clearTimeout(prepared.expiry);
        video = prepared.video;
        prepared = undefined;
    } else {
        clearPreparedVideo();
        video = createVideo(source);
    }
    video.className = "absolute inset-0 size-full object-cover";
    video.setAttribute("aria-hidden", "true");
    if (poster) video.poster = poster;
    host.append(video);

    let started = false;
    let fallback: ReturnType<typeof setTimeout> | undefined;
    let allowShortBuffer = false;
    const startWhenBuffered = () => {
        if (started || video.readyState < 2) return;
        // 首帧先保持静止；优先等至少一秒缓冲，慢网络最多额外等待 1.5 秒。
        fallback ??= setTimeout(() => {
            allowShortBuffer = true;
            startWhenBuffered();
        }, 1_500);
        let bufferedAhead = 0;
        for (let index = 0; index < video.buffered.length; index++) {
            if (video.buffered.start(index) <= video.currentTime && video.buffered.end(index) > video.currentTime) {
                bufferedAhead = video.buffered.end(index) - video.currentTime;
                break;
            }
        }
        const target = Math.min(1, video.duration || 1);
        if (!allowShortBuffer && video.readyState < 4 && bufferedAhead < target) return;
        started = true;
        clearTimeout(fallback);
        // 自动播放被浏览器拒绝时保留封面或首帧，不循环重试。
        void video.play().catch(() => undefined);
    };
    const events = ["loadeddata", "progress", "canplay", "canplaythrough"] as const;
    for (const event of events) video.addEventListener(event, startWhenBuffered);
    startWhenBuffered();

    return () => {
        clearTimeout(fallback);
        for (const event of events) video.removeEventListener(event, startWhenBuffered);
        releaseVideo(video);
    };
}

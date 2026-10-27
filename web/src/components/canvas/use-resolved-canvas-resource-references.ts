import { useEffect, useMemo, useState } from "react";

import type { CanvasResourceReference } from "@/lib/canvas/canvas-resource-references";
import { loadCanvasDrawingPreview } from "@/lib/canvas/canvas-drawing-storage";
import { getResourceAccess, resolveResourceAccessURL, resourceIdFromStorageKey } from "@/services/api/resources";
import { resolveImageUrl } from "@/services/image-storage";

type ResolvedPreview = {
    identity: string;
    url: string;
};

type CachedPreview = {
    promise: Promise<{ url: string; refreshAt?: number }>;
    expiresAt: number;
    url?: string;
};

// 签名地址的有效期由 getResourceAccess 的缓存判断；这里只合并短时间内同一素材的重复解析。
// 不能永久持有：地址过期后重新打开提示词面板仍会拿到旧地址，@ 缩略图就显示空白。
const STORAGE_PREVIEW_CACHE_MS = 60_000;
// 同一素材加载失败后强制换新地址的最短间隔，避免坏文件反复请求后端。
const PREVIEW_REFRESH_COOLDOWN_MS = 30_000;

const previewPromiseCache = new Map<string, CachedPreview>();
const forcedRefreshAt = new Map<string, number>();
const pendingForceRefresh = new Set<string>();
const previewFailureListeners = new Set<() => void>();

/** 缩略图加载失败时调用：作废对应地址，并让正在使用它的提示词编辑器换新签名地址。 */
export function reportResourcePreviewFailure(url: string) {
    if (!url) return;
    const now = Date.now();
    let refreshed = false;
    previewPromiseCache.forEach((entry, identity) => {
        if (entry.url !== url) return;
        previewPromiseCache.delete(identity);
        if (now - (forcedRefreshAt.get(identity) ?? 0) < PREVIEW_REFRESH_COOLDOWN_MS) return;
        forcedRefreshAt.set(identity, now);
        pendingForceRefresh.add(identity);
        refreshed = true;
    });
    if (refreshed) previewFailureListeners.forEach((listener) => listener());
}
export function useResolvedCanvasResourceReferences(references: CanvasResourceReference[], options?: { projectId?: string }) {
    const projectId = options?.projectId;
    const [refreshEpoch, setRefreshEpoch] = useState(0);
    useEffect(() => {
        const listener = () => setRefreshEpoch((value) => value + 1);
        previewFailureListeners.add(listener);
        return () => {
            previewFailureListeners.delete(listener);
        };
    }, []);
    const requests = useMemo(
        () => references.flatMap((reference) => {
            const identity = previewIdentity(reference, projectId);
            return identity ? [{ reference, identity }] : [];
        }),
        [projectId, references],
    );
    const [resolvedById, setResolvedById] = useState<Record<string, ResolvedPreview>>({});
    const [refreshVersion, setRefreshVersion] = useState(0);

    useEffect(() => {
        if (!requests.length) return;
        let cancelled = false;
        let refreshTimer: ReturnType<typeof setTimeout> | undefined;
        void Promise.all(
            requests.map(async ({ reference, identity }) => ({ id: reference.id, identity, ...await resolveReferencePreview(reference, identity, projectId) })),
        ).then((resolved) => {
            if (cancelled) return;
            setResolvedById((current) => {
                let changed = false;
                const next = { ...current };
                resolved.forEach(({ id, identity, url }) => {
                    if (!url || (current[id]?.identity === identity && current[id]?.url === url)) return;
                    next[id] = { identity, url };
                    changed = true;
                });
                return changed ? next : current;
            });
            const nextRefresh = Math.min(...resolved.map((item) => item.refreshAt ?? Number.POSITIVE_INFINITY));
            if (Number.isFinite(nextRefresh)) refreshTimer = setTimeout(() => setRefreshVersion((version) => version + 1), Math.max(1_000, nextRefresh - Date.now()));
        });
        return () => {
            cancelled = true;
            if (refreshTimer) clearTimeout(refreshTimer);
        };
    }, [projectId, requests, refreshVersion, refreshEpoch]);

    return useMemo(
        () => references.map((reference) => {
            const identity = previewIdentity(reference, projectId);
            const resolved = identity ? resolvedById[reference.id] : undefined;
            return resolved?.identity === identity && resolved.url !== reference.previewUrl ? { ...reference, previewUrl: resolved.url } : reference;
        }),
        [projectId, references, resolvedById],
    );
}

function previewIdentity(reference: CanvasResourceReference, projectId?: string) {
    // 只有当 drawingId 和 projectId 同时存在时才返回绘图身份,避免不完整的缓存键导致运行时错误
    if (reference.drawingId) {
        if (!projectId) return "";
        return `drawing:${projectId}:${reference.drawingId}:${reference.drawingRevision || 0}`;
    }
    const storageKey = reference.kind === "video" ? reference.previewStorageKey : reference.storageKey;
    if (!storageKey || !["image", "video", "character"].includes(reference.kind)) return "";
    return `${reference.kind}:${storageKey}`;
}

function resolveReferencePreview(reference: CanvasResourceReference, identity: string, projectId?: string): Promise<{ url: string; refreshAt?: number }> {
    const cached = previewPromiseCache.get(identity);
    if (cached && cached.expiresAt > Date.now()) return cached.promise;
    const fallback = reference.previewUrl || "";
    const storageKey = reference.kind === "video" ? reference.previewStorageKey : reference.storageKey;
    const forceRefresh = pendingForceRefresh.delete(identity);
    let resolving: Promise<{ url: string; refreshAt?: number }>;
    if (reference.drawingId && projectId) {
        resolving = loadCanvasDrawingPreview(projectId, reference.drawingId).then(async (preview) => ({ url: preview ? await blobToDataUrl(preview) : "" }));
    } else if (resourceIdFromStorageKey(storageKey)) {
        resolving = getResourceAccess(storageKey, "display", "original", "", { forceRefresh }).then((access) => ({ url: resolveResourceAccessURL(access.url), refreshAt: Date.parse(access.refreshAt) }));
    } else {
        resolving = resolveImageUrl(storageKey, "", { cacheMiss: true }).then((url) => ({ url }));
    }
    const entry: CachedPreview = { promise: Promise.resolve({ url: "" }), expiresAt: Date.now() + STORAGE_PREVIEW_CACHE_MS };
    entry.promise = resolving
        .catch(() => ({ url: "", refreshAt: Date.now() + 30_000 }))
        .then((result) => {
            if (!result.url) {
                if (previewPromiseCache.get(identity) === entry) previewPromiseCache.delete(identity);
                return { ...result, url: fallback };
            }
            entry.url = result.url;
            if (Number.isFinite(result.refreshAt)) entry.expiresAt = Math.min(entry.expiresAt, result.refreshAt!);
            return result;
        });
    previewPromiseCache.set(identity, entry);
    return entry.promise;
}

function blobToDataUrl(blob: Blob) {
    return new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onerror = () => reject(reader.error || new Error("读取绘图预览失败"));
        reader.onload = () => resolve(typeof reader.result === "string" ? reader.result : "");
        reader.readAsDataURL(blob);
    });
}

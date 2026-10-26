import { useEffect, useState } from "react";
import { agentAvatarClipForState, agentAvatarStateForRunStatus, DEFAULT_CANVAS_APPEARANCE, isAgentAvatarBundleManifest, type AgentAvatarBundleManifest } from "@/lib/canvas/agent-appearance";
import { agentAvatarBundleFileURL } from "@/services/api/appearance";
import { useAppearanceStore } from "@/stores/use-appearance-store";

export function useAgentAvatar(status = "idle") {
    const appearance = useAppearanceStore((state) => state.appearance.canvas) || DEFAULT_CANVAS_APPEARANCE;
    const appearanceResolved = useAppearanceStore((state) => state.resolved);
    const appearanceRevision = useAppearanceStore((state) => state.appearance.revision);
    const live = appearance.avatarType === "live2d" && Boolean(appearance.live2dResourceId && appearance.live2dEntry);
    const avatarState = agentAvatarStateForRunStatus(status);
    const bundleManifestURL = appearance.avatarBundleResourceId ? agentAvatarBundleFileURL(appearance.avatarBundleResourceId, "avatar.json", false, appearanceRevision) : "/doll/avatar.json";
    const [bundleManifestState, setBundleManifestState] = useState<{ url: string; manifest: AgentAvatarBundleManifest | null } | null>(null);
    useEffect(() => {
        if (!appearanceResolved || appearance.avatarType !== "bundle") {
            setBundleManifestState(null);
            return;
        }
        const controller = new AbortController();
        void (async () => {
            try {
                const response = await fetch(bundleManifestURL, { credentials: "include", signal: controller.signal });
                if (!response.ok) throw new Error(`无法读取 Agent 动画包（HTTP ${response.status}）`);
                const manifest: unknown = await response.json();
                if (!isAgentAvatarBundleManifest(manifest)) throw new Error("avatar.json 格式无效");
                setBundleManifestState({ url: bundleManifestURL, manifest });
            } catch {
                if (!controller.signal.aborted) setBundleManifestState({ url: bundleManifestURL, manifest: null });
            }
        })();
        return () => controller.abort();
    }, [appearanceResolved, appearance.avatarType, bundleManifestURL]);
    const currentBundleManifestState = bundleManifestState?.url === bundleManifestURL ? bundleManifestState : null;
    const bundleManifest = currentBundleManifestState?.manifest ?? null;
    const bundleManifestLoading = appearance.avatarType === "bundle" && currentBundleManifestState === null;
    const clip = bundleManifest ? agentAvatarClipForState(appearance, avatarState, bundleManifest) : "";
    const clipFile = bundleManifest?.clips[clip] || "";
    const mediaURL = appearance.avatarType === "bundle" && clipFile ? (appearance.avatarBundleResourceId ? agentAvatarBundleFileURL(appearance.avatarBundleResourceId, clipFile, false, appearanceRevision) : `/doll/${encodeURIComponent(clipFile)}`) : "";
    const [mediaError, setMediaError] = useState(false);
    useEffect(() => setMediaError(false), [mediaURL]);
    return {
        appearance,
        live,
        avatarState,
        mediaURL,
        mediaError,
        onMediaError: () => setMediaError(true),
        loading: bundleManifestLoading || !appearanceResolved,
    };
}

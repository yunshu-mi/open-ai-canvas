import { create } from "zustand";
import type { CanvasResourceReference } from "@/lib/canvas/canvas-resource-references";
import type { CloudAgentMessagePresentation } from "@/services/cloud-agent-conversations";
import type { CanvasNodeData } from "@/types/canvas";

export type AgentPrefillRequest = CloudAgentMessagePresentation & { id: number; text: string; autoSubmit?: boolean; requiresVision?: boolean };

/**
 * 当前页面提供给平台级 Agent 的画布上下文。只有画布编辑器会发布它；
 * 离开画布时清空，此后的对话轮次不再绑定该画布。
 */
export type AgentCanvasContext = {
    canvasId: string;
    domainProjectId?: string;
    nodeCount: number;
    selectedNodeIds: string[];
    references: CanvasResourceReference[];
    canvasNodes?: readonly CanvasNodeData[];
    runningNodeId?: string | null;
    prefillRequest?: AgentPrefillRequest | null;
    onFocusNode?: (nodeId: string) => void;
};

type GlobalAgentStore = {
    /** 平台唯一 Agent 面板的开合状态，跨路由保持。 */
    open: boolean;
    launcherHost: HTMLDivElement | null;
    canvas: AgentCanvasContext | null;
    setLauncherHost: (host: HTMLDivElement | null) => void;
    openAgent: () => void;
    closeAgent: () => void;
    setCanvas: (canvas: AgentCanvasContext | null) => void;
    reset: () => void;
};

export const useGlobalAgentStore = create<GlobalAgentStore>()((set) => ({
    open: false,
    launcherHost: null,
    canvas: null,
    setLauncherHost: (launcherHost) => set({ launcherHost }),
    openAgent: () => set({ open: true }),
    closeAgent: () => set({ open: false }),
    setCanvas: (canvas) => set({ canvas }),
    reset: () => set({ open: false, canvas: null }),
}));

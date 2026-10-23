import { afterEach, describe, expect, test } from "bun:test";
import { useThemeStore } from "../src/stores/use-theme-store";
import { useCanvasThemeStore } from "../src/stores/canvas/use-canvas-theme-store";
import { useCanvasStore } from "../src/stores/canvas/use-canvas-store";
import { canvasAppearanceForPreset } from "../src/lib/canvas/canvas-appearance";

const workspaceBefore = useThemeStore.getState().theme;
const canvasBefore = useCanvasThemeStore.getState().theme;
afterEach(() => {
    useThemeStore.setState({ theme: workspaceBefore });
    useCanvasThemeStore.setState({ theme: canvasBefore, active: false });
});

describe("canvas theme ownership", () => {
    test("defaults to black while allowing an explicit light theme", () => {
        expect(useCanvasThemeStore.getInitialState().theme).toBe("dark");
        useCanvasThemeStore.getState().setTheme("light");
        expect(useCanvasThemeStore.getState().theme).toBe("light");
    });
    test("editing the canvas theme never mutates the workspace preference", () => {
        useThemeStore.setState({ theme: "light" });
        useCanvasThemeStore.setState({ active: true });
        useCanvasThemeStore.getState().setTheme("dark");
        expect(useThemeStore.getState().theme).toBe("light");
        expect(useCanvasThemeStore.getState().theme).toBe("dark");
        useCanvasThemeStore.setState({ active: false });
        expect(useThemeStore.getState().theme).toBe("light");
    });
    test("workspace switches cannot change canvas preference", () => {
        useCanvasThemeStore.setState({ theme: "light" });
        useThemeStore.getState().setTheme("dark");
        expect(useCanvasThemeStore.getState().theme).toBe("light");
    });
    test("new canvases use upstream workspace presets without changing the active canvas theme", () => {
        useThemeStore.setState({ theme: "light" });
        useCanvasThemeStore.setState({ theme: "dark", active: false });
        const lightId = useCanvasStore.getState().createProject("浅色画布");
        expect(useCanvasStore.getState().openProject(lightId)?.appearance).toEqual(canvasAppearanceForPreset("light"));
        expect(useCanvasThemeStore.getState().theme).toBe("dark");

        useCanvasThemeStore.setState({ theme: "dark", active: true });
        useThemeStore.setState({ theme: "dark" });
        const darkId = useCanvasStore.getState().createProject("深色画布");
        expect(useCanvasStore.getState().openProject(darkId)?.appearance).toEqual(canvasAppearanceForPreset("dark"));
        useCanvasStore.getState().deleteProjects([lightId, darkId]);
    });
    test("rejects malformed canvas themes", () => {
        useCanvasThemeStore.setState({ theme: "light", active: true });
        useCanvasThemeStore.getState().setTheme("invalid" as "light");
        expect(useCanvasThemeStore.getState().theme).toBe("light");
    });
    test("persistence source contract excludes transient route state", async () => {
        const source = await Bun.file(new URL("../src/stores/canvas/use-canvas-theme-store.ts", import.meta.url)).text();
        expect(source).toContain('partialize: ({ theme }) => ({ theme })');
    });
});

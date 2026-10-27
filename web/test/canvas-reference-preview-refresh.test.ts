import { expect, test } from "bun:test";
import { resolve } from "node:path";

test("引用预览同时支持到期刷新和加载失败后强制换签名", () => {
    const result = Bun.spawnSync([process.execPath, "test", "./test/fixtures/canvas-reference-preview-refresh.harness.tsx"], { cwd: resolve(import.meta.dir, ".."), timeout: 15000 });
    expect(result.exitCode, new TextDecoder().decode(result.stdout) + new TextDecoder().decode(result.stderr)).toBe(0);
});

#!/usr/bin/env bash
set -Eeuo pipefail

cd "$(dirname "$0")/.."
temporary="$(mktemp -d)"
trap 'rm -rf "$temporary"' EXIT
workflow=.github/workflows/publish-images.yml

# 校验事件入口，避免普通提交或手动运行绕过正式标签限制。
tr -d '\r' < "$workflow" | sed -n '/^on:/,/^permissions:/p' > "$temporary/triggers"
if grep -Eq 'branches:|workflow_dispatch:|workflow_call:|pull_request:|release:' "$temporary/triggers"; then
    echo 'Image publishing must only be triggered by stable tag pushes' >&2
    exit 1
fi
grep -Fxq '  push:' "$temporary/triggers"
grep -Fxq '    tags:' "$temporary/triggers"
[[ "$(grep -c '^      - ' "$temporary/triggers")" == 2 ]]
grep -Fxq '      - "v[0-9]+.[0-9]+.[0-9]+"' "$temporary/triggers"
grep -Fxq '      - "v[0-9]+.[0-9]+.[0-9]+.[0-9]+"' "$temporary/triggers"

# 直接执行工作流里的真实发布校验，隔离版本文件与输出。
tr -d '\r' < "$workflow" | awk '
    /^  prepare:/ { prepare=1 }
    prepare && /^        run: \|/ { script=1; next }
    script { if (!/^          /) exit; sub(/^          /, ""); print }
' > "$temporary/validate.sh"
[[ -s "$temporary/validate.sh" ]]
git() { printf '2026-01-01T00:00:00Z\n'; }
export -f git

validate() {
    local ref="$1" tag="$2" version="$3" heading="$4"
    printf '%s\n' "$version" > "$temporary/VERSION"
    printf '## %s\n' "$heading" > "$temporary/CHANGELOG.md"
    : > "$temporary/output"
    (
        cd "$temporary"
        GITHUB_REF="$ref" GITHUB_REF_NAME="$tag" REPOSITORY_OWNER=YUNSHU-MI \
            GITHUB_OUTPUT="$temporary/output" bash -e -o pipefail "$temporary/validate.sh"
    )
}

for tag in v1.2.3 v1.6.9.2 v12.34.56 v12.34.56.78; do
    validate "refs/tags/$tag" "$tag" "$tag" "$tag"
    grep -Fxq "version=$tag" "$temporary/output"
    grep -Fxq "image_tag=${tag#v}" "$temporary/output"
    grep -Fxq 'owner=yunshu-mi' "$temporary/output"
done

reject() {
    if validate "$@" > "$temporary/rejected.log" 2>&1; then
        echo "Unexpectedly accepted release: $*" >&2
        exit 1
    fi
}

reject refs/heads/main main v1.2.3 v1.2.3
reject refs/heads/v1.2.3 v1.2.3 v1.2.3 v1.2.3
for tag in v1.2.3-rc.1 v1.2.3-beta v1.2.3-preview.1 v1.2.3.rc1 v1.2 v1.2.3.4.5 1.2.3 vlatest; do
    reject "refs/tags/$tag" "$tag" "$tag" "$tag"
done
reject refs/tags/v1.2.3 v1.2.3 v1.2.4 v1.2.3
reject refs/tags/v1.2.3 v1.2.3 v1.2.3 v1.2.4
printf 'PASS: stable tags only; branches, prereleases and inconsistent metadata rejected\n'

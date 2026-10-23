#!/usr/bin/env bash
set -Eeuo pipefail

cd "$(dirname "$0")/.."
temporary="$(mktemp -d)"
trap 'rm -rf "$temporary"' EXIT
unset REPOSITORY IMAGE_REPOSITORY COMPOSE_URL UPDATER_INSTALL_URL
export REPOSITORY_REF=v1.6.9.2 CANVAS_IMAGE_TAG=v1.6.9.2

# 只加载真实脚本的定义，不执行需要 root 和网络的 main。
load_installer() {
    source <(sed -e 's/\r$//' -e '/^main "\$@"$/d' "$1")
}

(
    load_installer scripts/install-server-image.sh
    [[ "$REPOSITORY" == yunshu-mi/open-ai-canvas ]]
    [[ "$IMAGE_REPOSITORY" == ghcr.io/yunshu-mi/open-ai-canvas ]]
    [[ "$COMPOSE_URL" == https://raw.githubusercontent.com/yunshu-mi/open-ai-canvas/v1.6.9.2/docker-compose.deploy.yml ]]
    [[ "$UPDATER_INSTALL_URL" == https://raw.githubusercontent.com/yunshu-mi/open-ai-canvas/v1.6.9.2/scripts/install-host-updater.sh ]]
    INSTALL_DIR="$temporary/default"
    mkdir -p "$INSTALL_DIR"
    prepare_environment
    for image in backend web yingce-agent; do
        grep -Fq "=ghcr.io/yunshu-mi/open-ai-canvas-${image}:1.6.9.2" "$INSTALL_DIR/.env"
    done
)
printf 'PASS: default repository and generated image references\n'

(
    REPOSITORY=another-owner/open-ai-canvas
    load_installer scripts/install-server-image.sh
    [[ "$IMAGE_REPOSITORY" == ghcr.io/another-owner/open-ai-canvas ]]
    [[ "$COMPOSE_URL" == https://raw.githubusercontent.com/another-owner/open-ai-canvas/v1.6.9.2/docker-compose.deploy.yml ]]
    [[ "$UPDATER_INSTALL_URL" == https://raw.githubusercontent.com/another-owner/open-ai-canvas/v1.6.9.2/scripts/install-host-updater.sh ]]
    INSTALL_DIR="$temporary/custom"
    mkdir -p "$INSTALL_DIR"
    # 用本地安装器替代下载，确认子进程收到相同的仓库和目录。
    curl() {
        cat > "$4" <<'INSTALLER'
set -eu
test "$REPOSITORY" = another-owner/open-ai-canvas
test -d "$INSTALL_DIR"
touch "$INSTALL_DIR/updater-installed"
INSTALLER
    }
    docker() { return 0; }
    install_host_updater
    [[ -f "$INSTALL_DIR/updater-installed" ]]
)
printf 'PASS: repository override reaches updater installer\n'

for repository in yunshu-mi/open-ai-canvas another-owner/open-ai-canvas; do
    (
        if [[ "$repository" == yunshu-mi/open-ai-canvas ]]; then
            unset REPOSITORY
        else
            REPOSITORY="$repository"
        fi
        load_installer scripts/install-host-updater.sh
        INSTALL_DIR="$temporary/${repository%%/*}"
        mkdir -p "$INSTALL_DIR"
        UPDATER_ENV="$INSTALL_DIR/updater.env"
        SOCKET_DIR="$INSTALL_DIR/socket"
        printf 'CANVAS_UPDATER_TOKEN=%064d\nPOSTGRES_DB=keep\n' 0 > "$INSTALL_DIR/.env"
        ensure_token
        grep -Fxq "CANVAS_UPDATER_REPOSITORY=$repository" "$UPDATER_ENV"
        grep -Fxq "CANVAS_UPDATER_INSTALL_DIR=$INSTALL_DIR" "$UPDATER_ENV"
        grep -Fxq 'POSTGRES_DB=keep' "$INSTALL_DIR/.env"
    )
done
printf 'PASS: updater service persists default and overridden repositories\n'

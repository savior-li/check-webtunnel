#!/bin/bash

set -e

REPO="savior-li/check-webtunnel"
BUILD_DIR="./build"
TAG=${1:-"v1.0.0"}
RELEASE_NAME=${2:-"Tor Bridge Collector v1.0.0"}
RELEASE_NOTES=${3:-"Tor Bridge Collector 首次发布\n\n支持功能:\n- WebTunnel 桥梁采集\n- HTTP/HTTPS/SOCKS5 代理支持\n- SQLite 数据持久化\n- 桥梁有效性验证\n- 多格式导出 (torrc/JSON)\n- 中英文双语界面"}

echo "========================================"
echo "  Tor Bridge Collector Release Script"
echo "========================================"
echo ""

if ! gh auth status &>/dev/null; then
    echo "错误: GitHub CLI 未登录"
    echo ""
    echo "请先登录 GitHub:"
    echo "  gh auth login"
    echo ""
    echo "或者设置 GH_TOKEN 环境变量:"
    echo "  export GH_TOKEN=your_github_token"
    exit 1
fi

echo "[1/4] 检查构建产物..."
if [ ! -d "$BUILD_DIR" ] || [ -z "$(ls -A $BUILD_DIR)" ]; then
    echo "构建产物不存在，正在构建..."
    chmod +x ./scripts/build.sh
    ./scripts/build.sh
else
    echo "构建产物已存在，跳过构建"
    ls -lh "$BUILD_DIR"
fi

echo ""
echo "[2/4] 创建 Git tag: $TAG"
git tag -a "$TAG" -m "Release $TAG" 2>/dev/null || echo "Tag $TAG 已存在"

echo ""
echo "[3/4] 推送 tag 到远程..."
git push origin "$TAG"

echo ""
echo "[4/4] 创建 GitHub Release..."
gh release create "$TAG" \
    --title "$RELEASE_NAME" \
    --notes "$RELEASE_NOTES" \
    --repo "$REPO"

echo ""
echo "上传二进制文件到 Release..."
for BINARY in "$BUILD_DIR"/*; do
    if [ -f "$BINARY" ]; then
        FILENAME=$(basename "$BINARY")
        echo "  上传: $FILENAME"
        gh release upload "$TAG" "$BINARY" --repo "$REPO"
    fi
done

echo ""
echo "========================================"
echo "  Release 完成!"
echo "========================================"
echo ""
echo "Release 地址: https://github.com/$REPO/releases/tag/$TAG"

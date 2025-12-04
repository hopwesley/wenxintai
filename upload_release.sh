#!/usr/bin/env bash
# 只上传：release/static、release/wxt_lnx，以及 promote_static.sh 到远端，不做任何远端操作

set -e

KEY="$HOME/Documents/conf/tx_ninja.pem"
HOST="root@43.138.18.5"
REMOTE_DIR="/root/wenxintai/upload"

STATIC_LOCAL="./release/static"
BIN_LOCAL="./release/wxt_lnx"
PROMOTE_LOCAL="./promote_static.sh"

# 本地检查
[ -d "$STATIC_LOCAL" ] || { echo "缺少 $STATIC_LOCAL；先 make deploy/deploy-all"; exit 1; }
[ -f "$BIN_LOCAL" ]    || { echo "缺少 $BIN_LOCAL；先 make deploy-all"; exit 1; }
[ -f "$PROMOTE_LOCAL" ] || { echo "缺少 $PROMOTE_LOCAL；请把脚本放在项目根目录"; exit 1; }

# 上传静态目录（形成 $REMOTE_DIR/static）
scp -i "$KEY" -r "$STATIC_LOCAL" "$HOST:$REMOTE_DIR/"

# 上传可执行文件到 $REMOTE_DIR
scp -i "$KEY" "$BIN_LOCAL" "$HOST:$REMOTE_DIR/"

# 上传 promote_static.sh 到 $REMOTE_DIR
scp -i "$KEY" "$PROMOTE_LOCAL" "$HOST:$REMOTE_DIR/"

echo "✅ 上传完成：$HOST:$REMOTE_DIR （仅 scp，无远端操作）"
echo "   - static/ → $REMOTE_DIR/static"
echo "   - wxt_lnx → $REMOTE_DIR/wxt_lnx"
echo "   - promote_static.sh → $REMOTE_DIR/promote_static.sh"

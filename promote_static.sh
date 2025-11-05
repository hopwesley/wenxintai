#!/usr/bin/env bash
# 把 /root/wenxintai/static 提升为 /var/www/wenxintai/static
# 1) 备份旧版本  2) 覆盖同步  3) 修权限

set -euo pipefail

# 源与目标
SRC="/root/wenxintai/static"
DST="/var/www/wenxintai/static"
BACKUP_BASE="/var/www/wenxintai/backups"
TS=$(date +%Y%m%d_%H%M%S)

# 兼容你可能的拼写（若 /root/wenixntai/static 存在就优先用）
if [ -d "/root/wenixntai/static" ]; then
  SRC="/root/wenixntai/static"
fi

# 基本检查
[ -d "$SRC" ] || { echo "❌ 源目录不存在：$SRC"; exit 1; }

# 确保目标/备份目录存在
mkdir -p "$DST" "$BACKUP_BASE"

echo "➡️  源：$SRC"
echo "➡️  目标：$DST"
echo "➡️  备份目录：$BACKUP_BASE"

# 备份当前线上版本
if [ -n "$(ls -A "$DST" 2>/dev/null || true)" ]; then
  BK="$BACKUP_BASE/static_$TS"
  mkdir -p "$BK"
  echo "📦  备份当前 $DST -> $BK"
  cp -a "$DST/." "$BK/"
fi

# 同步新版（优先用 rsync，没有就用 cp）
if command -v rsync >/dev/null 2>&1; then
  echo "🔁 rsync 同步（含删除多余文件）..."
  rsync -a --delete "$SRC/" "$DST/"
else
  echo "🔁 cp -a 同步（不删除多余文件，若需要请手动清理）..."
  rm -rf "$DST/"*   # 如不想清空，可注释本行
  cp -a "$SRC/." "$DST/"
fi

# 修权限（nginx 用户可读）
echo "🔐 修权限..."
chown -R nginx:nginx /var/www/wenxintai
find /var/www/wenxintai -type d -exec chmod 755 {} \;
find /var/www/wenxintai -type f -exec chmod 644 {} \;

echo "✅ 完成。静态资源已更新到：$DST"

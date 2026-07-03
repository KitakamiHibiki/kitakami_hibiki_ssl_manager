#!/bin/bash
# 证书签发测试脚本 - 独立于项目，直接使用 lego + Let's Encrypt
# 用法: ./test_cert_issue.sh <域名> <邮箱>

set -e

DOMAIN="${1:-test.kitakamihibiki.top}"
EMAIL="${2:-test@test.com}"
OUT_DIR="./test_certs_$(date +%Y%m%d_%H%M%S)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${CYAN}[$(date +%H:%M:%S)]${NC} $1"; }
ok()   { echo -e "${GREEN}[OK]${NC} $1"; }
err()  { echo -e "${RED}[ERR]${NC} $1"; exit 1; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# ---------- 检查 lego ----------
if ! command -v lego &>/dev/null; then
  warn "lego 未安装，正在通过 go install 安装..."
  go install github.com/go-acme/lego/v4/cmd/lego@latest || err "安装 lego 失败，请确认已安装 Go"
  export PATH="$PATH:$(go env GOPATH)/bin"
  command -v lego &>/dev/null || err "lego 安装后仍不可用"
fi
ok "lego $(lego --version 2>&1 | head -1)"

mkdir -p "$OUT_DIR"

# ---------- ACME 注册 + 申请 ----------
log "向 Let's Encrypt 申请证书 for $DOMAIN ..."
CHALLENGE_OUTPUT=$(lego \
  --email "$EMAIL" \
  --domains "$DOMAIN" \
  --dns manual \
  --path "$OUT_DIR/.lego" \
  --accept-tos \
  run 2>&1) || true

# 提取 DNS 挑战信息
CHALLENGE_FQDN=$(echo "$CHALLENGE_OUTPUT" | grep -o "_acme-challenge\.[^ ]*" | head -1)
CHALLENGE_VALUE=$(echo "$CHALLENGE_OUTPUT" | grep -o 'with the following value: "[^"]*"' | head -1 | cut -d'"' -f2)

if [ -z "$CHALLENGE_FQDN" ] || [ -z "$CHALLENGE_VALUE" ]; then
  echo "$CHALLENGE_OUTPUT"
  err "未能获取 DNS 挑战值"
fi

ok "DNS 挑战值已获取"
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  主机记录: ${CHALLENGE_FQDN%.$DOMAIN}${NC}"
echo -e "${YELLOW}  完整域名: ${CHALLENGE_FQDN}${NC}"
echo -e "${YELLOW}  记录值:   ${CHALLENGE_VALUE}${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# ---------- 等待用户添加 DNS ----------
read -p "请在 DNS 中添加上述 TXT 记录，然后按回车继续... "

# ---------- 验证 DNS ----------
log "验证 DNS TXT 记录..."
MAX_RETRY=60
VERIFIED=false
for i in $(seq 1 $MAX_RETRY); do
  TXT_RESULT=$(powershell -Command "Resolve-DnsName -Name '$CHALLENGE_FQDN' -Type TXT -Server 8.8.8.8 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Strings" 2>/dev/null || echo "")
  if echo "$TXT_RESULT" | grep -qF "$CHALLENGE_VALUE"; then
    VERIFIED=true
    ok "DNS 验证通过 (${i}s)"
    break
  fi
  printf "\r  等待 DNS 传播... ${i}s / ${MAX_RETRY}s"
  sleep 1
done
echo ""

if [ "$VERIFIED" = false ]; then
  err "DNS TXT 记录未生效（等待${MAX_RETRY}秒）"
fi

# ---------- 完成挑战 ----------
log "向 Let's Encrypt 提交 DNS 验证..."
lego \
  --email "$EMAIL" \
  --domains "$DOMAIN" \
  --dns manual \
  --path "$OUT_DIR/.lego" \
  --accept-tos \
  run 2>&1 | tail -5

# ---------- 检查结果 ----------
CERT_DIR="$OUT_DIR/.lego/certificates"
if [ -f "$CERT_DIR/${DOMAIN}.crt" ]; then
  ok "证书签发成功！"
  echo ""
  echo -e "${GREEN}========================================${NC}"
  echo -e "${GREEN}  证书签发测试完成！${NC}"
  echo -e "${GREEN}  证书文件: ${CERT_DIR}/${DOMAIN}.crt${NC}"
  echo -e "${GREEN}  私钥文件: ${CERT_DIR}/${DOMAIN}.key${NC}"
  echo -e "${GREEN}  完整链:   ${CERT_DIR}/${DOMAIN}.issuer.crt${NC}"
  echo -e "${GREEN}  过期时间: $(openssl x509 -enddate -noout -in "$CERT_DIR/${DOMAIN}.crt" 2>/dev/null | cut -d= -f2 || echo "未知")${NC}"
  echo -e "${GREEN}========================================${NC}"
  
  # 复制到更方便的位置
  cp "$CERT_DIR/${DOMAIN}.crt" "$OUT_DIR/${DOMAIN}.fullchain.pem"
  cat "$CERT_DIR/${DOMAIN}.issuer.crt" >> "$OUT_DIR/${DOMAIN}.fullchain.pem" 2>/dev/null || true
  cp "$CERT_DIR/${DOMAIN}.key" "$OUT_DIR/${DOMAIN}.privkey.pem"
  ok "证书已复制到: $OUT_DIR/"
else
  err "证书签发失败，请检查上方输出"
fi

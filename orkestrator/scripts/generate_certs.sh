#!/bin/bash

# Получаем путь к корню проекта
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ORKESTRATOR_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$ORKESTRATOR_DIR")"
AGENT_DIR="$PROJECT_ROOT/agent"

# Создаем директорию для сертификатов на сервере
mkdir -p "$ORKESTRATOR_DIR/certs"

# IP адрес сервера
SERVER_IP="192.168.0.89"

# Генерируем приватный ключ
openssl genrsa -out "$ORKESTRATOR_DIR/certs/server.key" 2048

# Создаем конфигурационный файл для сертификата с SAN
cat > "$ORKESTRATOR_DIR/certs/cert.conf" <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = RU
ST = Moscow
L = Moscow
O = CalcServer
CN = $SERVER_IP

[v3_req]
keyUsage = digitalSignature, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names
basicConstraints = CA:FALSE

[alt_names]
IP.1 = $SERVER_IP
EOF

# Генерируем самоподписанный сертификат с SAN
openssl req -new -x509 -sha256 -key "$ORKESTRATOR_DIR/certs/server.key" \
    -out "$ORKESTRATOR_DIR/certs/server.crt" -days 365 \
    -config "$ORKESTRATOR_DIR/certs/cert.conf" -extensions v3_req

# Удаляем временный конфигурационный файл
rm "$ORKESTRATOR_DIR/certs/cert.conf"

# Создаем директорию для сертификатов на агенте
mkdir -p "$AGENT_DIR/certs"

# Копируем сертификат на агент (ключ остается только на сервере)
cp "$ORKESTRATOR_DIR/certs/server.crt" "$AGENT_DIR/certs/server.crt"

echo "Сертификаты созданы:"
echo "  - $ORKESTRATOR_DIR/certs/server.crt - сертификат сервера"
echo "  - $ORKESTRATOR_DIR/certs/server.key - приватный ключ сервера (только на сервере)"
echo "  - $AGENT_DIR/certs/server.crt - сертификат скопирован на агент"


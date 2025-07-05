#!/bin/bash
# generate_certs.sh
#
# This script generates self-signed certificates for mTLS:
# - A Root CA (ca.crt, ca.key)
# - A Server Certificate (server.crt, server.key) signed by the CA
# - A Client Certificate (client.crt, client.key) signed by the CA
#
# Usage:
#   ./generate_certs.sh
#
# Output:
#   ./certs/ca.crt
#   ./certs/ca.key
#   ./certs/server.crt
#   ./certs/server.key
#   ./certs/client.crt
#   ./certs/client.key

set -euo pipefail

CERTS_DIR="certs"
SERVER_CN="macvmorx" # Common Name for the orchestrator server
CLIENT_CN="macvmagt"      # Common Name for the agents

echo "Generating mTLS certificates in directory: $CERTS_DIR"
mkdir -p "$CERTS_DIR"
cd "$CERTS_DIR"

# 1. Generate Root CA key and certificate
echo "Generating Root CA..."
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt #-subj "/CN=MacVMOrxRootCA"

# 2. Generate Server Key and Certificate Signing Request (CSR)
echo "Generating Server Certificate (for orchestrator and agent command server)..."
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr #-subj "/CN=$SERVER_CN"

# Create a server certificate extension file
cat > server.ext <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = $SERVER_CN
IP.1 = 127.0.0.1
# Add more DNS names or IPs if your orchestrator/agents will be accessed by hostname/IP
# E.g., DNS.3 = my-orchestrator.example.com
# E.g., IP.2 = 192.168.1.100
EOF

# Sign the server certificate with the CA
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out server.crt -days 365 -sha256 -extfile server.ext

# 3. Generate Client Key and Certificate Signing Request (CSR)
echo "Generating Client Certificate (for agents sending heartbeats and orchestrator sending commands)..."
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr #-subj "/CN=$CLIENT_CN"

# Create a client certificate extension file
cat > client.ext <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = clientAuth
EOF

# Sign the client certificate with the CA
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out client.crt -days 365 -sha256 -extfile client.ext

echo "Certificates generated successfully in ./certs/"
echo "ca.crt: Root CA certificate (distribute to all components for trusting)"
echo "server.crt, server.key: Server certificate and key (for orchestrator's listener and agent's command listener)"
echo "client.crt, client.key: Client certificate and key (for agent's heartbeat sender and orchestrator's command sender)"

cd ..
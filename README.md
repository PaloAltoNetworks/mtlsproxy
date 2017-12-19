# mtlsproxy

MTLS proxy is a simple proxy service that can be run as a sidecar of an unsecure service.
It will allow Aporeto system cert issued bearer to connect to the backend service.

This is moslty use for securing jaeger right now.

## example

```bash
export MTLSPROXY_CID_URL="https://localhost:4444"
export MTLSPROXY_CID_CACERT="$CERTS_FOLDER/ca-chain-system.pem"
export MTLSPROXY_LISTEN=":19443"
export MTLSPROXY_ISSUING_CERT_PASS="aporeto"
export MTLSPROXY_PUBLIC_CERT_PASS="public"
export MTLSPROXY_IP_ALT_NAME="127.0.0.1"
export MTLSPROXY_DNS_ALT_NAME="localhost"
export MTLSPROXY_CLIENT_CERT_PASS="aporeto"
export MTLSPROXY_BACKEND="http://127.0.0.1:16686"

./mtlsproxy
```
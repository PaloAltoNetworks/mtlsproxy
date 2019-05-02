# mtlsproxy

MTLS proxy is a simple proxy service that can be run as a sidecar of an unsecure service and provide mutual TLS authentication.

## example

```bash
export MTLSPROXY_LISTEN=":19443"
export MTLSPROXY_LOG_LEVEL="info"
export MTLSPROXY_LOG_FORMAT="console"
export MTLSPROXY_LISTEN=":19443"
export MTLSPROXY_BACKEND="http://127.0.0.1:16686"
export MTLSPROXY_BACKEND_NAME="jaeger"
export MTLSPROXY_CERT="$CERTS_FOLDER/public-cert.pem"
export MTLSPROXY_CERT_KEY="$CERTS_FOLDER/public-key.pem"
export MTLSPROXY_CERT_KEY_PASS="public"
export MTLSPROXY_CLIENTS_CA="$CERTS_FOLDER/ca-auditers-cert.pem"

./mtlsproxy
```

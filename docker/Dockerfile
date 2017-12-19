FROM scratch

LABEL \
    contact=info@aporeto.com \
    vendor=aporeto \
    application=aporeto \
    component=mtlsproxy \
    visibility=public

ADD app/mtlsproxy /mtlsproxy
ADD app/ca-certificates.pem /etc/ssl/certs/ca-certificates.pem

EXPOSE 443

ENTRYPOINT ["/mtlsproxy"]

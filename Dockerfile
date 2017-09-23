FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
ADD iugaevtapi iugaevtapi
EXPOSE 4002
ENTRYPOINT ["/iugaevtapi"]
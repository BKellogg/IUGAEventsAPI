FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
ADD iugaevtapi iugaevtapi
EXPOSE 80
EXPOSE 443
ENTRYPOINT ["/iugaevtapi"]
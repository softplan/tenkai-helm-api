FROM ubuntu:18.04
ADD ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
ADD build/tenkai-helm-api /app
CMD ["/app/tenkai-helm-api"]
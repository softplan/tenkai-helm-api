FROM scratch
#ADD ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
#ADD .helm/ /app/.helm/
ADD build/tenkai-helm-api /app
ADD app.yaml /app
CMD ["/app/tenkai-helm-api"]
FROM cgr.dev/chainguard/go AS builder
COPY . /app
RUN cd /app && go build -o reverse-registry .

FROM cgr.dev/chainguard/glibc-dynamic
COPY --from=builder /app/reverse-registry /usr/bin/
COPY ./config/config.local.yaml /etc/reverse-registry/config.local.yaml
CMD ["/usr/bin/reverse-registry", "server", "--config=/etc/reverse-registry/config.local.yaml"]

FROM --platform=$BUILDPLATFORM golang:1.16-alpine as builder

RUN apk add ca-certificates && \
    apk add tzdata && \
    apk add --update gcc musl-dev && \
    apk add gcc-aarch64-linux-gnu

COPY . $GOPATH/src/github.com/vilisseranen/castellers
WORKDIR $GOPATH/src/github.com/vilisseranen/castellers

ARG TARGETOS
ARG TARGETARCH

RUN if [ "${TARGETARCH}" = "arm64" ]; then CC=aarch64-alpine-linux-musl-gcc; fi && \
    env GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=1 CC=$CC go build -o /go/bin/import

FROM scratch

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/import /app
COPY mail/templates /mail/templates
COPY translations /translations
COPY sql /sql
COPY VERSION /VERSION

VOLUME ["/data", "/var/log", "/etc/castellers"]

ENV APP_DB_NAME /data/castellers.db
ENV APP_LOG_FILE /var/log/castellers.log

EXPOSE 8080

WORKDIR /

ENTRYPOINT ["/app"]

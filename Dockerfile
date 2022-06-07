FROM --platform=$BUILDPLATFORM golang:1.16-alpine as builder

RUN apk add ca-certificates && \
    apk add tzdata

COPY . $GOPATH/src/github.com/vilisseranen/castellers
WORKDIR $GOPATH/src/github.com/vilisseranen/castellers

#RUN go get -d -v -u
#ARG TARGETOS TARGETARCH
#RUN env GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /go/bin/import
RUN go build -o /go/bin/import

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

ENTRYPOINT /app

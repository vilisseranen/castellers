FROM alpine:3.11

VOLUME ["/data", "/var/log", "/etc/castellers"]

ENV APP_DB_NAME /data/castellers.db
ENV APP_LOG_FILE /var/log/castellers.log

ENV GOROOT=/usr/lib/go \
    GOPATH=/gopath \
    GOBIN=/gopath/bin \
    PATH=$PATH:$GOROOT/bin:$GOPATH/bin

EXPOSE 8080

COPY . /gopath/src/castellers
COPY templates /templates
COPY sql /sql
COPY VERSION /VERSION

WORKDIR /gopath/src/castellers

RUN apk add -U git go=~1.13 && \
    apk add --update gcc musl-dev && \
    apk add --no-cache ca-certificates && \
    apk add --no-cache tzdata && \
    go get -u && \
    mv /gopath/bin/castellers /app && \
    apk del git go gcc musl-dev && \
    rm -rf /gopath && \
    rm -rf /var/cache/apk/* && \
    rm -rf /root/.cache

WORKDIR /

ENTRYPOINT /app

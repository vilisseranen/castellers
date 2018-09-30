FROM alpine:3.8

COPY frontend/dist /static
COPY templates /templates

RUN ls /static

VOLUME ["/data", "/var/log", "/etc/castellers"]

ENV APP_DB_NAME /data/castellers.db
ENV APP_LOG_FILE /var/log/castellers.log

ENV GOROOT=/usr/lib/go \
    GOPATH=/gopath \
    GOBIN=/gopath/bin \
    PATH=$PATH:$GOROOT/bin:$GOPATH/bin

EXPOSE 8080

WORKDIR /gopath/src/castellers
COPY . /gopath/src/castellers

RUN apk add -U git go && \
    apk add --update gcc musl-dev && \
    apk add --no-cache ca-certificates && \
    go get -v castellers && \
    mv /gopath/bin/castellers /app && \
    apk del git go gcc musl-dev && \
    rm -rf /gopath && \
    rm -rf /var/cache/apk/* && \
    rm -rf /root/.cache

WORKDIR /

ENTRYPOINT /app

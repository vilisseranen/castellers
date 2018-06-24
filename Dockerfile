#FROM alpine:3.7
FROM ubuntu:16.04

COPY castellers /app
COPY frontend/dist /static

VOLUME ["/data/", "/var/log"]

ENV APP_DB_NAME castellers.db
ENV APP_LOG_FILE castellers.log

EXPOSE 8080

ENTRYPOINT /app

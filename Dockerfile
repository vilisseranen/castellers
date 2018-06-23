#FROM alpine:3.7
FROM ubuntu:16.04

MAINTAINER Clément Contini <vilisseranen@gmail.com>

COPY castellers /app
COPY frontend/dist /static

EXPOSE 8080

ENTRYPOINT /app

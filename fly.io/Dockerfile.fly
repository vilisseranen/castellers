FROM vilisseranen/castellers:dev as castellers

FROM alpine as litestream

ADD https://github.com/benbjohnson/litestream/releases/download/v0.3.8/litestream-v0.3.8-linux-amd64-static.tar.gz /tmp/litestream.tar.gz
RUN tar -C /usr/local/bin -xzf /tmp/litestream.tar.gz

FROM alpine

COPY --from=castellers /usr/local/go/lib/time/ /usr/local/go/lib/time/
COPY --from=castellers /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=castellers /app /app
COPY --from=castellers /mail/templates /mail/templates
COPY --from=castellers /translations /translations
COPY --from=castellers /sql /sql
COPY --from=castellers /VERSION /VERSION
COPY --from=litestream /usr/local/bin/litestream /usr/local/bin/litestream

RUN apk add bash

EXPOSE 8080

# Copy Litestream configuration file & startup script.
COPY fly.io/etc/litestream.yml /etc/litestream.yml
COPY fly.io/scripts/run.sh /scripts/run.sh

RUN chmod +x /scripts/run.sh

CMD [ "/scripts/run.sh" ]
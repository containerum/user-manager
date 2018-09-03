FROM golang:1.10-alpine as builder
RUN apk add --update make git
WORKDIR src/git.containerum.net/ch/user-manager
COPY . .
RUN VERSION=$(git describe --abbrev=0 --tags) make build-for-docker

FROM alpine:3.7 as alpine
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
# -0 means no compression.  Needed because go's
# tz loader doesn't handle compressed data.
RUN zip -r -0 /zoneinfo.zip .

FROM alpine:3.7
# app
COPY --from=builder /tmp/user-manager /
# migrations
COPY pkg/migrations /migrations
# timezone data
ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /
# tls certificates
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV GIN_MODE=debug \
    CH_USER_DB="postgres" \
    CH_USER_PG_LOGIN="usermanager" \
    CH_USER_PG_PASSWORD="ae9Oodai3aid" \
    CH_USER_PG_ADDR="postgres:5432" \
    CH_USER_PG_DBNAME="usermanager" \
    CH_USER_PG_NOSSL=true \
    CH_USER_MIGRATIONS_PATH="migrations" \
    CH_USER_MAIL="http" \
    CH_USER_MAIL_URL="http://ch-mail-templater:7070/" \
    CH_USER_RECAPTCHA="dummy" \
    CH_USER_RECAPTCHA_KEY="recaptcha_key" \
    CH_USER_OAUTH_CLIENTS="http" \
    CH_USER_AUTH_HTTP_ADDR="http://ch-auth:1111/" \
    CH_USER_PERMISSIONS="http" \
    CH_USER_PERMISSIONS_URL="http://permissions:4242" \
    CH_USER_TELEGRAM=false \
    CH_USER_TELEGRAM_BOT_ID="" \
    CH_USER_TELEGRAM_BOT_TOKEN="" \
    CH_USER_TELEGRAM_BOT_CHAT_ID="" \
    CH_USER_LISTEN_ADDR=":8111" \
    CH_USER_USER_MANAGER="impl"

ENTRYPOINT ["/user-manager"]

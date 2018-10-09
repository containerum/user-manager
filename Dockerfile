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
    DB="postgres" \
    PG_LOGIN="usermanager" \
    PG_PASSWORD="ae9Oodai3aid" \
    PG_ADDR="postgres:5432" \
    PG_DBNAME="usermanager" \
    PG_NOSSL=true \
    MIGRATIONS_PATH="migrations" \
    MAIL="http" \
    MAIL_URL="http://ch-mail-templater:7070/" \
    RECAPTCHA="dummy" \
    RECAPTCHA_KEY="recaptcha_key" \
    OAUTH_CLIENTS="http" \
    AUTH_URL="http://ch-auth:1111/" \
    PERMISSIONS="http" \
    PERMISSIONS_URL="http://permissions:4242" \
    TELEGRAM=false \
    TELEGRAM_BOT_ID="" \
    TELEGRAM_BOT_TOKEN="" \
    TELEGRAM_BOT_CHAT_ID="" \
    PORT=":8111" \
    USER_MANAGER="impl" \
    EVENTS="http" \
    EVENTS_URL="http://events-api:1667"

ENTRYPOINT ["/user-manager"]

FROM golang:1.9-alpine as builder
WORKDIR src/git.containerum.net/ch/user-manager
COPY . .
RUN CGO_ENABLED=0 go build -v -tags "jsoniter" -ldflags="-w -s -extldflags '-static'" -o /bin/user-manager

FROM alpine:latest as alpine
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
# -0 means no compression.  Needed because go's
# tz loader doesn't handle compressed data.
RUN zip -r -0 /zoneinfo.zip .

FROM alpine:3.7
# app
COPY --from=builder /bin/user-manager /
# migrations
COPY migrations /migrations
# timezone data
ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /
# tls certificates
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV GIN_MODE=release \
    CH_USER_LOG_LEVEL=5 \
    CH_USER_DB="postgres" \
    CH_USER_PG_URL="postgres://usermanager:ae9Oodai3aid@192.168.88.200:5432/usermanager?sslmode=disable" \
    CH_USER_MAIL="http" \
    CH_USER_MAIL_URL="http://ch-mail-templater:7070/" \
    CH_USER_RECAPTCHA="dummy" \
    CH_USER_RECAPTCHA_KEY="recaptcha_key" \
    CH_USER_OAUTH_CLIENTS="http" \
    CH_USER_AUTH_GRPC_ADDR="ch-auth:1112" \
    CH_USER_WEB_API="http" \
    CH_USER_WEB_API_URL="http://web-api:5000" \
    CH_USER_RESOURCE_SERVICE="http" \
    CH_USER_RESOURCE_SERVICE_URL="http://resource-service:1213" \
    CH_USER_LISTEN_ADDR=":8111" \
    CH_USER_USER_MANAGER="impl"
ENTRYPOINT ["/user-manager"]

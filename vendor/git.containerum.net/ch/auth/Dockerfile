FROM golang:1.10-alpine as builder
RUN apk add --update make git
WORKDIR src/git.containerum.net/ch/auth
COPY . .
RUN VERSION=$(git describe --abbrev=0 --tags) make build-for-docker

FROM alpine:3.7

VOLUME ["/keys", "/storage"]

COPY --from=builder /tmp/auth /
ENV CH_AUTH_HTTP_LISTENADDR=0.0.0.0:1111 \
    CH_AUTH_GRPC_LISTENADDR=0.0.0.0:1112 \
    CH_AUTH_LOG_MODE=text \
    CH_AUTH_LOG_LEVEL=4 \
    CH_AUTH_TOKENS=jwt \
    CH_AUTH_JWT_SIGNING_METHOD=HS256 \
    CH_AUTH_ISSUER=containerum.com \
    CH_AUTH_ACCESS_TOKEN_LIFETIME=15m \
    CH_AUTH_REFRESH_TOKEN_LIFETIME=48h \
    CH_AUTH_JWT_SIGNING_KEY_FILE=/keys/jwt.key \
    CH_AUTH_JWT_VALIDATION_KEY_FILE=/keys/jwt.key \
    CH_AUTH_STORAGE=buntdb \
    CH_AUTH_BUNT_STORAGE_FILE=/storage/storage.db \
    CH_AUTH_TRACER=zipkin \
    CH_AUTH_ZIPKIN_COLLECTOR=nop

EXPOSE 1111 1112

CMD ["/auth"]

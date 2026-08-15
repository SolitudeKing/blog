FROM golang:1.26-alpine AS build

WORKDIR /app
# CGO_ENABLED=0 产出静态二进制，避免对 alpine musl libc 的隐式依赖。
# -trimpath 去掉绝对路径信息，-buildid= 让二进制可复现构建。
ENV CGO_ENABLED=0 GOOS=linux GOFLAGS=-trimpath
COPY go.mod go.sum ./
RUN go mod download
COPY ./cmd ./cmd
COPY ./internal ./internal
RUN go build -trimpath -ldflags="-s -w -buildid=" -o /out/api ./cmd/api

FROM alpine:3.21
# alpine 最小镜像不带 CA 信任库与时区数据：MySQL/Redis TLS、S3 HTTPS 与 time.LoadLocation 都依赖这些包。
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app -u 10001 app \
    && mkdir -p /app/storage/uploads && chown -R app:app /app
WORKDIR /app
COPY --from=build --chown=app:app /out/api /app/api
USER app
EXPOSE 8080
CMD ["/app/api"]

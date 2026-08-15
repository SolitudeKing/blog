FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/api ./cmd/api

FROM alpine:3.21
# alpine 最小镜像不带 CA 信任库与时区数据：MySQL/Redis TLS、S3 HTTPS 与 time.LoadLocation 都依赖这些包。
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/api /app/api
EXPOSE 8080
CMD ["/app/api"]

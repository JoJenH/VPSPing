FROM golang:1.21-alpine AS builder

ARG GOPROXY=https://goproxy.cn,direct
ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine

RUN sed -i "s#dl-cdn.alpinelinux.org#${ALPINE_MIRROR}#g" /etc/apk/repositories

WORKDIR /build

ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o vpsping ./cmd/vpsping

FROM alpine:latest

ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine

RUN sed -i "s#dl-cdn.alpinelinux.org#${ALPINE_MIRROR}#g" /etc/apk/repositories && \
    apk --no-cache add ca-certificates sqlite

WORKDIR /app

COPY --from=builder /build/vpsping .
COPY --from=builder /build/config ./config

RUN mkdir -p data logs output

VOLUME ["/app/data", "/app/logs", "/app/output"]

CMD ["./vpsping", "run"]

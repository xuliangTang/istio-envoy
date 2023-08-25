FROM golang:1.19-alpine as builder
RUN mkdir /src
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
ADD . /src
WORKDIR /src
RUN GOPROXY=https://goproxy.cn go build -o envoy-controller mygateway/main.go  && chmod +x envoy-controller


FROM alpine:3.15
RUN mkdir /app
WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add tzdata
ENV TZ=Asia/Shanghai
ENV ZONEINFO=/app/zoneinfo.zip

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /app

COPY --from=builder /src/envoy-controller /app
COPY --from=builder /src/cue.mod /app/cue.mod
COPY --from=builder /src/mygateway/tpls/xds.cue /app/mygateway/tpls/xds.cue
COPY --from=builder /src/mygateway/tpls/common /app/mygateway/tpls/common
COPY --from=builder /src/mygateway/tpls/filters /app/mygateway/tpls/filters

ENTRYPOINT  ["./envoy-controller"]
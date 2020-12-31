FROM golang:1.12-alpine as builder

WORKDIR /build

COPY . .

RUN GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOPROXY=https://goproxy.cn  go build -o app .

FROM uhub.service.ucloud.cn/infra/base-centos7:v1

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /

COPY --from=builder /build/app .

COPY ./configs/ ./configs/

ENTRYPOINT ["/app"]

CMD [ "-conf","/configs/pro.toml" ]
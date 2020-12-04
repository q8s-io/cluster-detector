# build builder
FROM golang:1.12-alpine as builder

ARG CODEPATH

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
	&& apk update \
	&& apk add git \
	&& rm -rf /var/cache/apk/*

WORKDIR $GOPATH/src/$CODEPATH


COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn go build -o /kube-watcher detector.go

# build server
FROM alpine:3.8

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /

COPY --from=builder /kube-watcher .
COPY ./configs/ ./configs/
RUN chmod +x /kube-watcher

ENTRYPOINT ["/kube-watcher"]
CMD ["-conf", "/configs/pro.toml"]


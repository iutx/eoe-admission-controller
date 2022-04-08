FROM golang:1.16-alpine3.14 as build
LABEL maintainer="iutx<root@viper.run>"
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY pkg/ pkg/
COPY cmd/ cmd/

RUN cd cmd/eoe \
    && go build -o /build/dist/eoe .


FROM alpine:3.12
LABEL maintainer="iutx<root@viper.run>"
ENV LANG=en_US.UTF-8 \
    TZ="Asia/Shanghai"
COPY --from=build /build/dist/eoe /opt

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.4/main/" > /etc/apk/repositories \
    && apk add --no-cache -U bash

WORKDIR /opt
EXPOSE 443

CMD ["./eoe"]
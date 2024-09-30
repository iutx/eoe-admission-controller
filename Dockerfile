FROM registry.erda.cloud/erda-x/golang:1.22 AS build
LABEL maintainer="iutx<root@viper.run>"
ENV CGO_ENABLED=0 \
    GOPROXY=https://goproxy.cn

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY pkg/ pkg/
COPY cmd/ cmd/

RUN cd cmd/eoe \
    && go build -o /build/dist/eoe .


FROM registry.erda.cloud/erda-x/debian-bookworm:12
LABEL maintainer="iutx<root@viper.run>"
ENV LANG=en_US.UTF-8 \
    TZ="Asia/Shanghai"

COPY --from=build /build/dist/eoe /opt

WORKDIR /opt
EXPOSE 443

CMD ["./eoe"]
FROM golang:1.20-bullseye as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
ENV CGO_ENABLED=1 GO111MODULE=on
RUN go mod download
RUN go test ./...

RUN cd cmd && \
    GIT_COMMIT=$(git rev-list -1 HEAD) && \
    BUILD_TIME=$(date "+%Y-%m-%d_%H:%M:%S") && \
    go build -ldflags "-X main.GitCommit=$GIT_COMMIT -X main.BuildTime=$BUILD_TIME" \
    -o /build/restcontent .

FROM ubuntu:22.04
LABEL maintainer="shenjinti@fourz.cn"
LABEL org.opencontainers.image.source=https://github.com/restsend/restcontent
RUN apt-get update && apt-get install -y ca-certificates tzdata
ENV DEBIAN_FRONTEND noninteractive
ENV LANG C.UTF-8

WORKDIR /app
COPY --from=builder /build/restcontent /app/
ADD entrypoint.sh /app/
ADD templates /app/templates
ADD static /app/static

EXPOSE 8000
CMD ["/app/entrypoint.sh"]
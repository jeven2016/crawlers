#ARG DOCKER_REGISTRY

FROM docker.io/library/golang:1.19.13 AS builder
#ARG GOPROXY

WORKDIR /workspace
COPY . /workspace
COPY docker/task /workspace
RUN cd /workspace && \
        export GOPROXY=https://goproxy.io && \
        ./task build

RUN ls /workspace/bin/

FROM alpine
COPY --from=builder /workspace/bin/crawlers /usr/bin/crawlers
ENV TZ Asia/Shanghai
ENTRYPOINT ["crawlers"]

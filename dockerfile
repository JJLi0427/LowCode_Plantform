FROM alpine:latest AS builder

ARG TARGETARCH

RUN apk update && apk add --no-cache \
    build-base \
    pkgconfig \
    opencv-dev \
    mariadb-dev \
    curl-dev \
    build

WORKDIR /LowCode_Builder
COPY ${TARGETARCH}/ .

RUN ls apps
RUN chmod +x ./apps/build_apps.sh && \
    ./apps/build_apps.sh

FROM alpine:latest

RUN apk update
RUN apk add --no-cache libstdc++
RUN apk add --no-cache libcurl
RUN apk add --no-cache mariadb-connector-c
RUN apk add --no-cache opencv
RUN apk add --no-cache bash

WORKDIR /LowCode_Docker
COPY --from=builder /LowCode_Builder/ .

EXPOSE 8088

CMD ["./run"]
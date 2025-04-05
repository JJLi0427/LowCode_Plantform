FROM alpine:latest AS builder

ARG TARGETARCH

RUN apk update && apk add --no-cache \
    build-base \
    pkgconfig \
    opencv-dev \
    mariadb-dev \
    curl-dev \
    bash

WORKDIR /LowCode_Builder
COPY ${TARGETARCH}/ .

RUN chmod +x ./apps/build_apps.sh && \
    ./apps/build_apps.sh



FROM alpine:latest

RUN apk update && apk add --no-cache libstdc++ \
    mariadb-connector-c \
    opencv \
    libcurl \
    bash

WORKDIR /LowCode_Docker
COPY --from=builder /LowCode_Builder/ .

RUN chmod +x ./run

EXPOSE 8088

CMD ["./run"]
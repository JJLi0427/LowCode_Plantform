FROM golang:1.20 AS builder

WORKDIR /LowCode_Docker

RUN apt-get update && apt-get install -y \
    g++ \
    build-essential \
    pkg-config \
    libopencv-dev \
    default-libmysqlclient-dev \
    curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN chmod +x build.sh && ./build.sh

FROM alpine:latest

RUN apk add --no-cache bash

WORKDIR /LowCode_Docker

COPY --from=builder /LowCode_Docker/build ./

EXPOSE 8088

CMD ["./run"]
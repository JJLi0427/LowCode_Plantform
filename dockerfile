FROM alpine:latest

RUN apk add --no-cache bash

WORKDIR /LowCode_Docker

COPY ./${TARGETARCH}/* .

EXPOSE 8088

CMD ["./run"]
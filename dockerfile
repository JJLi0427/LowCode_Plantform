FROM alpine:latest

RUN apk add --no-cache bash

WORKDIR /LowCode_Docker

# Copy pre-built artifacts from the build directory
COPY ./ ./

EXPOSE 8088

CMD ["./run"]
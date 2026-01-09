FROM alpine:latest
RUN apk add --no-cache python3 time
WORKDIR /code
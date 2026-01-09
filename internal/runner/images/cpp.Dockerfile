FROM alpine:latest

RUN apk add --no-cache g++ time

WORKDIR /code


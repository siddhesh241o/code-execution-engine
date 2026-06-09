FROM golang:1.25-alpine
RUN apk add --no-cache time
WORKDIR /code

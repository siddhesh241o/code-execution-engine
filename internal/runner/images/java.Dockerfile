FROM eclipse-temurin:17-alpine
RUN apk add --no-cache time
WORKDIR /code
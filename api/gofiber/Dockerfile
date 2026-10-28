### Build app in docker
# Match go.mod (go 1.25.7+). Plain golang:alpine can resolve to an older major (e.g. 1.24).
FROM golang:1.26-alpine AS builder
ENV GOTOOLCHAIN=auto

WORKDIR /src
COPY . /src

RUN go mod download && go mod verify
###RUN apk update && apk add librdkafka-dev pkgconf
RUN go build -o /out/app .

### Build image
FROM alpine:3

RUN apk update && apk add --no-cache --update bash openssl ca-certificates
#RUN apk update && apk add --no-cache --update bash openssl ca-certificates librdkafka-dev pkgconf

WORKDIR /app

COPY --from=builder /out/app .
COPY --from=builder /src/config/json config/json
COPY --from=builder /src/config/build.env .env
ENV DEPLOY_ENV=production

### Unused view at this moment
COPY --from=builder /src/view/ ./views

EXPOSE 3000
ENTRYPOINT ["/app/app"]
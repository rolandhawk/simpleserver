FROM golang:1.12.5-alpine3.9 AS builder

WORKDIR /app
RUN apk add --no-cache git
COPY . /app

RUN go build -o app

FROM alpine:3.9.4

RUN apk add --no-cache bash
COPY --from=builder /app/app .

ENTRYPOINT ./app

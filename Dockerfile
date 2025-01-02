FROM golang:1.23.0-alpine3.20 AS builder
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY main.go .
RUN go build main.go && \
    mv main text-qrcode

FROM github.com/soulteary/docker-text-qrcode:lib
RUN apk add bash
COPY --from=builder /app/text-qrcode /usr/local/bin/text-qrcode
CMD ["text-qrcode"]
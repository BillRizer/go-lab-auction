FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY *.go .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/weather-api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/weather-api .

EXPOSE 8080

CMD ["/app/weather-api"]
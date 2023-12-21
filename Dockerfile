FROM golang:1.21.5 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./twitter-moon .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/twitter-moon /app/example.env ./
EXPOSE ${PORT}
CMD "./twitter-moon"
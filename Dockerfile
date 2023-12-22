FROM node:lts-alpine AS node-builder

WORKDIR /app
COPY tailwind.config.js base.css ./
COPY templates/ ./templates/
RUN npm install -D tailwindcss && npx tailwindcss -i ./base.css -o ./style.css

FROM golang:1.21.5 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
COPY --from=node-builder /app/style.css ./assets/css/style.css
RUN CGO_ENABLED=0 GOOS=linux go build -o ./twitter-moon .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/twitter-moon /app/example.env ./
EXPOSE ${PORT}
CMD "./twitter-moon"
# base image
FROM golang:1.24-alpine AS base
WORKDIR /app
RUN apk add --no-cache git

# build stage
FROM base AS builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# dev stage for hot reload
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
CMD ["air"]

# production image
FROM scratch AS prod
WORKDIR /root/
COPY --from=builder /app/main .
COPY config.prod.ini ./config.ini
CMD ["./main"]
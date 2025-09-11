FROM golang:1.23.4-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o arch ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache
WORKDIR /app
COPY --from=builder /src .
EXPOSE 8080
CMD ["./arch"]

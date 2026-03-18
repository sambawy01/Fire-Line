# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /fireline ./cmd/fireline

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /fireline /fireline

EXPOSE 8080
ENTRYPOINT ["/fireline"]

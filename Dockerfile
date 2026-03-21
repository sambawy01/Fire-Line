FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY migrations/ migrations/
RUN CGO_ENABLED=0 GOOS=linux go build -o fireline ./cmd/fireline

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/fireline .
COPY --from=builder /app/migrations/ ./migrations/
EXPOSE 8080
ENV PORT=8080
CMD ["./fireline"]

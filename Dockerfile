# Use the official Golang image as the base
FROM golang:1.25.4-alpine AS builder

# Set environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set working directory inside the container
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy go.mod and go.sum files for dependency installation
COPY . .

# Download dependencies
RUN go build -ldflags="-s -w" -o /app/web-app ./cmd/web

# Final lightweight stage
FROM alpine:3.21 AS final

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D appuser
USER appuser

# Copy the compiled binary from the builder 
COPY --from=builder /app/web-app /bin/web-app

# Expose the application's port
EXPOSE 4000

# Run the application
CMD ["/bin/app/web"]

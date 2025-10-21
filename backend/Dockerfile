# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:3.20

# Add a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory to the user's home
WORKDIR /home/appuser

# Copy the binary from builder stage and set ownership
COPY --chown=appuser:appgroup --from=builder /app/main .

# Create storage directory and set ownership
RUN mkdir -p ./storage && chown -R appuser:appgroup ./storage

# Switch to the non-root user
USER appuser

# Expose port
EXPOSE 8000

# Run the binary
CMD ["./main"]

# Stage 1: Build the Go application
FROM golang:alpine AS builder

# Set the working directory
WORKDIR /app

# Install git and required build tools
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 ensures a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o gotube .

# Stage 2: Create a minimal runtime environment
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Install FFmpeg, which is required for GoTube to merge videos and convert to MP3
RUN apk add --no-cache ffmpeg tzdata

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/gotube .

# Copy the public directory for the web frontend
COPY --from=builder /app/public ./public

# Create downloads directory with proper permissions
RUN mkdir -p downloads && chmod 777 downloads

# Expose port 1004
EXPOSE 1004

# Run the web server by default
CMD ["./gotube", "serve", "--port", "1004"]

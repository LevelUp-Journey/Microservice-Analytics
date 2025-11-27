# Stage 1: Build the application
FROM golang:1.23.3-alpine AS builder

# Install git (required for go install)
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install swag for generating Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy the rest of the source code
COPY . .

# Generate Swagger documentation
RUN swag init

# Build the application for a Linux environment
# CGO_ENABLED=0 is used to build a statically linked binary
# -o main specifies the output file name
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Create the final lightweight image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port the application runs on
EXPOSE 8080

# The command to run the application
CMD ["./main"]

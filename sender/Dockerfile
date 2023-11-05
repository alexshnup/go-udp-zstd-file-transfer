# Use the official GoLang image as the build environment
FROM golang:latest AS builder

# Set the working directory within the container
WORKDIR /app

# Copy the Go project source code into the container
COPY . .

# Build the Go binary
RUN go build -o myapp

# # Use a lightweight Alpine image for the runtime environment
# FROM alpine:latest

# # Install any required system dependencies
# # For example, you may need 'ca-certificates' if your application makes https requests.
# RUN apk --no-cache add ca-certificates

# # Set the working directory within the container
# WORKDIR /usr/local/bin/

# # Copy the binary from the builder stage into the current stage
# COPY --from=builder /app/myapp .

# Define the command to run your Go application
CMD ["./myapp"]
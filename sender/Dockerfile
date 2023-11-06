# Builder stage
FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/myapp .
RUN chmod +x /app/myapp
RUN ls /app

# Final stage
FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /app/myapp /usr/local/bin/myapp
CMD ["myapp"]

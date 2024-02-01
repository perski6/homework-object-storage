# Builder stage
FROM golang:1.21 AS builder
WORKDIR /mnt/homework
COPY . .
RUN go build -o homework-object-storage

# Final stage
FROM alpine
# If you need bash and curl, install them
RUN apk add --no-cache bash curl
# Copy the binary from the builder stage
COPY --from=builder /mnt/homework/homework-object-storage /usr/local/bin/homework-object-storage
ENTRYPOINT ["/usr/local/bin/homework-object-storage"]
# Stage 1: Build the Go binary
FROM golang:1.21 as builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -o /app/build/bin/external-dns-hetzner-webhook ./cmd/webhook

# Stage 2: Create the final image
FROM gcr.io/distroless/static-debian11:nonroot

USER 20000:20000

# Copy the binary from the builder stage
COPY --from=builder /app/build/bin/external-dns-hetzner-webhook /opt/external-dns-hetzner-webhook/app

# Set the entrypoint
ENTRYPOINT ["/opt/external-dns-hetzner-webhook/app"]
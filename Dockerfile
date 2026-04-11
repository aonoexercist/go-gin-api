# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# We use $(date) inside the RUN command to get the timestamp of the build
RUN go build -ldflags "-X main.BuildVersion=${VERSION} \
                       -X main.GitCommit=${COMMIT} \
                       -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
             -o main .

# Run stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
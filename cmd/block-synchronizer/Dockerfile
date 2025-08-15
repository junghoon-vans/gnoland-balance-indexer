FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy entire project
COPY . .

# Build the specific service
WORKDIR /app/cmd/block-synchronizer
RUN go mod download && go build -o /app/block-synchronizer .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/block-synchronizer .

CMD ["./block-synchronizer"]
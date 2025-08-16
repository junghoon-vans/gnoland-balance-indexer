FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy shared module
COPY shared/ ./shared/

# Copy block-synchronizer service
COPY services/block-synchronizer/ ./services/block-synchronizer/

# Build the service
WORKDIR /app/services/block-synchronizer
RUN go mod download && go build -o /app/block-synchronizer .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/block-synchronizer .

CMD ["./block-synchronizer"]
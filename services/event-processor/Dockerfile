FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy shared module
COPY shared/ ./shared/

# Copy event-processor service
COPY services/event-processor/ ./services/event-processor/

# Build the service
WORKDIR /app/services/event-processor
RUN go mod download && go build -o /app/event-processor .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/event-processor .

CMD ["./event-processor"]

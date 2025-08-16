FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy entire project
COPY . .

# Build the specific service
WORKDIR /app/cmd/balance-api
RUN go mod download && go build -o /app/balance-api .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/balance-api .

EXPOSE 8080

CMD ["./balance-api"]

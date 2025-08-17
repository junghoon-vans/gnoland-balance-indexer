# Gno.land Balance Indexer

A high-performance microservices-based indexer for tracking token balances and transfers on the [Gno.land](https://gno.land/) blockchain.

## Features

- **Event-driven Architecture**: Real-time token balance tracking via SQS message queues
- **Horizontal Scaling**: Multiple event processor instances with idempotent message processing
- **High Performance**: Redis-based caching with configurable TTL strategies
- **Data Consistency**: Database constraints and idempotency patterns prevent duplicate processing
- **Enterprise-grade**: Comprehensive error handling, graceful shutdown, and observability

## Architecture

```mermaid
graph TB
    subgraph "Services"
        BS[Block Synchronizer]
        EP[Event Processor × N]
        API[Balance API]
    end

    GQL[GraphQL API]
    PG[(PostgreSQL)]
    SQS[SQS Queue]
    REDIS[(Redis)]

    GQL --> BS
    BS --> |Blocks & Transactions|PG
    BS --> |Token Events| SQS
    SQS --> EP
    EP --> |Balance Updates|PG
    API --> PG
    API --> REDIS
```

### Services

| Service | Purpose | Scaling |
|---------|---------|---------|
| **Block Synchronizer** | Fetches blocks from GraphQL API and publishes token events | Single instance |
| **Event Processor** | Consumes events from SQS and updates token balances | Horizontal (2+ instances) |
| **Balance API** | REST API for querying balances with Redis caching | Horizontal scaling ready |

## Quick Start

```bash
# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8080/health

# Get token balances
curl "http://localhost:8080/tokens/balances?address=g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
```

## API Reference

### Token Balances
```bash
# Get all token balances (aggregated)
GET /tokens/balances

# Get balances for specific address
GET /tokens/balances?address={address}

# Get balances for specific token
GET /tokens/{token_path}/balances

# Get balance for specific token and address
GET /tokens/{token_path}/balances?address={address}
```

### Transfer History
```bash
# Get all transfers
GET /tokens/transfer-history

# Get transfers for specific address
GET /tokens/transfer-history?address={address}&limit={limit}
```

## Architecture Highlights

### Idempotent Processing
- **Event Deduplication**: `processed_events` table tracks handled events
- **Database Constraints**: `UNIQUE(tx_hash, event_id)` prevents duplicate transfers
- **Graceful Degradation**: System continues operating even if tracking fails

### Scalability & Resilience
- **Message Queues**: SQS ensures reliable event delivery between services
- **Caching Layer**: Redis middleware with configurable TTL reduces database load
- **Connection Pooling**: Optimized database connections via GORM
- **Health Checks**: Docker Compose health monitoring for all services

### Data Consistency
- **ACID Transactions**: Database transactions ensure balance consistency
- **Retry Logic**: Failed messages automatically retry via SQS visibility timeout
- **Sequential Processing**: Block synchronization prevents data gaps

## Development

### Prerequisites
- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 13+
- Redis 6+

### Setup
```bash
# Install dependencies and hooks
make setup-hooks

# Run tests
make test

# Format and lint code
make fmt

# Build all services
make build
```

## Configuration

Environment variables are configured via `.env` files for each service.
See below `.env.example` for reference configuration.
- [balance-api](services/balance-api/config/.env.example)
- [block-synchronizer](services/block-synchronizer/config/.env.example)
- [event-processor](services/event-processor/config/.env.example)

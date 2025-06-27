# DynamoDB Writer

This service reads messages from Redis/Valkey streams and writes them to DynamoDB tables. It works in conjunction with
the kafka-consumer service, which ingests data from Kafka topics into the streams.

## Overview

DynamoDB Writer acts as a stream consumer and data persister, taking messages from Redis/Valkey streams and storing them
in DynamoDB tables. It supports:

- Reading from multiple streams
- Configurable batch sizes and intervals
- Error handling and retries
- Graceful shutdown

## Features

- **AWS SDK for Go v2**: Not the V1 AWS SDK :D
- **Stream Buffer Client**: Uses the stream-buffer library for reading messages from multiple streams
- **Throttleable**: Configurable batch sizes, stream interval, DynamoDB retries and backoff
- **Configurable**: Easy configuration via YAML files and environment variables

## Missing Features

- **Parallel processing**: We should stream in multiple go routines
- **Persist with channels**: stream into channels to be processed in sequence
- **Don't overwrite**: DynamoDB's batch write overwrite preexisting items
- **Metrics**: OpenMetrics/Prometheus
-

## Prerequisites

- Go 1.24.3 or higher
- DynamoDB server
- Redis or Valkey instance (same as used by kafka-consumer)

## Configuration

Configuration is managed through YAML files and environment variables. The default configuration file is located at
`config/base.yaml`.

```yaml
sc:
  log:
    level: "debug"
  dynamoDB:
    tableName: "stream-consumer"
    retryWaitMax: 10s
    retryMax: 5
  stream:
    group: "stream-consumer"
```

You can configure Redis or Valkey connection details through environment variables:

```bash
# For Redis
export SC_STREAM_REDIS_ADDRESS=localhost:6379
export SC_STREAM_REDIS_PASSWORD=your_password

# For Valkey
export SC_STREAM_VALKEY_ADDRESS=localhost:6379
export SC_STREAM_VALKEY_PASSWORD=your_password

# DynamoDB endpoint (defaults to default aws config)
export SC_DYNAMODB_ENDPOINT=localhost:8000
```

## Data Flow

The service processes messages from streams to DynamoDB storage:

```ascii
[Stream Buffer] -> [Message Batching] -> [DynamoDB Table]
      |                    |                     |
      +--------------------+---------------------+
                    Data persistence pipeline
```

Component interactions:

1. Service reads messages from configured streams
2. Messages are batched for efficient writing
3. Batches are written to DynamoDB
4. Messages are acknowledged in the stream
5. Error handling and retries occur at each step
6. Graceful shutdown ensures message processing completion

## Project Structure

```
.
├── cmd/        # Application entry point
├── config/     # Configuration loading and structure definitions
└── internal/   # Core implementation
    ├── dynamo/ # DynamoDB client and operations
    └── consumer/ # Stream reading and message processing
```

## License

See the [LICENSE](../LICENSE.md) file for details.
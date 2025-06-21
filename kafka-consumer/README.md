# Kafka Consumer

A simple Kafka consumer application that polls, validate and processes messages through a stream.

## Overview

Kafka Consumer acts as a scalable data ingestor, taking messages from topics into a Redis or Valkey Stream to be
processed by others. It supports multiple partitions, graceful shutdown, and configurable message validation.

## Features

- **Pure Go Kafka Client**: [franz-go](https://github.com/twmb/franz-go)
- **Configurable**: Easy configuration via YAML files and environment variables
  using [knadh/koanf](https://github.com/knadh/koanf)

## Missing Features

- **DLQ for invalid messages**: Messages from kafka that are invalid accordingly to the given JSON schema are simply
  logged and discarded
- **Metrics**: franz-go haz a plugin for Prometheus

## Things that would be nice but may be out of the scope:

- Use franz-go schema registry client and convenience Serde type for encoding and decoding
    - It would imply dropping the JSON schema validation, as it would be necessary to have all the structures for serde

## Usage Instructions

### Prerequisites

- Go 1.24.3 or higher
- Access to a Kafka cluster
- Redis or Valkey instance
- Schema validation service

### Installation

1. Clone the repository:

```bash
git clone https://github.com/mfelipe/go-feijoada.git
```

### Configuration

Configuration is managed through YAML files and environment variables. The default configuration file is located at
`config/base.yaml`.

```yaml
kc:
  log:
    level: "debug"
  maxProcessRoutines: 50
  partitionRecordsChannelSize: 10
  closeTimeout: 1m
```

You can configure Redis or Valkey connection details through environment variables:

```bash
# For Redis
export KC_REPOSITORY_REDIS_ADDRESS=localhost:6379
export KC_REPOSITORY_REDIS_PASSWORD=your_password

# For Valkey
export KC_REPOSITORY_VALKEY_ADDRESS=localhost:6379
export KC_REPOSITORY_VALKEY_PASSWORD=your_password
```

### Running the server

```bash
docker compose -f docker-compose.yml -p go-feijoada up -d
```

## Data Flow

The consumer processes messages from Kafka through schema validation to stream buffer storage.

```ascii
[Kafka Topics] -> [Consumer Groups] -> [Schema Validation] -> [Stream Buffer]
     |                  |                      |                    |
     +------------------+----------------------+--------------------+
              Data validation and processing pipeline
```

Component interactions:

1. Consumer subscribes to configured Kafka topics
2. Messages are distributed to partition-specific consumers
3. Each message is validated against its schema
4. Valid messages are forwarded to the stream buffer
5. Offsets are committed back to Kafka
6. Error handling and logging occur at each step
7. Graceful shutdown ensures message processing completion

## Integration with Other Services

Kafka Consumer depends on the following components in the go-feijoada project:

- **Schema Validator**: For validating data against stored schemas
- **Schema Repository**: Provides access to the known JSON schemas

## Project Structure

```
.
├── cmd/        # Application entry point with signal handling and consumer lifecycle management
├── config/     # Configuration loading and structure definitions
└── internal/   # Core consumer implementation with partition handling and message validation
```

## License

See the [LICENSE](../LICENSE.md) file for details.
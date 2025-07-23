# Go Feijoada

## Overview

Go Feijoada is a monorepo study project that brings together several Go microservices and utilities for data streaming, schema management, validation, and integration with Kafka, Redis/Valkey, and DynamoDB. Each module is designed to be reusable and composable, demonstrating best practices and modern Go patterns.

## What you can find here:

- **[schema-repository](./schema-repository/README.md)**: RESTful service for storing and retrieving JSON schemas with versioning, backed by Redis or Valkey.
- **[schema-validator](./schema-validator/README.md)**: Module for validating JSON data against cached or remote schemas.
- **[schemas](./schemas/README.md)**: Module for all JSON schema definitions and related code generation utilities.
- **[kafka-consumer](./kafka-consumer/README.md)**: Kafka consumer that ingests messages into Redis/Valkey streams and validates them.
- **[kafka-producer](./kafka-producer/README.md)**: Kafka producer for publishing messages to Kafka topics.
- **[stream-buffer](./stream-buffer/README.md)**: Go client for Redis/Valkey stream operations, used by other services for scalable stream processing.
- **[stream-consumer](./stream-consumer/README.md)**: Service that reads from Redis/Valkey streams and writes to DynamoDB tables.
- **[utils](./utils/README.md)**: Utility module for logging, configuration, HTTP client, and testing helpers.

## Why Monorepo?

All modules are in the same repository to make it easier to experiment, share code, and demonstrate integration. This is a study project, not a production system.

## Getting Started

Clone the repository:
```bash
git clone https://github.com/mfelipe/go-feijoada.git
```

Although each module has its own README with setup instructions, the project is supposed to be run as a whole. You can use the provided Docker Compose file to start all services together:

```bash
docker compose -f docker-compose.yml -p go-feijoada up -d
```

## AI Tools Used

Some code and documentation were generated or assisted by AI tools:
- AWS Q: Evolved a lot in the last 6 months, and is better used from the command line than as an IDE plugin. Still generates a lot of garbage and broken for Go.
- Google Jules: Good for generating boilerplate code, not so good on refactorings or changes on large codebases.
- GitHub Copilot: Haven't used much, but I liked what I saw so far.

## Upcoming features

As the base is set, the next steps will be to add more features and improve the existing ones. The main ideas are:
- Metrics: collect and expose metrics for each service, using Prometheus and Grafana (probably).
- Kubernetes: make the project deployable on Kubernetes for testing scalability.
- Testing: improve unit and integration tests.

Any contributions or suggestions are welcome!

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for details.

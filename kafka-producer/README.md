# Kafka Producer

A simple Kafka producer application for sending messages to Kafka topics. Part of the go-feijoada project.

## Overview

Kafka Producer acts as a scalable data publisher, sending messages to Kafka topics for downstream processing. It supports configurable topics, message formats, and integration with other services in the go-feijoada ecosystem.

## Features

- **Pure Go Kafka Client**: [franz-go](https://github.com/twmb/franz-go)
- **Configurable**: Easy configuration via YAML files and environment variables
- **Random Data**: Generate random data from known structures from the schemas project

## Usage Instructions

### Prerequisites
- Go 1.24.3 or higher
- Access to a Kafka cluster

### Installation

```bash
git clone https://github.com/mfelipe/go-feijoada.git
```

### Configuration

Configuration is managed through YAML files and environment variables. The default configuration file is located at `config/base.yaml`.

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE.md) file for details.


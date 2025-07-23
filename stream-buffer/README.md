# Stream Buffer

A Go client for Redis and Valkey stream operations. Part of the go-feijoada project.

## Overview

Stream Buffer provides high-level APIs for interacting with Redis/Valkey streams, including group reading, message addition, acknowledgment, and deletion. It is used by other services in the go-feijoada ecosystem for scalable stream processing.

## Features
- Group reading from streams
- Add, acknowledge, and delete messages
- Support for Redis and Valkey

## Usage Instructions

### Prerequisites
- Go 1.24 or higher
- Redis or Valkey instance

### Installation
```bash
git clone https://github.com/mfelipe/go-feijoada.git
```

### Example Usage
```go
import (
	"github.com/mfelipe/go-feijoada/stream-buffer"
	cfg "github.com/mfelipe/go-feijoada/stream-buffer/config"
)

// Create a new stream buffer client
stream := streambuffer.New(cfg.Config{})

// Add a message to a stream
err := buffer.Add(ctx, message)

// Read messages from a stream group
messages, err := buffer.ReadGroup(ctx)
```

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE.md) file for details.

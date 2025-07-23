# Schema Validator

A Go module for validating JSON data against pre-compiled JSON schemas. Part of the go-feijoada project.

## Overview
Schema Validator caches compiled JSON schemas for fast validation. If a schema is not cached, it fetches it from the schema-repository service. It exposes validation functionality for use by other services.

## Features
- Pre-compiled schema cache for performance
- Fetches uncached schemas from schema-repository
- Validates JSON data against schemas
- Simple API for integration

## Usage Instructions

### Prerequisites
- Go 1.24 or higher
- Access to schema-repository service

### Installation
```bash
git clone https://github.com/mfelipe/go-feijoada.git
```

### Example Usage
```go
import "github.com/mfelipe/go-feijoada/schema-validator"

// Initialize the schema validator with configuration
validator := schemavalidator.New(schema-validator.Config{})

// Validate data against a schema
err := validator.Validate(schemaID, data)
```

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE.md) file for details.

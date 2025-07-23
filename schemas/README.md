# Schemas

This directory contains all JSON schema definitions and related code generation utilities for the go-feijoada monorepo.

## Overview

- Stores versioned JSON schema files for all supported entities (e.g., product, order, address, payment, etc.)
- Provides automation scripts for generating Go models from schemas
- Centralizes schema management for validation and code generation

## Structure

- `schemas/`: Contains all JSON schema files, named and versioned (e.g., product-1.0.0.json, product-2.0.0.json)
- `models/`: Generated Go models organized by entity and version (e.g., models/product/v1_0_0/product.go)
- `generate-go-jsonschema.sh`: Bash script to automate Go model generation from JSON schemas
- `generate.go`: Go program for advanced schema/model generation
- `Makefile`: Automates schema code generation during build

## Usage

### Generate Go Models

Run the following command to generate Go models from all JSON schemas:

```bash
make build
```

Or run the script directly:

```bash
bash generate-go-jsonschema.sh
```

### Add a New Schema

1. Add your new JSON schema file to the `schemas/` directory, following the naming convention: `<entity>-<version>.json`.
2. Run `make build` to generate the corresponding Go model.

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE.md) file for details.

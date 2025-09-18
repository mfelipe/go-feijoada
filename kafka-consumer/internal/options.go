package internal

import schemavalidator "github.com/mfelipe/go-feijoada/schema-validator"

type Option func(*Consumer)

func WithStream(stream Stream) Option {
	return func(c *Consumer) {
		c.stream = stream
	}
}

func WithValidator(validator schemavalidator.SchemaValidator) Option {
	return func(c *Consumer) {
		c.validator = validator
	}
}

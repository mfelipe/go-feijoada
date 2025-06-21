# Go Feijoada

## What is it?

Go Feijoada is not me asking you to try one of my favorite dishes in the Brazilian cuisine. It's a study over a few
things I've been professionally working with in the last years, all mixed together in a chaotic but planned ~~recipe~~
architecture to showcase what can be done with such ~~ingredients~~ products, services, go modules and so on.

It can be extrapolated into more specific functionalities, but I left it generic on purpose to avoid making it a "
product"
or eventually raising questions on code ownership/intellectual property for past or present work I've done.

## What you can find here:

More details are available in each project READMEs.

- **[schema-repository](./schema-repository)**:
   - Webservice using Gin
   - Custom request binding validation
   - Custom request struct unmarshalling
  - Configurable by yaml file or environment variables
   - Redis or Valkey for persistence
  -
- **[schema-validator](./schema-validator)**:
  - Holds pre-compiled JSON schemas validator cache
  - Fetch uncached schemas from schema-repository
  - Provides JSON schema validation functionality

- **[kafka-consumer](./kafka-consumer)**:
  - Read messages from kafka topics and add them to Redis or Valkey streams
  - Validate the data against JSON schemas using schema-validator

- **[stream-buffer](./stream-buffer)**:
  - Client for reading or writing to Redis or Valkey streams
  -
- **[utils](./utils)**:
  - Utility module for useful common functionalities like logging, config and http client

# Vibe Coding

I've used a few AI tools to help me with this project. They can be extremely time-saving if used responsibly, mainly
(but not exclusively) for boring or repetitive tasks like creating new base projects, tests scenarios, READMEs, etc.

- [AWS Q](https://aws.amazon.com/en/q/)
- [Google Jules](https://jules.google.com)

## Other thoughts:

- The modules are in the same repository for a reason: this is a study, an experience, an example. I didn't want it to
  be scattered because they don't mean much separated.  

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for details.
